package storage

import (
	"database/sql"
	"time"

	"github.com/runar-rkmedia/gotally/models"
	"github.com/runar-rkmedia/gotally/types"
)

func modelToSessionUser(u models.User, s models.Session, activeGame *models.Game, rules *models.Rule) (types.SessionUser, error) {
	sess, err := modelToSession(s, activeGame, rules)
	if err != nil {
		return types.SessionUser{}, err
	}
	su := types.SessionUser{
		Session: sess,
		User:    modelToUser(u),
	}
	return su, nil
}
func modelToUser(u models.User) types.User {
	session := types.User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: nullTime(u.UpdatedAt),
		UserName:  u.Username,
	}
	return session
}

func modelToSession(s models.Session, activeGame *models.Game, r *models.Rule) (types.Session, error) {
	session := types.Session{
		ID:           s.ID,
		CreatedAt:    s.CreatedAt,
		UserID:       s.UserID,
		UpdatedAt:    nullTime(s.UpdatedAt),
		InvalidAfter: s.InvalidAfter,
	}
	if activeGame != nil {
		g, err := modelToGame(*activeGame, *r)
		if err != nil {
			return session, err
		}
		session.ActiveGame = &g
	}
	return session, nil
}
func modelToGame(s models.Game, r models.Rule) (types.Game, error) {
	cells, err := UnmarshalCellValues(s.Cells)
	if err != nil {
		return types.Game{}, err
	}
	session := types.Game{
		ID:        s.ID,
		CreatedAt: s.CreatedAt,
		UpdatedAt: nullTime(s.UpdatedAt),
		UserID:    s.UserID,
		Seed:      s.Seed,
		State:     s.State,
		Score:     s.Score,
		Moves:     s.Moves,
		PlayState: "",
		Rules:     modelToRule(r),
		Cells:     cells,
	}
	return session, err
}
func modelToRule(r models.Rule) types.Rules {
	session := types.Rules{
		ID:              r.ID,
		CreatedAt:       r.CreatedAt,
		Mode:            modelToRuleMode(r.Mode),
		Rows:            r.SizeX,
		Columns:         r.SizeY,
		RecreateOnSwipe: r.RecreateOnSwipe,
		NoReSwipe:       r.NoReswipe,
		NoMultiply:      r.NoMultiply,
		NoAddition:      r.NoAddition,
	}
	return session
}

func modelToRuleMode(m models.Mode) types.RuleMode {
	switch m {
	case models.ModeChallenge:
		return types.RuleModeChallenge
	case models.ModeInfiniteEasy:
		return types.RuleModeInfiniteEasy
	case models.ModeInfiniteNormal:
		return types.RuleModeInfiniteNormal
	case models.ModeInfiniteHard:
		return types.RuleModeInfiniteHard
	}
	return ""
}

func modelToPlayState(m models.NullPlayState) types.PlayState {
	if !m.Valid {
		return ""
	}
	switch m.PlayState {
	case models.PlayStateAbandoned:
		return types.PlayStateAbandoned
	case models.PlayStateCurrent:
		return types.PlayStateCurrent
	case models.PlayStateLost:
		return types.PlayStateLost
	case models.PlayStateWon:
		return types.PlayStateWon
	}
	return ""
}

func nullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
func toNullTimeNonNullable(t time.Time) sql.NullTime {
	return toNullTime(&t)
}
func toNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{
		Valid: true,
		Time:  *t,
	}

}
