package storage

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/runar-rkmedia/gotally/models"
	"github.com/runar-rkmedia/gotally/types"
)

func modelFromSessionUser(ctx context.Context, s types.SessionUser) (models.User, models.Session, *models.Game, *models.Rule, error) {
	mu := modelFromUser(s.User)
	ms, mg, mr, err := modelFromSession(ctx, s.Session)
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

func modelFromSession(ctx context.Context, s types.Session) (models.Session, *models.Game, *models.Rule, error) {
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
	mg, mr, err := modelFromGame(ctx, *s.ActiveGame)
	return ms, &mg, &mr, err

}
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

func uint64ToByteSlice(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}
func byteSliceToUint64(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, fmt.Errorf("expected byteslice for uin64 to be of length 8, but it was %d", len(b))
	}
	return binary.LittleEndian.Uint64(b), nil
}
