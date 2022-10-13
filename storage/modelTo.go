package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/models"
	"github.com/runar-rkmedia/gotally/types"
)

func modelToSessionUser(ctx context.Context, u models.User, s models.Session, activeGame *models.Game, rules *models.Rule) (types.SessionUser, error) {
	sess, err := modelToSession(ctx, s, activeGame, rules)
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

func modelToSession(ctx context.Context, s models.Session, activeGame *models.Game, r *models.Rule) (types.Session, error) {
	session := types.Session{
		ID:           s.ID,
		CreatedAt:    s.CreatedAt,
		UserID:       s.UserID,
		UpdatedAt:    nullTime(s.UpdatedAt),
		InvalidAfter: s.InvalidAfter,
	}
	if activeGame != nil {
		if r == nil {
			return session, fmt.Errorf("session has a game associated, but no rules.")
		}
		g, err := modelToGame(ctx, *activeGame, *r)
		if err != nil {
			return session, err
		}
		session.ActiveGame = &g
	}
	return session, nil
}
func modelToGame(ctx context.Context, s models.Game, r models.Rule) (types.Game, error) {
	cells, seed, state, err := UnmarshalInternalDataGame(ctx, s.Data)
	if err != nil {
		return types.Game{}, err
	}
	session := types.Game{
		ID:        s.ID,
		CreatedAt: s.CreatedAt,
		UpdatedAt: nullTime(s.UpdatedAt),
		UserID:    s.UserID,
		Seed:      seed,
		State:     state,
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

// func modelToPlayState(m models.NullPlayState) types.PlayState {
// 	if !m.Valid {
// 		return ""
// 	}
// 	switch m.PlayState {
// 	case models.PlayStateAbandoned:
// 		return types.PlayStateAbandoned
// 	case models.PlayStateCurrent:
// 		return types.PlayStateCurrent
// 	case models.PlayStateLost:
// 		return types.PlayStateLost
// 	case models.PlayStateWon:
// 		return types.PlayStateWon
// 	}
// 	return ""
// }

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

//	func toGameSwipeDirection(dir tallyv1.SwipeDirection) types.SwipeDirection {
//		switch dir {
//		case tallyv1.SwipeDirection_SWIPE_DIRECTION_UP:
//			return types.SwipeDirectionUp
//		case tallyv1.SwipeDirection_SWIPE_DIRECTION_RIGHT:
//			return types.SwipeDirectionRight
//		case tallyv1.SwipeDirection_SWIPE_DIRECTION_DOWN:
//			return types.SwipeDirectionDown
//		case tallyv1.SwipeDirection_SWIPE_DIRECTION_LEFT:
//			return types.SwipeDirectionLeft
//		}
//		return ""
//	}
func toModalDirection(dir types.SwipeDirection) tallyv1.SwipeDirection {
	switch dir {
	case types.SwipeDirectionUp:
		return tallyv1.SwipeDirection_SWIPE_DIRECTION_UP
	case types.SwipeDirectionRight:
		return tallyv1.SwipeDirection_SWIPE_DIRECTION_RIGHT
	case types.SwipeDirectionDown:
		return tallyv1.SwipeDirection_SWIPE_DIRECTION_DOWN
	case types.SwipeDirectionLeft:
		return tallyv1.SwipeDirection_SWIPE_DIRECTION_LEFT
	}
	return tallyv1.SwipeDirection_SWIPE_DIRECTION_UNSPECIFIED
}
