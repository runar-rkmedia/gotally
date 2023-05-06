package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/runar-rkmedia/gotally/sqlite"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
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
		TargetCellValue: uint64(rule.TargetCellValue.Int64),
		TargetScore:     uint64(rule.TargetScore.Int64),
		MaxMoves:        uint64(rule.MaxMoves.Int64),
		Rows:            uint8(rule.SizeY),
		Columns:         uint8(rule.SizeX),
		RecreateOnSwipe: rule.RecreateOnSwipe,
		NoReSwipe:       rule.NoReswipe,
		NoMultiply:      rule.NoMultiply,
		NoAddition:      rule.NoAddition,
	}, nil
}
func toTypeGame(createdGame *sqlite.Game, r *sqlite.Rule, seed, state uint64, cells []cell.Cell, playState string) (types.Game, error) {

	tRule, err := toTypeRule(*r)
	if err != nil {
		return types.Game{}, fmt.Errorf("failed to map rule %w", err)
	}
	tg := types.Game{
		ID:          createdGame.ID,
		CreatedAt:   createdGame.CreatedAt,
		Description: createdGame.Description.String,
		Name:        createdGame.Name.String,
		// session does not have an UpdatedAt-field, so the suffix-count is off by one
		UpdatedAt: &createdGame.UpdatedAt.Time,
		UserID:    createdGame.UserID,
		Seed:      seed,
		State:     state,
		Score:     uint64(createdGame.Score),
		Moves:     uint(createdGame.Moves),
		History:   createdGame.History,
		Cells:     cells,
		PlayState: playState,
		Rules:     tRule,
	}
	if err := tg.Validate(); err != nil {
		return tg, err
	}
	return tg, nil
}

var (
	ErrArgumentRequired = errors.New("argument required")
	ErrArgumentInvalid  = errors.New("argument invalid")
)
