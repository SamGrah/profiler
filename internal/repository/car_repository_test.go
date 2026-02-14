package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"carsapi/internal/models"
)

type fakeResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r fakeResult) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r fakeResult) RowsAffected() (int64, error) { return r.rowsAffected, nil }

type fakeRow struct {
	values []any
	err    error
}

func (r *fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	for i := range dest {
		switch d := dest[i].(type) {
		case *int64:
			*d = r.values[i].(int64)
		case *int:
			*d = r.values[i].(int)
		case *string:
			*d = r.values[i].(string)
		default:
			return errors.New("unsupported scan destination")
		}
	}

	return nil
}

type fakeRows struct {
	index  int
	values [][]any
	err    error
}

func (r *fakeRows) Next() bool {
	return r.index < len(r.values)
}

func (r *fakeRows) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}

	row := r.values[r.index]
	for i := range dest {
		switch d := dest[i].(type) {
		case *int64:
			*d = row[i].(int64)
		case *int:
			*d = row[i].(int)
		case *string:
			*d = row[i].(string)
		default:
			return errors.New("unsupported scan destination")
		}
	}
	r.index++
	return nil
}

func (r *fakeRows) Close() error { return nil }
func (r *fakeRows) Err() error   { return r.err }

type fakeDB struct {
	nextID int64
	cars   map[int64]*models.Car
}

func newFakeDB() *fakeDB {
	return &fakeDB{nextID: 1, cars: map[int64]*models.Car{}}
}

func (f *fakeDB) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	switch query {
	case createCarQuery:
		id := f.nextID
		f.nextID++
		f.cars[id] = &models.Car{
			ID:          id,
			InventoryID: args[0].(int64),
			Make:        args[1].(string),
			Model:       args[2].(string),
			Year:        args[3].(int),
			Color:       args[4].(string),
			VIN:         args[5].(string),
		}
		return fakeResult{lastInsertID: id, rowsAffected: 1}, nil
	case updateCarQuery:
		id := args[6].(int64)
		car, ok := f.cars[id]
		if !ok {
			return fakeResult{rowsAffected: 0}, nil
		}
		car.InventoryID = args[0].(int64)
		car.Make = args[1].(string)
		car.Model = args[2].(string)
		car.Year = args[3].(int)
		car.Color = args[4].(string)
		car.VIN = args[5].(string)
		return fakeResult{rowsAffected: 1}, nil
	case deleteCarQuery:
		id := args[0].(int64)
		if _, ok := f.cars[id]; !ok {
			return fakeResult{rowsAffected: 0}, nil
		}
		delete(f.cars, id)
		return fakeResult{rowsAffected: 1}, nil
	default:
		return nil, errors.New("unsupported query")
	}
}

func (f *fakeDB) QueryRowContext(_ context.Context, query string, args ...any) Row {
	if query != getCarByIDQuery {
		return &fakeRow{err: errors.New("unsupported query")}
	}
	id := args[0].(int64)
	car, ok := f.cars[id]
	if !ok {
		return &fakeRow{err: sql.ErrNoRows}
	}

	return &fakeRow{values: []any{car.ID, car.InventoryID, car.Make, car.Model, car.Year, car.Color, car.VIN}}
}

func (f *fakeDB) QueryContext(_ context.Context, query string, _ ...any) (Rows, error) {
	if query != getAllCarsQuery {
		return nil, errors.New("unsupported query")
	}

	values := make([][]any, 0, len(f.cars))
	for id := int64(1); id <= f.nextID; id++ {
		if car, ok := f.cars[id]; ok {
			values = append(values, []any{car.ID, car.InventoryID, car.Make, car.Model, car.Year, car.Color, car.VIN})
		}
	}

	return &fakeRows{values: values}, nil
}

func TestSQLiteCarRepositoryCreateAndGetByID(t *testing.T) {
	repo := NewSQLiteCarRepository(newFakeDB())
	ctx := context.Background()

	car := &models.Car{InventoryID: 1, Make: "Honda", Model: "Civic", Year: 2020, Color: "Blue", VIN: "VIN-1"}
	if err := repo.Create(ctx, car); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(ctx, car.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.VIN != "VIN-1" {
		t.Fatalf("GetByID() VIN = %q, want %q", got.VIN, "VIN-1")
	}
}

func TestSQLiteCarRepositoryGetAll(t *testing.T) {
	repo := NewSQLiteCarRepository(newFakeDB())
	ctx := context.Background()

	_ = repo.Create(ctx, &models.Car{InventoryID: 1, Make: "Ford", Model: "Focus", Year: 2019, Color: "Gray", VIN: "VIN-2"})
	_ = repo.Create(ctx, &models.Car{InventoryID: 1, Make: "Toyota", Model: "Corolla", Year: 2021, Color: "White", VIN: "VIN-3"})

	cars, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(cars) != 2 {
		t.Fatalf("GetAll() len = %d, want 2", len(cars))
	}
}

func TestSQLiteCarRepositoryUpdate(t *testing.T) {
	repo := NewSQLiteCarRepository(newFakeDB())
	ctx := context.Background()

	car := &models.Car{InventoryID: 1, Make: "Nissan", Model: "Sentra", Year: 2018, Color: "Red", VIN: "VIN-4"}
	_ = repo.Create(ctx, car)

	car.Color = "Black"
	if err := repo.Update(ctx, car); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, _ := repo.GetByID(ctx, car.ID)
	if got.Color != "Black" {
		t.Fatalf("Update() color = %q, want %q", got.Color, "Black")
	}
}

func TestSQLiteCarRepositoryDelete(t *testing.T) {
	repo := NewSQLiteCarRepository(newFakeDB())
	ctx := context.Background()

	car := &models.Car{InventoryID: 1, Make: "Mazda", Model: "3", Year: 2022, Color: "Green", VIN: "VIN-5"}
	_ = repo.Create(ctx, car)

	if err := repo.Delete(ctx, car.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := repo.GetByID(ctx, car.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetByID() error = %v, want sql.ErrNoRows", err)
	}
}
