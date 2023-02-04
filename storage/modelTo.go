package storage

import (
	"database/sql"
	"errors"
	"time"

	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/types"
)

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
func toNullString(str string) sql.NullString {
	if str == "" {
		return sql.NullString{}
	}
	return sql.NullString{
		Valid:  true,
		String: str,
	}
}
func toNullInt64(n uint64) sql.NullInt64 {
	if n == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Valid: true,
		Int64: int64(n),
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
