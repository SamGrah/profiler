package repository

import (
	"context"
	"database/sql"
)

type SQLDBAdapter struct {
	db *sql.DB
}

func NewSQLDBAdapter(db *sql.DB) *SQLDBAdapter {
	return &SQLDBAdapter{db: db}
}

func (a *SQLDBAdapter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return a.db.ExecContext(ctx, query, args...)
}

func (a *SQLDBAdapter) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return a.db.QueryRowContext(ctx, query, args...)
}

func (a *SQLDBAdapter) QueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	return a.db.QueryContext(ctx, query, args...)
}
