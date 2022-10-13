package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func GetAllRules(ctx context.Context, db DB) ([]Rule, error) {
	rules := []Rule{}
	sqlString := `SELECT id, slug, created_at, updated_at, description, mode, sizeX, sizeY, recreate_on_swipe, no_reswipe, no_multiply, no_addition
FROM tallyboard.rule t;`

	rows, err := db.QueryContext(ctx, sqlString)
	if err != nil {
		return rules, fmt.Errorf("failed to retrieve all rules: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		r := Rule{}
		err := rows.Scan(&r.ID, &r.Slug, &r.CreatedAt, &r.UpdatedAt, &r.Description, &r.Mode, &r.SizeX, &r.SizeY, &r.RecreateOnSwipe, &r.NoReswipe, &r.NoMultiply, &r.NoAddition)
		if err != nil {
			return rules, err
		}
		rules = append(rules, r)
	}
	return rules, err
}

func (r *Rule) CreateIfUnique(ctx context.Context, db DB) (*Rule, bool, error) {
	existing, err := RuleBySlugDescription(ctx, db, r.Slug, r.Description)
	if !errors.Is(err, sql.ErrNoRows) {
		return r, false, err

	}
	if existing != nil {
		return existing, false, nil
	}
	err = r.Insert(ctx, db)
	if err != nil {
		return r, false, err
	}
	return r, true, err

}
