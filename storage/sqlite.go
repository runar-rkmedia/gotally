package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/runar-rkmedia/go-common/logger"
	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/sqlite"
	"github.com/runar-rkmedia/gotally/types"
)

type sqliteStorage struct {
	db *sql.DB
	// rules are accessed very often, and there are not too many in any case
	// therefore we cache them
	ruleCache ruleCacheSqLite
	queries   sqlite.Queries
}

func NewSqliteStorage(l logger.AppLogger, dsn string) (*sqliteStorage, error) {
	db, err := newDb(l, dsn, true)
	if err != nil {
		return nil, err
	}
	p := &sqliteStorage{
		db,
		newRuleCacheSqlite(),
		*sqlite.New(db),
	}
	_, err = p.queries.InitializeDatabase(context.TODO())
	if err != nil {
		return nil, err
	}

	err = p.fetchRules(context.TODO())
	if err != nil {
		return p, err
	}
	if l.HasDebug() {
		l.Debug().
			Int("ruleCount", len(p.ruleCache.rules)).
			Msg("prefetched rules")
	}
	return p, nil
}
func (p *sqliteStorage) fetchRules(ctx context.Context) (err error) {
	ctx, span := tracerSqlite.Start(ctx, "fetchRules")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	rules, err := p.queries.GetAllRules(ctx)
	if err != nil {
		return err
	}
	p.ruleCache.addRulesToCache(rules)
	return nil
}

func (p *sqliteStorage) beginTx(ctx context.Context) (*sqlite.Queries, *sql.Tx, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	return p.queries.WithTx(tx), tx, nil
}

// Creates a user, session, game and makes sure the rule exists.
// This should only be used for new users, not to log in existing users.
func (p *sqliteStorage) CreateUserSession(ctx context.Context, payload types.CreateUserSessionPayload) (sess *types.SessionUser, err error) {
	ctx, span := tracerSqlite.Start(ctx, "fetchRules")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	err = payload.Validate()
	if err != nil {
		return nil, fmt.Errorf("payload-validation-failed: %w", err)
	}
	q, tx, err := p.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	userArgs := sqlite.InsertUserParams{
		ID:           payload.UserID,
		CreatedAt:    time.Now(),
		Username:     payload.Username,
		ActiveGameID: payload.Game.ID,
	}

	createdUser, err := q.InsertUser(ctx, userArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}
	sessionUser := types.SessionUser{
		User: types.User{
			ID:        createdUser.ID,
			CreatedAt: createdUser.CreatedAt,
			UpdatedAt: &createdUser.UpdatedAt.Time,
			UserName:  createdUser.Username,
		},
	}
	sessionArgs := sqlite.InsertSessionParams{
		ID:           payload.SessionID,
		CreatedAt:    time.Now(),
		InvalidAfter: payload.InvalidAfter,
		UserID:       userArgs.ID,
	}
	if sessionArgs.ID == "" {
		sessionArgs.ID = createID()
	}
	createdSession, err := q.InsertSession(ctx, sessionArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to insert session %w", err)
	}
	sessionUser.Session = types.Session{
		ID:           createdSession.ID,
		CreatedAt:    createdSession.CreatedAt,
		UserID:       createdSession.UserID,
		InvalidAfter: createdSession.InvalidAfter,
	}
	userArgs.ActiveGameID = payload.Game.ID
	modelGame, _, err := modelFromGame(ctx, payload.Game)
	if err != nil {
		return nil, fmt.Errorf("failed to create model from game")
	}
	playState, err := fromPlayState(payload.Game.PlayState)
	if err != nil {
		return nil, err
	}
	rule, err := p.ensureRuleExists(ctx, q, payload.Game.Rules)
	if err != nil {
		return nil, fmt.Errorf("failed in ensureRuleExists: %w", err)
	} else if rule.ID == "" {
		return nil, fmt.Errorf("failed to ensure the rule-set exists, the ID was not set")
	}
	insertGameParams := sqlite.InsertGameParams{
		ID:          modelGame.ID,
		CreatedAt:   modelGame.CreatedAt,
		UpdatedAt:   modelGame.UpdatedAt,
		UserID:      createdUser.ID,
		RuleID:      rule.ID,
		Score:       int64(modelGame.Score),
		Moves:       int64(modelGame.Moves),
		Description: sqlString(payload.Game.Description),
		PlayState:   playState,
		Data:        modelGame.Data,
	}
	createdGame, err := q.InsertGame(ctx, insertGameParams)
	if err != nil {
		return nil, fmt.Errorf("failed to insert game: %w", err)
	}
	mode, err := toMode(rule.Mode)
	if err != nil {
		return nil, fmt.Errorf("failed to map rule: %w", err)
	}
	sessionUser.ActiveGame = &types.Game{
		ID:        createdGame.ID,
		CreatedAt: createdGame.CreatedAt,
		UpdatedAt: &createdGame.UpdatedAt.Time,
		UserID:    createdGame.UserID,
		Seed:      payload.Game.Seed,
		State:     payload.Game.State,
		Score:     uint64(createdGame.Score),
		Moves:     uint(createdGame.Moves),
		Cells:     payload.Game.Cells,
		PlayState: payload.Game.PlayState,
		Rules: types.Rules{
			ID:              rule.ID,
			CreatedAt:       rule.CreatedAt,
			UpdatedAt:       &rule.UpdatedAt.Time,
			Description:     rule.Description.String,
			Mode:            mode,
			Rows:            uint8(rule.SizeY),
			Columns:         uint8(rule.SizeX),
			RecreateOnSwipe: rule.RecreateOnSwipe,
			NoReSwipe:       rule.NoReswipe,
			NoMultiply:      rule.NoMultiply,
			NoAddition:      rule.NoAddition,
		},
	}

	err = tx.Commit()
	return &sessionUser, err
}

func sqlString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
func (p *sqliteStorage) ensureRuleExists(ctx context.Context, q *sqlite.Queries, r types.Rules) (sqlite.Rule, error) {
	slug := r.Hash()
	existing := p.ruleCache.getCachedRule(slug)
	if existing != nil {
		return *existing, nil
	}
	x, err := q.GetRule(ctx, sqlite.GetRuleParams{
		// ID:   r.ID,
		Slug: slug,
	})
	if err == nil {
		p.ruleCache.addRulesToCache([]sqlite.Rule{x})
		return x, nil
	}

	if !errIsSqlNoRows(err) {
		return sqlite.Rule{}, err
	}
	mode, err := fromMode(r.Mode)
	if err != nil {
		return sqlite.Rule{}, err
	}
	insertParams := sqlite.InsertRuleParams{
		ID:              r.ID,
		Slug:            slug,
		CreatedAt:       time.Now(),
		Description:     sqlString(r.Description),
		Mode:            mode,
		SizeX:           int64(r.Columns),
		SizeY:           int64(r.Rows),
		RecreateOnSwipe: r.RecreateOnSwipe,
		NoReswipe:       r.NoReSwipe,
		NoMultiply:      r.NoMultiply,
		NoAddition:      r.NoAddition,
	}
	// doesn't seem to be supporting RETURN just yet: https://github.com/kyleconroy/sqlc/pull/1741
	return q.InsertRule(ctx, insertParams)
}

func toPlayState(p int64) (types.PlayState, error) {
	switch p {
	case PlayStateCurrent:
		return types.PlayStateCurrent, nil
	case PlayStateWon:
		return types.PlayStateWon, nil
	case PlayStateLost:
		return types.PlayStateLost, nil
	case PlayStateAbandoned:
		return types.PlayStateAbandoned, nil
	}
	return "", fmt.Errorf("unknown playstate: %d", p)
}
func fromPlayState(p types.PlayState) (PlayState, error) {
	switch p {
	case types.PlayStateCurrent:
		return PlayStateCurrent, nil
	case types.PlayStateWon:
		return PlayStateWon, nil
	case types.PlayStateLost:
		return PlayStateLost, nil
	case types.PlayStateAbandoned:
		return PlayStateAbandoned, nil
	}
	return -1, fmt.Errorf("unknown playstate: '%s'", p)
}
func toMode(p int64) (types.RuleMode, error) {
	switch p {
	case RuleModeInfiniteEasy:
		return types.RuleModeInfiniteEasy, nil
	case RuleModeInfiniteNormal:
		return types.RuleModeInfiniteNormal, nil
	case RuleModeInfiniteHard:
		return types.RuleModeInfiniteHard, nil
	case RuleModeChallenge:
		return types.RuleModeChallenge, nil
	case RuleModeTutorial:
		return types.RuleModeTutorial, nil
	}
	return "", fmt.Errorf("unknown mode: %d", p)
}
func fromMode(p types.RuleMode) (RuleMode, error) {
	switch p {
	case types.RuleModeInfiniteEasy:
		return RuleModeInfiniteEasy, nil
	case types.RuleModeInfiniteNormal:
		return RuleModeInfiniteNormal, nil
	case types.RuleModeInfiniteHard:
		return RuleModeInfiniteHard, nil
	case types.RuleModeChallenge:
		return RuleModeChallenge, nil
	case types.RuleModeTutorial:
		return RuleModeTutorial, nil
	}
	return -1, fmt.Errorf("unknown mode: '%s'", p)
}

type PlayState = int64
type RuleMode = int64
type InstructionKind = int64

const (
	PlayStateWon PlayState = iota + 1
	PlayStateLost
	PlayStateAbandoned
	PlayStateCurrent

	RuleModeInfiniteEasy RuleMode = iota + 1
	RuleModeInfiniteNormal
	RuleModeInfiniteHard
	RuleModeChallenge
	RuleModeTutorial
	InstructionKindSwipe InstructionKind = iota + 1
	InstructionKindCombine
)

func (p *sqliteStorage) GetUserBySessionID(ctx context.Context, payload types.GetUserPayload) (su *types.SessionUser, err error) {
	ctx, span := tracerSqlite.Start(ctx, "fetchRules")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	sess, err := p.queries.GetUserBySessionID(ctx, payload.ID)

	if err != nil {
		if errIsSqlNoRows(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failure in lookup user by session-id: %w", err)
	}

	rule := p.ruleCache.getCachedRule(sess.RuleID)
	if rule == nil {
		r, err := p.queries.GetRule(ctx, sqlite.GetRuleParams{
			ID: sess.RuleID,
		})
		if err != nil {
			return su, fmt.Errorf("failure in retrieving rule: %w", err)
		}
		if r.ID != "" {
			rule = &r
			p.ruleCache.addRulesToCache([]sqlite.Rule{r})
		}
	}
	if rule == nil {
		return nil, fmt.Errorf("the rule '%s' from the user-session was not found", sess.RuleID)
	}
	playState, err := toPlayState(sess.PlayState)
	mode, err := toMode(rule.Mode)
	cells, seed, state, err := UnmarshalInternalDataGame(ctx, sess.Data)
	if err != nil {
		return nil, err
	}
	// This is a bit ugly, await solutions in
	// https://github.com/kyleconroy/sqlc/issues/1630
	// for now, the conflicting types are named by their ordering in queries.sql
	tUser := types.User{
		ID:        sess.ID_2,
		CreatedAt: sess.CreatedAt_2,
		// session does not have an UpdatedAt-field, so the suffix-count is off by one
		UpdatedAt: &sess.UpdatedAt.Time,
		UserName:  sess.Username,
	}

	tRule := types.Rules{
		ID:              rule.ID,
		CreatedAt:       rule.CreatedAt,
		UpdatedAt:       &rule.UpdatedAt.Time,
		Description:     rule.Description.String,
		Mode:            mode,
		Rows:            uint8(rule.SizeX),
		Columns:         uint8(rule.SizeY),
		RecreateOnSwipe: rule.RecreateOnSwipe,
		NoReSwipe:       rule.NoReswipe,
		NoMultiply:      rule.NoMultiply,
		NoAddition:      rule.NoAddition,
	}

	tGame := &types.Game{
		ID:        sess.ID_3,
		CreatedAt: sess.CreatedAt_3,
		// session does not have an UpdatedAt-field, so the suffix-count is off by one
		UpdatedAt:   &sess.UpdatedAt_2.Time,
		Description: sess.Description.String,
		UserID:      tUser.ID,
		Seed:        seed,
		State:       state,
		Score:       uint64(sess.Score),
		Moves:       uint(sess.Moves),
		Cells:       cells,
		PlayState:   playState,
		Rules:       tRule,
	}
	su = &types.SessionUser{
		Session: types.Session{
			ID:           sess.ID,
			CreatedAt:    sess.CreatedAt,
			UserID:       tUser.ID,
			InvalidAfter: sess.InvalidAfter,
			ActiveGame:   tGame,
		},
		User: tUser,
	}
	return su, nil
}
func (p *sqliteStorage) CombinePath(ctx context.Context, payload types.CombinePathPayload) (err error) {
	ctx, span := tracerSqlite.Start(ctx, "CombinePath")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
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
	q, tx, err := p.beginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	dataHistory, err := MarshalInternalDataHistory(ctx, payload.State, payload.Cells, &instr)
	if err != nil {
		return fmt.Errorf("%s %w: dataHistory", err, ErrArgumentInvalid)
	}
	dataGame, err := MarshalInternalDataGame(ctx, payload.Seed, payload.State, payload.Cells)
	if err != nil {
		return fmt.Errorf("%s %w: dataGame", err, ErrArgumentInvalid)
	}
	g, err := q.GetGame(ctx, payload.GameID)
	if err != nil {
		return fmt.Errorf("failed to find game %w", err)
	}
	// Sanity check
	if g.Score+int64(payload.Points) != int64(payload.Score) {
		return fmt.Errorf("mismatch between score and points")
	}
	g.Score = int64(payload.Points)
	g.UpdatedAt = toNullTimeNonNullable(time.Now())
	updateGameArgs := sqlite.UpdateGameParams{
		UpdatedAt: sql.NullTime{},
		UserID:    g.UserID,
		RuleID:    g.RuleID,
		Score:     g.Score + int64(payload.Points),
		Moves:     g.Moves + 1,
		PlayState: g.PlayState,
		Data:      dataGame,
		ID:        g.ID,
	}
	updatedGame, err := q.UpdateGame(ctx, updateGameArgs)
	if err != nil {
		return fmt.Errorf("failed to update the game")
	}

	inserGameHistoryArgs := sqlite.InsertGameHistoryParams{
		CreatedAt: time.Now(),
		GameID:    updatedGame.ID,
		Move:      updatedGame.Moves,
		Kind:      InstructionKindCombine,
		Points:    int64(payload.Points),
		Data:      dataHistory,
	}
	_, err = q.InsertGameHistory(ctx, inserGameHistoryArgs)
	if err != nil {
		return fmt.Errorf("failed to save game-history: %w", err)
	}
	err = tx.Commit()
	return err
}
func (p *sqliteStorage) SwipeBoard(ctx context.Context, payload types.SwipePayload) (err error) {
	ctx, span := tracerSqlite.Start(ctx, "SwipeBoard")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	q, tx, err := p.beginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := payload.Validate(); err != nil {
		return err
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
	gameData, err := MarshalInternalDataGame(ctx, payload.Seed, payload.State, payload.Cells)
	if err != nil {
		return fmt.Errorf("%s %w: Cells", err, ErrArgumentInvalid)
	}
	g, err := q.GetGame(ctx, payload.GameID)
	if err != nil {
		return fmt.Errorf("failed to find game %w", err)
	}
	updateGameArgs := sqlite.UpdateGameParams{
		UpdatedAt: toNullTimeNonNullable(time.Now()),
		UserID:    g.UserID,
		RuleID:    g.RuleID,
		Score:     g.Score,
		Moves:     g.Moves + 1,
		PlayState: g.PlayState,
		Data:      gameData,
		ID:        g.ID,
	}
	updatedGame, err := q.UpdateGame(ctx, updateGameArgs)
	if err != nil {
		return fmt.Errorf("failed to update the game")
	}

	inserGameHistoryArgs := sqlite.InsertGameHistoryParams{
		CreatedAt: time.Now(),
		GameID:    updatedGame.ID,
		Move:      updatedGame.Moves,
		Kind:      InstructionKindSwipe,
		Points:    0,
		Data:      historyData,
	}
	_, err = q.InsertGameHistory(ctx, inserGameHistoryArgs)
	if err != nil {
		return fmt.Errorf("failed to save game-history: %w", err)
	}
	err = tx.Commit()
	return err
}
func (p *sqliteStorage) NewGameForUser(ctx context.Context, payload types.NewGamePayload) (tg types.Game, err error) {
	ctx, span := tracerSqlite.Start(ctx, "NewGameForUser")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	q, tx, err := p.beginTx(ctx)
	if err != nil {
		return tg, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	// TODO: simplyfy with custom sql-code.
	if payload.Game.ID == "" {
		return tg, fmt.Errorf("%w: Game.Id", ErrArgumentRequired)
	}
	if payload.Game.UserID == "" {
		return tg, fmt.Errorf("%w: UserID", ErrArgumentRequired)
	}
	if len(payload.Game.Cells) == 0 {
		return tg, fmt.Errorf("%w: Game.Cells", ErrArgumentRequired)
	}
	u, err := q.GetUser(ctx, payload.Game.UserID)
	if err != nil {
		return tg, fmt.Errorf("failed to to retrieve user %w", err)
	}

	if u.ActiveGameID != "" {
		activeGame, err := q.GetGame(ctx, u.ActiveGameID)
		if err != nil {
			return tg, fmt.Errorf("failed to to retrieve activegame for user %w", err)
		}
		if activeGame.PlayState == PlayStateCurrent {
			activeGame.PlayState = PlayStateAbandoned
			activeGame.UpdatedAt = toNullTimeNonNullable(time.Now())
			params := sqlite.UpdateGameParams{
				UpdatedAt: toNullTimeNonNullable(time.Now()),
				UserID:    activeGame.UserID,
				RuleID:    activeGame.RuleID,
				Score:     activeGame.Score,
				Moves:     activeGame.Moves,
				PlayState: PlayStateAbandoned,
				Data:      activeGame.Data,
				ID:        activeGame.ID,
			}
			_, err := q.UpdateGame(ctx, params)
			if err != nil {
				return tg, fmt.Errorf("failed to to update activegame for user %w", err)
			}
		}
	}
	modelGame, _, err := modelFromGame(ctx, payload.Game)
	if err != nil {
		return tg, err
	}
	r, err := p.ensureRuleExists(ctx, q, payload.Game.Rules)
	if err != nil {
		return tg, fmt.Errorf("failed to save the rules for the game: %w", err)
	} else {
		modelGame.RuleID = r.ID
	}
	playState, err := fromPlayState(payload.Game.PlayState)
	if err != nil {
		return tg, fmt.Errorf("failed to convert playstate: %w", err)
	}
	gameParams := sqlite.InsertGameParams{
		ID:          payload.Game.ID,
		CreatedAt:   createdAt(payload.Game.CreatedAt),
		UserID:      u.ID,
		RuleID:      r.ID,
		Score:       int64(payload.Game.Score),
		Moves:       int64(payload.Game.Moves),
		Description: sqlString(payload.Game.Description),
		PlayState:   playState,
		Data:        modelGame.Data,
	}
	mode, err := toMode(r.Mode)
	if err != nil {
		return tg, fmt.Errorf("failed to map rule: %w", err)
	}
	createdGame, err := q.InsertGame(ctx, gameParams)
	if err != nil {
		return tg, fmt.Errorf("failed to save the game for: %w", err)
	}
	u.ActiveGameID = createdGame.ID

	u.UpdatedAt = toNullTimeNonNullable(time.Now())
	updateUserParams := sqlite.UpdateUserParams{
		UpdatedAt:    toNullTimeNonNullable(time.Now()),
		Username:     u.Username,
		ActiveGameID: createdGame.ID,
		ID:           u.ID,
	}
	updatedUser, err := q.UpdateUser(ctx, updateUserParams)
	if err != nil {
		return tg, fmt.Errorf("failed to update userut: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return tg, err
	}
	tg = types.Game{
		ID:          createdGame.ID,
		CreatedAt:   createdGame.CreatedAt,
		Description: createdGame.Description.String,
		// session does not have an UpdatedAt-field, so the suffix-count is off by one
		UpdatedAt: &createdGame.UpdatedAt.Time,
		UserID:    updatedUser.ID,
		Seed:      payload.Game.Seed,
		State:     payload.Game.State,
		Score:     uint64(createdGame.Score),
		Moves:     uint(createdGame.Moves),
		Cells:     payload.Game.Cells,
		PlayState: payload.Game.PlayState,
		Rules: types.Rules{
			ID:              r.ID,
			CreatedAt:       r.CreatedAt,
			UpdatedAt:       &r.UpdatedAt.Time,
			Description:     r.Description.String,
			Mode:            mode,
			Rows:            uint8(r.SizeY),
			Columns:         uint8(r.SizeX),
			RecreateOnSwipe: r.RecreateOnSwipe,
			NoReSwipe:       r.NoReswipe,
			NoMultiply:      r.NoMultiply,
			NoAddition:      r.NoAddition,
		},
	}

	return tg, err
}

// ensures the date is set
// if the supplied date is valid, it is returned
// otherwise, the current time is used
func createdAt(t time.Time) time.Time {
	// A simple "null"-check
	if t.Year() > 2000 {
		return t
	}
	return time.Now()
}
