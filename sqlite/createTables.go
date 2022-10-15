package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
)

var (
	//go:embed schema.sql
	schema string
)

func (q *Queries) InitializeDatabase(ctx context.Context) (sql.Result, error) {
	return q.db.ExecContext(ctx, schema)

}
