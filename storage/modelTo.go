package storage

import (
	"database/sql"
	"errors"
	"time"

	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/types"
)

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

var (
	ErrArgumentRequired = errors.New("argument required")
	ErrArgumentInvalid  = errors.New("argument invalid")
)
