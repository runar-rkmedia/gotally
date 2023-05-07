package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/runar-rkmedia/go-common/logger"
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
func (p *sqliteStorage) Stats(ctx context.Context) (sess *types.Statistics, err error) {

	ctx, span := tracerSqlite.Start(ctx, "Stats")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	stats, err := p.queries.Stats(ctx)
	if err != nil {
		return nil, err
	}
	s := types.Statistics{
		Users:          stats.Users,
		Session:        stats.Session,
		Games:          stats.Games,
		GamesWon:       stats.GamesWon,
		GamesLost:      stats.GamesLost,
		GamesAbandoned: stats.GamesAbandoned,
		GamesCurrent:   stats.GamesCurrent,
		// LongestGame:    stats.LongestGame.(uint64),
		// HighestScore:   stats.HighestScore.(uint64),
		// HistoryDataStdDev: math.Sqrt(float64(stats.HistoryDataStdDev)),
		// CombineDataAvg: stats.CombineDataAvg.Float64,
		// CombineDataMax: stats.CombineDataMax.(uint64),
		// CombineDataMin: stats.CombineDataMin.(uint64),
	}
	if v, err := toUint64(stats.LongestGame); err != nil {
		return nil, fmt.Errorf("failure for type LongestGame: %w", err)
	} else {
		s.LongestGame = v
	}
	if v, err := toUint64(stats.HighestScore); err != nil {
		return nil, fmt.Errorf("failure for type HighestScore: %w", err)
	} else {
		s.HighestScore = v
	}
	return &s, err
}

func toUint64(v any) (uint64, error) {
	if v == nil {
		return 0, nil
	}
	switch n := v.(type) {
	case int64:
		return uint64(n), nil
	case uint64:
		return n, nil
	default:
		return 0, fmt.Errorf("unhandled toUint64-cast for type %T", v)
	}
}
func toFloat64(v any) (float64, error) {
	if v == nil {
		return 0, nil
	}
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	default:
		return 0, fmt.Errorf("unhandled toFloat64 for type %T", v)
	}
}

func (p *sqliteStorage) GetGameChallenges(ctx context.Context, payload types.GetGameChallengePayload) (response []types.GameTemplate, err error) {
	ctx, span := tracerSqlite.Start(ctx, "GetGameChallenges")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	q, tx, err := p.beginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	list, err := q.GetGameChallengesTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get game-challenges: %w", err)
	}
	stats, err := q.GetChallengeStatsForUser(ctx, payload.StatsForUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats for game-challenges")
	}
	response = make([]types.GameTemplate, len(list))
	p.fetchRules(ctx)
	for i := 0; i < len(list); i++ {
		r := p.ruleCache.getCachedRule(list[i].RuleID)
		if r == nil {
			rr, err := q.GetRule(ctx, sqlite.GetRuleParams{ID: list[i].RuleID})
			if err != nil {
				return response, fmt.Errorf("failed to fetch rule")
			}
			p.ruleCache.addRulesToCache([]sqlite.Rule{rr})
			if rr.ID == "" {
				return response, fmt.Errorf("expected to find the rule with id '%s' in the database, but it was not found", list[i].RuleID)
			}
			r = &rr
		}
		rule, err := toTypeRule(*r)
		if err != nil {
			return response, err
		}
		cells, _, _, err := UnmarshalInternalDataGame(ctx, list[i].Data)
		response[i] = types.GameTemplate{
			ID:              list[i].ID,
			CreatedAt:       list[i].CreatedAt,
			UpdatedAt:       fromNullTime(list[i].UpdatedAt),
			ChallengeNumber: nullIntToIntP(list[i].ChallengeNumber),
			IdealMoves:      nullIntToIntP(list[i].IdealMoves),
			CreatedByID:     list[i].CreatedBy,
			UpdatedBy:       list[i].UpdatedBy.String,
			Description:     list[i].Description.String,
			Name:            list[i].Name,
			Cells:           cells,
			Rules:           rule,
		}
		for j := 0; j < len(stats); j++ {
			if stats[j].TemplateID.String != list[i].ID {
				continue
			}
			if response[i].Stats == nil {
				response[i].Stats = []types.PlayStats{}
			}
			response[i].Stats = append(response[i].Stats, types.PlayStats{
				GameID:   stats[j].GameID,
				UserID:   stats[j].UserID,
				Username: stats[j].Username,
				Score:    uint64(stats[j].Score),
				Moves:    uint64(stats[j].Moves),
			})

		}

	}

	return
}
func (p *sqliteStorage) CreateGameTemplate(ctx context.Context, payload types.CreateGameTemplatePayload) (response *types.GameTemplate, err error) {

	ctx, span := tracerSqlite.Start(ctx, "CreateGameTemplate")
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

	rule, err := p.ensureRuleExists(ctx, q, payload.Rules)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure rule existance: %w", err)
	}

	data, err := MarshalInternalDataGame(ctx, 0, 0, payload.Cells)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal datagame: %w", err)
	}
	templateArgs := sqlite.InserTemplateParams{
		ID:          payload.ID,
		CreatedAt:   payload.CreatedAt,
		RuleID:      rule.ID,
		CreatedBy:   payload.CreatedByID,
		Name:        payload.Name,
		Description: sqlString(payload.Description),
		IdealMoves:  toNullInt64(uint64(payload.IdealMoves)),
		IdealScore:  toNullInt64(uint64(payload.IdealScore)),
		Data:        data,
	}
	if payload.ChallengeNumber != nil {
		templateArgs.ChallengeNumber.Int64 = int64(*payload.ChallengeNumber)
		templateArgs.ChallengeNumber.Valid = true
	}
	t, err := q.InserTemplate(ctx, templateArgs)
	if err != nil {
		return nil, fmt.Errorf("Failed to insert template: %w", err)
	}

	typeRule, err := toTypeRule(rule)
	if err != nil {
		return nil, fmt.Errorf("failed to map rule: %w", err)
	}
	response = &types.GameTemplate{
		ID:              t.ID,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       fromNullTime(t.UpdatedAt),
		ChallengeNumber: nullIntToIntP(t.ChallengeNumber),
		CreatedByID:     t.CreatedBy,
		UpdatedBy:       t.UpdatedBy.String,
		Description:     t.Description.String,
		Name:            t.Name,
		Cells:           payload.Cells,
		Rules:           typeRule,
	}
	err = tx.Commit()
	return response, err

}

// Creates a user, session, game and makes sure the rule exists.
// This should only be used for new users, not to log in existing users.
func (p *sqliteStorage) CreateUserSession(ctx context.Context, payload types.CreateUserSessionPayload) (sess *types.SessionUser, err error) {
	if payload.TemplateID == "" {
		panic("TODO: add templateid to all CreateUserSession-calls")
	}
	ctx, span := tracerSqlite.Start(ctx, "CreateUserSession")
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
	data, err := MarshalInternalDataGame(ctx, payload.Game.Seed, payload.Game.State, payload.Game.Cells)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal datagame: %w", err)
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
		ID:          payload.Game.ID,
		CreatedAt:   payload.Game.CreatedAt,
		UpdatedAt:   toNullTime(payload.Game.UpdatedAt),
		UserID:      createdUser.ID,
		RuleID:      rule.ID,
		Score:       int64(payload.Game.Score),
		Moves:       int64(payload.Game.Moves),
		Name:        sqlString(payload.Game.Name),
		Description: sqlString(payload.Game.Description),
		PlayState:   playState,
		Data:        data,
		DataAtStart: data,
		History:     []byte{},
		TemplateID:  toNullString(payload.TemplateID),
	}
	createdGame, err := q.InsertGame(ctx, insertGameParams)
	if err != nil {
		return nil, fmt.Errorf("failed to insert game: %w", err)
	}
	activeGame, err := toTypeGame(&createdGame, &rule, payload.Game.Seed, payload.Game.State, payload.Game.Cells, payload.Game.PlayState)
	if err != nil {
		return nil, err
	}
	sessionUser.ActiveGame = &activeGame
	if err := sessionUser.ActiveGame.Validate(); err != nil {
		return nil, err
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
		ID:   r.ID,
		Slug: slug,
	})
	if err == nil {
		p.ruleCache.addRulesToCache([]sqlite.Rule{x})
		return x, nil
	}

	if !errIsSqlNoRows(err) {
		return sqlite.Rule{}, fmt.Errorf("failed to retrive rule %#v: %w", r, err)
	}
	mode, err := fromMode(r.Mode)
	if err != nil {
		return sqlite.Rule{}, fmt.Errorf("failed to convert rule.Mode from rule %#v: %w", r, err)
	}
	if r.ID == "" {
		// TODO: is it valid at this point to have a rule-id?
		// return sqlite.Rule{}, fmt.Errorf("The id of the rule cannot be empty")
		r.ID = createID()
	}
	// TODO: this check may not belong here.
	if r.Mode == types.RuleModeChallenge && r.TargetCellValue == 0 {
		return sqlite.Rule{}, fmt.Errorf("mode %s requires TargetCellValue", r.Mode)
	}
	insertParams := sqlite.InsertRuleParams{
		ID:              r.ID,
		Slug:            slug,
		CreatedAt:       time.Now(),
		Description:     sqlString(r.Description),
		Mode:            int64(mode),
		SizeX:           int64(r.Columns),
		SizeY:           int64(r.Rows),
		RecreateOnSwipe: r.RecreateOnSwipe,
		NoReswipe:       r.NoReSwipe,
		NoMultiply:      r.NoMultiply,
		NoAddition:      r.NoAddition,
		MaxMoves:        toNullInt64(r.MaxMoves),
		TargetCellValue: toNullInt64(r.TargetCellValue),
		TargetScore:     toNullInt64(r.TargetCellValue),
	}
	if insertParams.SizeX == 0 || insertParams.SizeY == 0 {
		return sqlite.Rule{}, fmt.Errorf("The rules has invalid size: %#v", insertParams)
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
	switch RuleMode(p) {
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
	return "", fmt.Errorf("unknown mode during toMode: %d", p)
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
	return -1, fmt.Errorf("unknown mode during fromMode: '%s'", p)
}

type PlayState = int64
type RuleMode int64
type InstructionKind = int64

const (
	PlayStateWon PlayState = iota + 1
	PlayStateLost
	PlayStateAbandoned
	PlayStateCurrent
)
const (
	RuleModeInfiniteEasy RuleMode = iota + 1
	RuleModeInfiniteNormal
	RuleModeInfiniteHard
	RuleModeChallenge
	RuleModeTutorial
)
const (
	InstructionKindSwipe InstructionKind = iota + 1
	InstructionKindCombine
	InstructionKindInit
)

func (mode RuleMode) String() string {
	switch mode {
	case RuleModeInfiniteEasy:
		return fmt.Sprintf("infinite easy (%d)", mode)
	case RuleModeInfiniteNormal:
		return fmt.Sprintf("infinite normal (%d)", mode)
	case RuleModeInfiniteHard:
		return fmt.Sprintf("infinite hard (%d)", mode)
	case RuleModeChallenge:
		return fmt.Sprintf("challenge (%d)", mode)
	case RuleModeTutorial:
		return fmt.Sprintf("tutorial (%d)", mode)
	}
	return fmt.Sprintf("err: Invalid rulemode: %d", mode)
}
func (mode RuleMode) MarshalJSON() ([]byte, error) {
	// return []byte(`"banana"`), nil
	return json.Marshal(mode.String())
}

func nullIntToUint(i sql.NullInt64) uint64 {
	if !i.Valid {
		return 0
	}
	return uint64(i.Int64)
}
func nullIntToIntP(i sql.NullInt64) *int {
	if !i.Valid {
		return nil
	}
	n := int(i.Int64)
	return &n
}

func (p *sqliteStorage) GetUserBySessionID(ctx context.Context, payload types.GetUserPayload) (su *types.SessionUser, err error) {
	ctx, span := tracerSqlite.Start(ctx, "fetchRules")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	if err := payload.Validate(); err != nil {
		return nil, err
	}
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
	if err != nil {
		return nil, err
	}
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

	tRule, err := toTypeRule(*rule)
	if err != nil {
		return su, fmt.Errorf("failed to map rule %w", err)
	}

	tGame := &types.Game{
		ID:        sess.ID_3,
		CreatedAt: sess.CreatedAt_3,
		// session does not have an UpdatedAt-field, so the suffix-count is off by one
		UpdatedAt:   &sess.UpdatedAt_2.Time,
		Name:        sess.Name.String,
		Description: sess.Description.String,
		UserID:      tUser.ID,
		Seed:        seed,
		State:       state,
		Score:       uint64(sess.Score),
		Moves:       uint(sess.Moves),
		Cells:       cells,
		History:     sess.History,
		PlayState:   playState,
		Rules:       tRule,
	}
	if err := tGame.Validate(); err != nil {
		return nil, err
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
func (p *sqliteStorage) UpdateGame(ctx context.Context, payload types.UpdateGamePayload) (err error) {
	ctx, span := tracerSqlite.Start(ctx, "CombinePath")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	if err := payload.Validate(); err != nil {
		return err
	}
	q, tx, err := p.beginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	dataGame, err := MarshalInternalDataGame(ctx, payload.Seed, payload.State, payload.Cells)
	if err != nil {
		return fmt.Errorf("%s %w: dataGame", err, ErrArgumentInvalid)
	}
	g, err := q.GetGame(ctx, payload.GameID)
	if err != nil {
		return fmt.Errorf("failed to find game %w", err)
	}
	g.UpdatedAt = toNullTimeNonNullable(time.Now())
	playState, err := fromPlayState(payload.PlayState)
	if err != nil {
		return fmt.Errorf("failed to map playstate")
	}
	updateGameArgs := sqlite.UpdateGameParams{
		UpdatedAt: toNullTimeNonNullable(time.Now()),
		UserID:    g.UserID,
		RuleID:    g.RuleID,
		Score:     int64(payload.Score),
		Moves:     int64(payload.Moves),
		PlayState: playState,
		Data:      dataGame,
		History:   payload.History,
		ID:        g.ID,
	}
	// if len(updateGameArgs.History) == 0 {
	// 	return fmt.Errorf("attempted to update game, but there was no history-field")
	// }
	if len(updateGameArgs.Data) == 0 {
		return fmt.Errorf("attempted to update game, but there was no data-field")
	}
	if payload.PlayState != "" {
		ps, err := fromPlayState(payload.PlayState)
		if err != nil {
			return fmt.Errorf("Failed converting playstate: %w", err)
		}
		updateGameArgs.PlayState = ps
	}
	updated, err := q.UpdateGame(ctx, updateGameArgs)
	if err != nil {
		return fmt.Errorf("failed to update the game")
	}
	if updateGameArgs.Moves != updated.Moves {
		return fmt.Errorf("Did not expect moves to be zero. (this is a temporary check, and should be removed in the future)")
	}

	err = tx.Commit()
	return err
}

func (p *sqliteStorage) Dump(ctx context.Context) (tg types.Dump, err error) {
	ctx, span := tracerSqlite.Start(ctx, "Dump")
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

	games, err := q.GetAllGames(ctx)
	if err != nil {
		return tg, fmt.Errorf("failed to retrieve all games")
	}
	tg.Games = games
	rules, err := q.GetAllRules(ctx)
	if err != nil {
		return tg, fmt.Errorf("failed to retrieve all rules")
	}
	tg.Rules = rules
	sessions, err := q.GetAllSessions(ctx)
	if err != nil {
		return tg, fmt.Errorf("failed to retrieve all sessions")
	}
	tg.Sessions = sessions
	users, err := q.GetAllUsers(ctx)
	if err != nil {
		return tg, fmt.Errorf("failed to retrieve all users")
	}
	tg.Users = users
	templates, err := q.GetAllTemplates(ctx)
	if err != nil {
		return tg, fmt.Errorf("failed to retrieve all templates")
	}
	tg.Template = templates

	return
}

func (p *sqliteStorage) GetOriginalGame(ctx context.Context, payload types.GetOriginalGamePayload) (tg types.Game, err error) {

	ctx, span := tracerSqlite.Start(ctx, "RestartGame")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	if err := payload.Validate(); err != nil {
		return types.Game{}, err
	}
	q, tx, err := p.beginTx(ctx)
	if err != nil {
		return tg, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	g, err := q.GetGame(ctx, payload.GameID)
	if err != nil {
		return tg, fmt.Errorf("failed to to retrieve game (payload %#v): %w", payload, err)
	}
	playstate, err := toPlayState(g.PlayState)
	if err != nil {
		return tg, fmt.Errorf("invalid gameParams.PlayState: %w", err)
	}
	rule, err := p.ensureRuleExists(ctx, q, types.Rules{ID: g.RuleID})
	if err != nil {
		return tg, fmt.Errorf("failed to retrieve rule with id '%s': %w", g.RuleID, err)
	}
	cells, seed, state, err := UnmarshalInternalDataGame(ctx, g.DataAtStart)
	fmt.Println("Got original game", cells)
	return toTypeGame(&g, &rule, seed, state, cells, playstate)

}
func (p *sqliteStorage) RestartGame(ctx context.Context, payload types.RestartGamePayload) (tg types.Game, err error) {

	ctx, span := tracerSqlite.Start(ctx, "RestartGame")
	defer func() {
		AnnotateSpanError(span, err)
		span.End()
	}()
	if payload.GameID == "" {
		return tg, fmt.Errorf("%w: GameID", ErrArgumentRequired)
	}
	if payload.UserID == "" {
		return tg, fmt.Errorf("%w: UserID", ErrArgumentRequired)
	}
	q, tx, err := p.beginTx(ctx)
	if err != nil {
		return tg, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	g, err := q.GetGame(ctx, payload.GameID)
	if err != nil {
		return tg, fmt.Errorf("failed to to retrieve game (payload %#v): %w", payload, err)
	}
	if g.ID != payload.GameID {
		return tg, fmt.Errorf("The game '%s' did not match supplied gameid %s", g.UserID, payload.GameID)
	}
	if g.UserID != payload.UserID {
		return tg, fmt.Errorf("The game '%s' did not match supplied userid %s, expected %s", g.ID, payload.UserID, g.UserID)
	}
	if g.Moves == 0 {
		return tg, fmt.Errorf("Cannot restart a game already at the start")
	}
	rule, err := p.ensureRuleExists(ctx, q, types.Rules{ID: g.RuleID})
	if err != nil {
		return tg, fmt.Errorf("failed to retrieve rule with id '%s': %w", g.RuleID, err)
	}
	if rule.ID == "" {
		return tg, fmt.Errorf("the returned rules was unexpectedly empty: %#v", rule)
	}
	//
	if len(g.History) == 0 {
		return tg, fmt.Errorf("the game's history contained no data for this move (payload: %#v): %#v", payload, g.History)
	}
	cells, seed, state, err := UnmarshalInternalDataGame(ctx, g.DataAtStart)
	newData, err := MarshalInternalDataGame(ctx, seed, state, cells)
	if err != nil {
		return tg, fmt.Errorf("failed to UnmarshalInternalDataHistory from gamehistory (payload %#v): %w", payload, err)
	}

	if g.PlayState == PlayStateCurrent {
		g.PlayState = PlayStateAbandoned
		g.UpdatedAt = toNullTimeNonNullable(time.Now())
		params := sqlite.SetPlayStateForGameParams{
			UpdatedAt: g.UpdatedAt,
			PlayState: PlayStateAbandoned,
			ID:        g.ID,
		}
		_, err := q.SetPlayStateForGame(ctx, params)
		if err != nil {
			return tg, fmt.Errorf("failed to to update activegame for user %w", err)
		}
	}
	var originalGameID string
	if g.BasedOnGame.Valid && g.BasedOnGame.String != "" {
		originalGameID = g.BasedOnGame.String
	} else {
		originalGameID = g.ID

	}
	gameParams := sqlite.InsertGameParams{
		ID:          createID(),
		CreatedAt:   time.Now(),
		UserID:      g.UserID,
		RuleID:      g.RuleID,
		Score:       0,
		Moves:       0,
		History:     []byte{},
		Description: g.Description,
		Name:        g.Name,
		PlayState:   PlayStateCurrent,
		Data:        newData,
		DataAtStart: newData,
		BasedOnGame: toNullString(originalGameID),
		TemplateID:  g.TemplateID,
	}
	playstate, err := toPlayState(gameParams.PlayState)
	if err != nil {
		return tg, fmt.Errorf("invalid gameParams.PlayState: %w", err)
	}

	createdGame, err := q.InsertGame(ctx, gameParams)
	if err != nil {
		return tg, fmt.Errorf("failed to save the game for: %w", err)
	}
	updateUserParams := sqlite.SetActiveGameFormUserParams{
		UpdatedAt:    toNullTimeNonNullable(time.Now()),
		ActiveGameID: createdGame.ID,
		ID:           g.UserID,
	}
	_, err = q.SetActiveGameFormUser(ctx, updateUserParams)
	if err != nil {
		return tg, fmt.Errorf("failed to update userut: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return tg, err
	}
	return toTypeGame(&createdGame, &rule, seed, state, cells, playstate)
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
	return p.newGameForUser(ctx, q, tx, payload)
}
func (p *sqliteStorage) newGameForUser(ctx context.Context, q *sqlite.Queries, tx *sql.Tx, payload types.NewGamePayload) (tg types.Game, err error) {
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
				History:   activeGame.History,
				Data:      activeGame.Data,
				ID:        activeGame.ID,
			}
			_, err := q.UpdateGame(ctx, params)
			if err != nil {
				return tg, fmt.Errorf("failed to to update activegame for user %w", err)
			}
		}
	}
	data, err := MarshalInternalDataGame(ctx, payload.Game.Seed, payload.Game.State, payload.Game.Cells)
	if err != nil {
		return tg, fmt.Errorf("failed to marshal datagame: %w", err)
	}
	r, err := p.ensureRuleExists(ctx, q, payload.Game.Rules)
	if err != nil {
		return tg, fmt.Errorf("failed to save the rules for the game: %w", err)
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
		Name:        sqlString(payload.Game.Name),
		Description: sqlString(payload.Game.Description),
		PlayState:   playState,
		Data:        data,
		DataAtStart: data,
		TemplateID:  toNullString(payload.TemplateID),
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
	_, err = q.UpdateUser(ctx, updateUserParams)
	if err != nil {
		return tg, fmt.Errorf("failed to update userut: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return tg, err
	}
	return toTypeGame(&createdGame, &r, payload.Game.Seed, payload.Game.State, payload.Game.Cells, payload.Game.PlayState)
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
