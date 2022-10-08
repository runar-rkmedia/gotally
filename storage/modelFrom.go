package storage

import (
	"github.com/runar-rkmedia/gotally/models"
	"github.com/runar-rkmedia/gotally/types"
)

func modelFromSessionUser(s types.SessionUser) (models.User, models.Session, *models.Game, *models.Rule, error) {
	mu := modelFromUser(s.User)
	ms, mg, mr, err := modelFromSession(s.Session)
	return mu, ms, mg, mr, err
}
func modelFromUser(u types.User) models.User {
	user := models.User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: toNullTime(u.UpdatedAt),
		Username:  u.UserName,
	}
	return user
}

func modelFromSession(s types.Session) (models.Session, *models.Game, *models.Rule, error) {
	ms := models.Session{
		ID:           s.ID,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    toNullTime(s.UpdatedAt),
		InvalidAfter: s.InvalidAfter,
		UserID:       s.UserID,
	}
	if s.ActiveGame == nil {
		return ms, nil, nil, nil
	}
	mg, mr, err := modelFromGame(*s.ActiveGame)
	return ms, &mg, &mr, err

}
func modelFromGame(g types.Game) (models.Game, models.Rule, error) {
	cells, err := MarshalCellValues(g.Cells)
	if err != nil {
		return models.Game{}, models.Rule{}, err
	}
	mg := models.Game{
		ID:        g.ID,
		CreatedAt: g.CreatedAt,
		UpdatedAt: toNullTime(g.UpdatedAt),
		UserID:    g.UserID,
		RuleID:    g.Rules.ID,
		Seed:      g.Seed,
		State:     g.State,
		Score:     g.Score,
		Moves:     g.Moves,
		PlayState: modelFromPlayState(g.PlayState),
		Cells:     cells,
	}
	mr := modelFromRule(g.Rules)
	mg.RuleID = mr.ID
	return mg, mr, err
}
func modelFromRule(r types.Rules) models.Rule {
	session := models.Rule{
		ID:              r.ID,
		Slug:            r.Hash(),
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       toNullTime(r.UpdatedAt),
		Description:     r.Description,
		Mode:            modelFromRuleMode(r.Mode),
		SizeX:           r.Columns,
		SizeY:           r.Rows,
		RecreateOnSwipe: r.RecreateOnSwipe,
		NoReswipe:       r.NoReSwipe,
		NoMultiply:      r.NoMultiply,
		NoAddition:      r.NoAddition,
	}
	return session
}

func modelFromRuleMode(m types.RuleMode) models.Mode {
	switch m {
	case types.RuleModeChallenge:
		return models.ModeChallenge
	case types.RuleModeInfiniteEasy:
		return models.ModeInfiniteEasy
	case types.RuleModeInfiniteNormal:
		return models.ModeInfiniteNormal
	case types.RuleModeInfiniteHard:
		return models.ModeInfiniteHard
	}
	return 0
}

func modelFromPlayStateNull(p types.PlayState) models.NullPlayState {
	m := modelFromPlayState(p)
	return models.NullPlayState{
		PlayState: m,
		Valid:     m != 0,
	}

}
func modelFromRuleModeNull(p types.RuleMode) models.NullMode {
	m := modelFromRuleMode(p)
	return models.NullMode{
		Mode:  m,
		Valid: m != 0,
	}

}
func modelFromPlayState(p types.PlayState) models.PlayState {

	switch p {
	case types.PlayStateAbandoned:
		return models.PlayStateAbandoned
	case types.PlayStateCurrent:
		return models.PlayStateCurrent
	case types.PlayStateLost:
		return models.PlayStateLost
	case types.PlayStateWon:
		return models.PlayStateWon
	}
	return 0
}
