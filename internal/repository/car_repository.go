package repository

import (
	"context"
	"database/sql"

	"carsapi/internal/models"
)

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) Row
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
}

type CarRepository interface {
	Create(ctx context.Context, car *models.Car) error
	GetByID(ctx context.Context, id int64) (*models.Car, error)
	GetAll(ctx context.Context) ([]*models.Car, error)
	Update(ctx context.Context, car *models.Car) error
	Delete(ctx context.Context, id int64) error
}
