package repository

import (
	"context"
	"database/sql"

	"carsapi/internal/models"
)

const (
	createCarQuery  = `INSERT INTO cars (inventory_id, make, model, year, color, vin) VALUES (?, ?, ?, ?, ?, ?)`
	getCarByIDQuery = `SELECT id, inventory_id, make, model, year, color, vin FROM cars WHERE id = ?`
	getAllCarsQuery = `SELECT id, inventory_id, make, model, year, color, vin FROM cars ORDER BY id ASC`
	updateCarQuery  = `UPDATE cars SET inventory_id = ?, make = ?, model = ?, year = ?, color = ?, vin = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	deleteCarQuery  = `DELETE FROM cars WHERE id = ?`
)

type SQLiteCarRepository struct {
	db DB
}

func NewSQLiteCarRepository(db DB) *SQLiteCarRepository {
	return &SQLiteCarRepository{db: db}
}

func (r *SQLiteCarRepository) Create(ctx context.Context, car *models.Car) error {
	result, err := r.db.ExecContext(ctx, createCarQuery, car.InventoryID, car.Make, car.Model, car.Year, car.Color, car.VIN)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	car.ID = id
	return nil
}

func (r *SQLiteCarRepository) GetByID(ctx context.Context, id int64) (*models.Car, error) {
	row := r.db.QueryRowContext(ctx, getCarByIDQuery, id)

	car := &models.Car{}
	err := row.Scan(&car.ID, &car.InventoryID, &car.Make, &car.Model, &car.Year, &car.Color, &car.VIN)
	if err != nil {
		return nil, err
	}

	return car, nil
}

func (r *SQLiteCarRepository) GetAll(ctx context.Context) ([]*models.Car, error) {
	rows, err := r.db.QueryContext(ctx, getAllCarsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cars := make([]*models.Car, 0)
	for rows.Next() {
		car := &models.Car{}
		if err := rows.Scan(&car.ID, &car.InventoryID, &car.Make, &car.Model, &car.Year, &car.Color, &car.VIN); err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cars, nil
}

func (r *SQLiteCarRepository) Update(ctx context.Context, car *models.Car) error {
	result, err := r.db.ExecContext(ctx, updateCarQuery, car.InventoryID, car.Make, car.Model, car.Year, car.Color, car.VIN, car.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *SQLiteCarRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, deleteCarQuery, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
