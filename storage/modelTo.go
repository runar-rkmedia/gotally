package storage

import (
	"database/sql"
	"errors"
	"time"

	tallyv1 "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
	"github.com/runar-rkmedia/gotally/sqlite"
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
func fromNullTime(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
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

func toTypeRule(rule sqlite.Rule) (types.Rules, error) {
	mode, err := toMode(rule.Mode)
	if err != nil {
		return types.Rules{}, err
	}
	return types.Rules{
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
	}, nil
}

var (
	ErrArgumentRequired = errors.New("argument required")
	ErrArgumentInvalid  = errors.New("argument invalid")
)
