package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/runar-rkmedia/go-common/logger"
	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/models"
	"github.com/runar-rkmedia/gotally/types"
	"github.com/xo/dburl"
	"golang.org/x/net/context"
)

type persistantStorage struct {
	db        *sql.DB
	ruleCache ruleCache
}

// NOTE: sqlboiler does not seem to be a perfect match here.
// I'm thinking of switching to https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html

func NewPersistantStorage(l logger.AppLogger, dsn string) (*persistantStorage, error) {
	db, err := newDb(l, dsn, true)
	if err != nil {
		return nil, err
	}
	p := &persistantStorage{
		db,
		newRuleCache(),
	}

	err = p.fetchRules(context.TODO())
	if err != nil {
		return p, err
	}

	return p, nil

}

func (p *persistantStorage) GetUserBySessionID(ctx context.Context, payload types.GetUserPayload) (*types.SessionUser, error) {
	ctx, span := tracerMysql.Start(ctx, "GetUserBySessionID")
	defer span.End()
	s, err := models.SessionByID(ctx, p.db, payload.ID)
	if sqlOk(err) != nil {
		return nil, fmt.Errorf("Failed to find session for user by ID: %w", err)
	}
	if s == nil {
		return nil, nil
	}
	u, err := models.UserByID(ctx, p.db, s.UserID)
	if sqlOk(err) != nil {
		return nil, fmt.Errorf("Failed to find user from session: %w", err)
	}
	if u == nil {
		return nil, fmt.Errorf("sessions user-object was unexpectedly nil")
	}
	if !u.ActiveGameID.Valid {
		us, err := modelToSessionUser(ctx, *u, *s, nil, nil)
		return &us, err
	}
	g, err := models.GameByID(ctx, p.db, u.ActiveGameID.String)
	if sqlOk(err) != nil {
		return nil, fmt.Errorf("Failed to find game for user: %w", err)
	}
	if g == nil {
		us, err := modelToSessionUser(ctx, *u, *s, nil, nil)
		return &us, err
	}
	r, err := models.RuleByID(ctx, p.db, g.RuleID)
	if sqlOk(err) != nil {
		return nil, fmt.Errorf("Failed to find rule for game: %w", err)
	}
	us, err := modelToSessionUser(ctx, *u, *s, g, r)
	if sqlOk(err) != nil {
		return nil, fmt.Errorf("failed to convert model-data to SessionUser: %w", err)
	}
	return &us, err
}
func (p *persistantStorage) ensureRuleExists(ctx context.Context, db models.DB, r models.Rule) (*models.Rule, error) {
	existing := p.ruleCache.getCachedRule(r.Slug)
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
	ctx, span := tracerMysql.Start(ctx, "CreateUserSession")
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
	modelGame, modelRule, err := modelFromGame(ctx, payload.Game)
	if err != nil {
		return nil, err
	}
	modelGame.UserID = user.ID
	// TOOD: only save rule if it does not exist yet.
	if r, err := p.ensureRuleExists(ctx, tx, modelRule); err != nil {
		return nil, err
	} else if r == nil {
		return nil, fmt.Errorf("failed to ensure the rule-set exists, it returned nil")
	} else if r.ID == "" {
		return nil, fmt.Errorf("failed to ensure the rule-set exists, the ID was not set")
	} else {
		modelRule = *r
		modelGame.RuleID = r.ID
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
	session, err := modelToSessionUser(ctx, user, sess, &modelGame, &modelRule)
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
	ctx, span := tracerMysql.Start(ctx, "SwipeBoard")
	defer span.End()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := payload.Validate(); err != nil {
		return err
	}
	h := models.GameHistory{
		CreatedAt: time.Now(),
		GameID:    payload.GameID,
		Move:      uint(payload.Moves),
		Points:    0,
		Kind:      0,
		Data:      nil,
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
	instr := tallyv1.Instruction{
		InstructionOneof: &tallyv1.Instruction_Swipe{
			Swipe: toModalDirection(payload.SwipeDirection),
		},
	}
	historyData, err := MarshalInternalDataHistory(ctx, payload.State, payload.Cells, &instr)
	if err != nil {
		return fmt.Errorf("%s %w: Cells", err, ErrArgumentInvalid)
	}
	h.Data = historyData
	gameData, err := MarshalInternalDataGame(ctx, payload.Seed, payload.State, payload.Cells)
	if err != nil {
		return fmt.Errorf("%s %w: Cells", err, ErrArgumentInvalid)
	}
	g, err := models.GameByID(ctx, tx, payload.GameID)
	if err != nil {
		return fmt.Errorf("failed to find game %w", err)
	}
	g.Moves++
	g.Data = gameData
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
	ctx, span := tracerMysql.Start(ctx, "CombinePath")
	defer span.End()
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := payload.Validate(); err != nil {
		return err
	}
	instr := tallyv1.Instruction{
		InstructionOneof: &tallyv1.Instruction_Combine{
			Combine: &tallyv1.Indexes{
				Index: payload.Path,
			},
		},
	}
	dataHistory, err := MarshalInternalDataHistory(ctx, payload.State, payload.Cells, &instr)
	if err != nil {
		return fmt.Errorf("%s %w: dataHistory", err, ErrArgumentInvalid)
	}
	dataGame, err := MarshalInternalDataGame(ctx, payload.Seed, payload.State, payload.Cells)
	if err != nil {
		return fmt.Errorf("%s %w: dataGame", err, ErrArgumentInvalid)
	}
	h := models.GameHistory{
		CreatedAt: time.Now(),
		GameID:    payload.GameID,
		Move:      uint(payload.Moves),
		Points:    payload.Points,
		Kind:      models.KindCombine,
		Data:      dataHistory,
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
	g.Data = dataGame
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
	ctx, span := tracerMysql.Start(ctx, "fetchRules")
	defer span.End()
	rules, err := models.GetAllRules(ctx, p.db)
	if err != nil {
		return err
	}
	p.ruleCache.addRulesToCache(rules)
	return nil
}
func (p *persistantStorage) NewGameForUser(ctx context.Context, payload types.NewGamePayload) (types.Game, error) {
	ctx, span := tracerMysql.Start(ctx, "NewGameForUser")
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
		activeGame, err := models.GameByID(ctx, tx, u.ActiveGameID.String)
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
	modelGame, modelRules, err := modelFromGame(ctx, payload.Game)
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
	tg, err = modelToGame(ctx, modelGame, modelRules)
	if err != nil {
		return tg, err
	}
	err = tx.Commit()
	if err != nil {
		return tg, err
	}
	return tg, err
}
