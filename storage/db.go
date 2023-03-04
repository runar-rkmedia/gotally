package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/XSAM/otelsql"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/xo/dburl"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func newDb(l logger.AppLogger, dsn string, withOpenTelemetry bool) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn must be set for the database-connection")
	}

	u, err := parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dsn-string: %w", err)
	}
	if l.HasDebug() {
		l.Debug().
			Str("host", u.Host).
			Str("path", u.Path).
			Str("user", u.User.Username()).
			Str("scheme", u.Scheme).
			Str("driver", u.Driver).
			Str("redacted", u.Redacted()).
			Msg("Attempting to connect to database with these details (some are hidden)")
	}
	// db, err := passfile.OpenURL(u, ".", "db")
	var attr attribute.KeyValue
	switch u.Driver {
	case "mysql":
		attr = semconv.DBSystemMySQL
	case "sqlite3":
		attr = semconv.DBSystemSqlite
	default:
		l.Warn().Str("driver", u.Driver).Msg("unmapped driver-attribute for opentelemetry")
	}

	db, err := otelsql.Open(u.Driver, u.DSN, otelsql.WithAttributes(attr), otelsql.WithSpanOptions(otelsql.SpanOptions{
		Ping:           true,
		RowsNext:       false,
		DisableErrSkip: true,
		DisableQuery:   false,
		RecordError: func(err error) bool {
			return !errors.Is(err, sql.ErrNoRows)
		},
		OmitConnResetSession: false,
		OmitConnPrepare:      false,
		OmitConnQuery:        false,
		OmitRows:             false,
		OmitConnectorConnect: false,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed in open db from url: %w", err)
	}
	if l.HasDebug() {
		l.Debug().
			Msg("Successfully connected to database")
	}

	err = db.Ping()
	if err != nil {
		return nil, err

	}
	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(attr))
	if err != nil {
		return nil, err
	}
	return db, err
}

func createID() string {
	return gonanoid.Must()
}

func errIsSqlNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

var (
	tracerMysql  = otel.Tracer("database")
	tracerSqlite = otel.Tracer("sqlite")
)

func parse(dsn string) (*dburl.URL, error) {
	v, err := dburl.Parse(dsn)
	if err != nil {
		return nil, err
	}
	switch v.Driver {
	case "mysql":
		q := v.Query()
		q.Set("parseTime", "true")
		v.RawQuery = q.Encode()
		return dburl.Parse(v.String())
	case "sqlite3":
		q := v.Query()
		q.Set("_loc", "auto")
		v.RawQuery = q.Encode()
		return dburl.Parse(v.String())
	}
	return v, nil
}
