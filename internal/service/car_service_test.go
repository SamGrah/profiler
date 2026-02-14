package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"carsapi/internal/models"
)

type fakeCarRepository struct {
	cars   map[int64]*models.Car
	nextID int64
}

func newFakeCarRepository() *fakeCarRepository {
	return &fakeCarRepository{cars: map[int64]*models.Car{}, nextID: 1}
}

func (f *fakeCarRepository) Create(_ context.Context, car *models.Car) error {
	car.ID = f.nextID
	f.nextID++
	copyCar := *car
	f.cars[car.ID] = &copyCar
	return nil
}

func (f *fakeCarRepository) GetByID(_ context.Context, id int64) (*models.Car, error) {
	car, ok := f.cars[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	copyCar := *car
	return &copyCar, nil
}

func (f *fakeCarRepository) GetAll(_ context.Context) ([]*models.Car, error) {
	out := make([]*models.Car, 0, len(f.cars))
	for id := int64(1); id <= f.nextID; id++ {
		if car, ok := f.cars[id]; ok {
			copyCar := *car
			out = append(out, &copyCar)
		}
	}
	return out, nil
}

func (f *fakeCarRepository) Update(_ context.Context, car *models.Car) error {
	if _, ok := f.cars[car.ID]; !ok {
		return sql.ErrNoRows
	}
	copyCar := *car
	f.cars[car.ID] = &copyCar
	return nil
}

func (f *fakeCarRepository) Delete(_ context.Context, id int64) error {
	if _, ok := f.cars[id]; !ok {
		return sql.ErrNoRows
	}
	delete(f.cars, id)
	return nil
}

func TestCarServiceCreate(t *testing.T) {
	svc := NewCarService(newFakeCarRepository())

	car := &models.Car{InventoryID: 1, Make: "Toyota", Model: "Camry", Year: 2023, Color: "White", VIN: "VIN-SVC-1"}
	created, err := svc.Create(context.Background(), car)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("Create() returned ID = 0")
	}
}

func TestCarServiceCreateValidation(t *testing.T) {
	svc := NewCarService(newFakeCarRepository())

	_, err := svc.Create(context.Background(), &models.Car{InventoryID: 1})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("Create() error = %v, want ErrValidation", err)
	}
}

func TestCarServiceGetByIDNotFound(t *testing.T) {
	svc := NewCarService(newFakeCarRepository())

	_, err := svc.GetByID(context.Background(), 999)
	if !errors.Is(err, ErrCarNotFound) {
		t.Fatalf("GetByID() error = %v, want ErrCarNotFound", err)
	}
}

func TestCarServiceUpdate(t *testing.T) {
	repo := newFakeCarRepository()
	svc := NewCarService(repo)
	ctx := context.Background()

	created, _ := svc.Create(ctx, &models.Car{InventoryID: 1, Make: "Tesla", Model: "Model 3", Year: 2022, Color: "Black", VIN: "VIN-SVC-2"})
	created.Color = "Red"

	updated, err := svc.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Color != "Red" {
		t.Fatalf("Update() color = %q, want Red", updated.Color)
	}
}

func TestCarServiceDelete(t *testing.T) {
	repo := newFakeCarRepository()
	svc := NewCarService(repo)
	ctx := context.Background()

	created, _ := svc.Create(ctx, &models.Car{InventoryID: 1, Make: "Audi", Model: "A4", Year: 2020, Color: "Gray", VIN: "VIN-SVC-3"})

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err := svc.GetByID(ctx, created.ID)
	if !errors.Is(err, ErrCarNotFound) {
		t.Fatalf("GetByID() error = %v, want ErrCarNotFound", err)
	}
}
