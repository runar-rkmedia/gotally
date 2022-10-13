package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
	"github.com/runar-rkmedia/go-common/logger"
	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/models"
	"github.com/runar-rkmedia/gotally/types"
	"github.com/xo/dburl"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"golang.org/x/net/context"
)

// NOTE: sqlboiler does not seem to be a perfect match here.
// I'm thinking of switching to https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html

func NewPersistantStorage(l logger.AppLogger, dsn string) (*persistantStorage, error) {
	if dsn == "" {
		dsn = os.Getenv("DSN")
	}
	if dsn == "" {
		// database for local development, don't worry, this is not the password I use for everything, I swear!
		dsn = "my://root:secret@localhost/tallyboard"
	}

	models.SetErrorLogger(log.Logger)
	models.SetLogger(log.Logger)

	u, err := parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn-string: %w", err)
	}
	if l.HasDebug() {
		l.Debug().
			Str("host", u.Host).
			Str("user", u.User.Username()).
			Str("scheme", u.Scheme).
			Str("driver", u.Driver).
			Str("redacted", u.Redacted()).
			Msg("Attempting to connect to database with these details (some are hidden)")
	}
	// db, err := passfile.OpenURL(u, ".", "db")
	var attr attribute.KeyValue
	switch u.Driver {
	case "mysql":
		attr = semconv.DBSystemMySQL
	default:
		l.Warn().Str("driver", u.Driver).Msg("unmapped driver-attribute for opentelemetry")
	}

	db, err := otelsql.Open(u.Driver, u.DSN, otelsql.WithAttributes(attr), otelsql.WithSpanOptions(otelsql.SpanOptions{
		Ping:           true,
		RowsNext:       false,
		DisableErrSkip: true,
		DisableQuery:   false,
		RecordError: func(err error) bool {
			if err == sql.ErrNoRows {
				return false
			}
			return true
		},
		OmitConnResetSession: false,
		OmitConnPrepare:      false,
		OmitConnQuery:        false,
		OmitRows:             false,
		OmitConnectorConnect: false,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed during passfile.OpenUrl: %w", err)
	}
	if l.HasDebug() {
		l.Debug().
			Msg("Successfully connected to database")
	}

	err = db.Ping()
	if err != nil {
		return nil, err

	}
	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(attr))
	if err != nil {
		return nil, err
	}
	p := &persistantStorage{
		db,
		ruleCache{
			make(map[string]models.Rule),
			sync.RWMutex{},
		},
	}

	err = p.fetchRules(context.TODO())
	if err != nil {
		return p, err
	}

	return p, nil

}

type ruleCache struct {
	rules map[string]models.Rule
	sync.RWMutex
}
type persistantStorage struct {
	db        *sql.DB
	ruleCache ruleCache
}

func createID() string {
	return gonanoid.Must()
}

func sqlOk(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

var (
	tracer = otel.Tracer("database")
)

func (p *persistantStorage) GetUserBySessionID(ctx context.Context, payload types.GetUserPayload) (*types.SessionUser, error) {
	ctx, span := tracer.Start(ctx, "GetUserBySessionID")
	defer span.End()
	s, err := models.SessionByID(ctx, p.db, payload.ID)
	if err != nil {
		return nil, fmt.Errorf("Failed to find session for user by ID: %w", err)
	}
	if s == nil {
		return nil, fmt.Errorf("session was unexpectedly nil")
	}
	u, err := s.User(ctx, p.db)
	if err != nil {
		return nil, fmt.Errorf("Failed to find user from session: %w", err)
	}
	if u == nil {
		return nil, fmt.Errorf("sessions user-object was unexpectedly nil")
	}
	if !u.ActiveGameID.Valid {
		us, err := modelToSessionUser(*u, *s, nil, nil)
		return &us, err
	}
	g, err := u.Game(ctx, p.db)
	if sqlOk(err) != nil {
		return nil, fmt.Errorf("Failed to find game for user: %w", err)
	}
	if g == nil {
		us, err := modelToSessionUser(*u, *s, nil, nil)
		return &us, err
	}
	r, err := g.Rule(ctx, p.db)
	if sqlOk(err) != nil {
		return nil, fmt.Errorf("Failed to find rule for game: %w", err)
	}
	us, err := modelToSessionUser(*u, *s, g, r)
	if err != nil {
		return nil, fmt.Errorf("failed to convert model-data to SessionUser: %w", err)
	}
	return &us, err
}
func (p *persistantStorage) getCachedRule(hash string) *models.Rule {
	p.ruleCache.RLock()
	defer p.ruleCache.RUnlock()
	r := p.ruleCache.rules[hash]
	return &r
}
func (p *persistantStorage) ensureRuleExists(ctx context.Context, db models.DB, r models.Rule) (*models.Rule, error) {
	existing := p.getCachedRule(r.Slug)
	if existing != nil {
		return existing, nil
	}
	created, _, err := r.CreateIfUnique(ctx, db)
	if err != nil {
		return nil, err
	}
	return created, nil
}
func (p *persistantStorage) CreateUserSession(ctx context.Context, payload types.CreateUserSessionPayload) (*types.SessionUser, error) {
	ctx, span := tracer.Start(ctx, "CreateUserSession")
	defer span.End()
	err := payload.Validate()
	if err != nil {
		return nil, err
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	user := models.User{
		ID:        createID(),
		CreatedAt: time.Now(),
		Username:  payload.Username,
	}
	if user.ID == "" {
		user.ID = createID()
	}
	if err := user.Upsert(ctx, tx); err != nil {
		return nil, err
	}
	sess := models.Session{
		CreatedAt:    time.Now(),
		InvalidAfter: payload.InvalidAfter,
		UserID:       user.ID,
	}
	if sess.ID == "" {
		sess.ID = createID()
	}
	if err := sess.Upsert(ctx, tx); err != nil {
		return nil, err
	}
	user.ActiveGameID = sql.NullString{Valid: true, String: payload.Game.ID}
	modelGame, modelRule, err := modelFromGame(payload.Game)
	if err != nil {
		return nil, err
	}
	modelGame.UserID = user.ID
	// TOOD: only save rule if it does not exist yet.
	if r, err := p.ensureRuleExists(ctx, tx, modelRule); err != nil {
		return nil, err
	} else {
		modelRule = *r
	}
	if err := modelGame.Insert(ctx, tx); err != nil {
		return nil, err
	}
	user.UpdatedAt = toNullTimeNonNullable(time.Now())
	err = user.Update(ctx, tx)
	if err != nil {
		return nil, err

	}
	modelGame.UserID = user.ID
	modelGame.RuleID = modelRule.ID
	if err := modelGame.Update(ctx, tx); err != nil {
		return nil, err
	}
	session, err := modelToSessionUser(user, sess, &modelGame, &modelRule)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	return &session, err
}

func parse(dsn string) (*dburl.URL, error) {
	v, err := dburl.Parse(dsn)
	if err != nil {
		return nil, err
	}
	switch v.Driver {
	case "mysql":
		q := v.Query()
		q.Set("parseTime", "true")
		v.RawQuery = q.Encode()
		return dburl.Parse(v.String())
	case "sqlite3":
		q := v.Query()
		q.Set("_loc", "auto")
		v.RawQuery = q.Encode()
		return dburl.Parse(v.String())
	}
	return v, nil
}
func (p *persistantStorage) SwipeBoard(ctx context.Context, payload types.SwipePayload) error {
	ctx, span := tracer.Start(ctx, "SwipeBoard")
	defer span.End()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if payload.GameID == "" {
		return fmt.Errorf("%w: GameId", ErrArgumentRequired)
	}
	if payload.Moves <= 0 {
		return fmt.Errorf("%w: Moves", ErrArgumentRequired)
	}
	if payload.State <= 0 {
		return fmt.Errorf("%w: Moves", ErrArgumentRequired)
	}
	if payload.SwipeDirection == "" {
		return fmt.Errorf("%w: SwipeDirection", ErrArgumentRequired)
	}
	if len(payload.Cells) == 0 {
		return fmt.Errorf("%w: Cells", ErrArgumentRequired)
	}
	h := models.GameHistory{
		CreatedAt: time.Now(),
		GameID:    payload.GameID,
		Move:      uint(payload.Moves),
		// Kind:      models.kind,
		State:  payload.State,
		Points: 0,
		Data:   nil,
	}
	switch payload.SwipeDirection {
	case types.SwipeDirectionUp:
		h.Kind = models.KindSwipeUp
	case types.SwipeDirectionRight:
		h.Kind = models.KindSwipeRight
	case types.SwipeDirectionDown:
		h.Kind = models.KindSwipeDown
	case types.SwipeDirectionLeft:
		h.Kind = models.KindSwipeLeft
	default:
		return fmt.Errorf("%w: SwipeDirection (%s)", ErrArgumentInvalid, payload.SwipeDirection)
	}
	cells, err := MarshalCellValues(payload.Cells)
	if err != nil {
		return fmt.Errorf("%s %w: Cells", err, ErrArgumentInvalid)
	}
	g, err := models.GameByID(ctx, tx, payload.GameID)
	if err != nil {
		return fmt.Errorf("failed to find game %w", err)
	}
	g.Moves++
	g.Cells = cells
	g.State = payload.State
	g.UpdatedAt = toNullTimeNonNullable(time.Now())
	err = g.Update(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to update the game")
	}

	err = h.Insert(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to save game-history: %w", err)
	}
	err = tx.Commit()
	return err
}
func (p *persistantStorage) CombinePath(ctx context.Context, payload types.CombinePathPayload) error {
	ctx, span := tracer.Start(ctx, "CombinePath")
	defer span.End()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if payload.GameID == "" {
		return fmt.Errorf("%w: GameId", ErrArgumentRequired)
	}
	if payload.Moves <= 0 {
		return fmt.Errorf("%w: Moves", ErrArgumentRequired)
	}
	if payload.Points == 0 {
		return fmt.Errorf("%w: Points", ErrArgumentRequired)
	}
	if payload.Score == 0 {
		return fmt.Errorf("%w: Score", ErrArgumentRequired)
	}
	if payload.State == 0 {
		return fmt.Errorf("%w: State", ErrArgumentRequired)
	}
	if len(payload.Cells) == 0 {
		return fmt.Errorf("%w: Cells", ErrArgumentRequired)
	}
	h := models.GameHistory{
		CreatedAt: time.Now(),
		GameID:    payload.GameID,
		Move:      uint(payload.Moves),
		Kind:      models.KindCombine,
		State:     payload.State,
		Points:    payload.Points,
		Data:      nil,
	}
	cells, err := MarshalCellValues(payload.Cells)
	if err != nil {
		return fmt.Errorf("%s %w: Cells", err, ErrArgumentInvalid)
	}
	g, err := models.GameByID(ctx, tx, payload.GameID)
	if err != nil {
		return fmt.Errorf("failed to find game %w", err)
	}
	g.Moves++
	if g.Score+uint64(payload.Points) != payload.Score {
		return fmt.Errorf("mismatch between score and points")
	}
	g.Score = uint64(payload.Points)
	g.Cells = cells
	g.State = payload.State
	g.UpdatedAt = toNullTimeNonNullable(time.Now())
	err = g.Update(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to update the game")
	}

	err = h.Insert(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to save game-history: %w", err)
	}
	err = tx.Commit()
	return err
}

func (p *persistantStorage) fetchRules(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "fetchRules")
	defer span.End()
	rules, err := models.GetAllRules(ctx, p.db)
	if err != nil {
		return err
	}
	p.ruleCache.Lock()
	defer p.ruleCache.Unlock()
	for i := 0; i < len(rules); i++ {
		p.ruleCache.rules[rules[i].Slug] = rules[i]
	}
	return nil
}
func (p *persistantStorage) NewGameForUser(ctx context.Context, payload types.NewGamePayload) (types.Game, error) {
	ctx, span := tracer.Start(ctx, "NewGameForUser")
	defer span.End()
	tg := types.Game{}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return tg, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if payload.Game.ID == "" {
		return tg, fmt.Errorf("%w: Game.Id", ErrArgumentRequired)
	}
	if payload.Game.UserID == "" {
		return tg, fmt.Errorf("%w: UserID", ErrArgumentRequired)
	}
	if len(payload.Game.Cells) == 0 {
		return tg, fmt.Errorf("%w: Game.Cells", ErrArgumentRequired)
	}
	u, err := models.UserByID(ctx, tx, payload.Game.UserID)
	if err != nil {
		return tg, fmt.Errorf("failed to to retrieve user %w", err)
	}

	if u.ActiveGameID.Valid {
		activeGame, err := u.Game(ctx, tx)
		if err != nil {
			return tg, fmt.Errorf("failed to to retrieve activegame for user %w", err)
		}
		if activeGame.PlayState == models.PlayStateCurrent {
			activeGame.PlayState = models.PlayStateAbandoned
			activeGame.UpdatedAt = toNullTimeNonNullable(time.Now())
			err = activeGame.Update(ctx, tx)
			if err != nil {
				return tg, fmt.Errorf("failed to to update activegame for user %w", err)
			}
		}
	}
	modelGame, modelRules, err := modelFromGame(payload.Game)
	if err != nil {
		return tg, err
	}
	r, err := p.ensureRuleExists(ctx, tx, modelRules)
	if err != nil {
		return tg, fmt.Errorf("failed to save the rules for the game: %w", err)
	} else {
		modelRules = *r
		modelGame.RuleID = r.ID
	}
	err = modelGame.Insert(ctx, tx)
	if err != nil {
		return tg, fmt.Errorf("failed to save the game for: %w", err)
	}
	u.ActiveGameID = sql.NullString{Valid: true, String: modelGame.ID}

	u.UpdatedAt = toNullTimeNonNullable(time.Now())
	err = u.Update(ctx, tx)
	if err != nil {
		return tg, fmt.Errorf("failed to update userut: %w", err)
	}
	tg, err = modelToGame(modelGame, modelRules)
	if err != nil {
		return tg, err
	}
	err = tx.Commit()
	if err != nil {
		return tg, err
	}
	return tg, err
}

var (
	ErrArgumentRequired = errors.New("argument required")
	ErrArgumentInvalid  = errors.New("argument invalid")
)
