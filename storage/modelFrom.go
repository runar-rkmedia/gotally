package storage

import (
	"context"

	"github.com/runar-rkmedia/gotally/models"
	"github.com/runar-rkmedia/gotally/types"
)

func modelFromGame(ctx context.Context, g types.Game) (models.Game, models.Rule, error) {
	data, err := MarshalInternalDataGame(ctx, g.Seed, g.State, g.Cells)
	if err != nil {
		return models.Game{}, models.Rule{}, err
	}
	mg := models.Game{
		ID:        g.ID,
		CreatedAt: g.CreatedAt,
		UpdatedAt: toNullTime(g.UpdatedAt),
		UserID:    g.UserID,
		RuleID:    g.Rules.ID,
		Score:     g.Score,
		Moves:     g.Moves,
		PlayState: modelFromPlayState(g.PlayState),
		Data:      data,
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
