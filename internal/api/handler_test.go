package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"carsapi/internal/models"
	"carsapi/internal/service"
)

type fakeCarService struct {
	cars   map[int64]*models.Car
	nextID int64
}

func newFakeCarService() *fakeCarService {
	return &fakeCarService{cars: map[int64]*models.Car{}, nextID: 1}
}

func (f *fakeCarService) Create(_ context.Context, car *models.Car) (*models.Car, error) {
	if car.InventoryID <= 0 || car.Make == "" || car.Model == "" || car.Year == 0 || car.Color == "" || car.VIN == "" {
		return nil, service.ErrValidation
	}
	car.ID = f.nextID
	f.nextID++
	copyCar := *car
	f.cars[car.ID] = &copyCar
	return &copyCar, nil
}

func (f *fakeCarService) GetByID(_ context.Context, id int64) (*models.Car, error) {
	car, ok := f.cars[id]
	if !ok {
		return nil, service.ErrCarNotFound
	}
	copyCar := *car
	return &copyCar, nil
}

func (f *fakeCarService) GetAll(_ context.Context) ([]*models.Car, error) {
	out := make([]*models.Car, 0, len(f.cars))
	for id := int64(1); id <= f.nextID; id++ {
		if car, ok := f.cars[id]; ok {
			copyCar := *car
			out = append(out, &copyCar)
		}
	}
	return out, nil
}

func (f *fakeCarService) Update(_ context.Context, car *models.Car) (*models.Car, error) {
	if _, ok := f.cars[car.ID]; !ok {
		return nil, service.ErrCarNotFound
	}
	if car.InventoryID <= 0 || car.Make == "" || car.Model == "" || car.Year == 0 || car.Color == "" || car.VIN == "" {
		return nil, service.ErrValidation
	}
	copyCar := *car
	f.cars[car.ID] = &copyCar
	return &copyCar, nil
}

func (f *fakeCarService) Delete(_ context.Context, id int64) error {
	if _, ok := f.cars[id]; !ok {
		return service.ErrCarNotFound
	}
	delete(f.cars, id)
	return nil
}

func TestCreateCarHandler(t *testing.T) {
	h := NewCarHandler(newFakeCarService())
	body := []byte(`{"inventory_id":1,"make":"Ford","model":"Fiesta","year":2018,"color":"Blue","vin":"VIN-API-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/cars", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.HandleCars(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestGetCarByIDHandlerNotFound(t *testing.T) {
	h := NewCarHandler(newFakeCarService())
	req := httptest.NewRequest(http.MethodGet, "/api/cars/99", nil)
	rec := httptest.NewRecorder()

	h.HandleCarByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestListCarsHandler(t *testing.T) {
	fake := newFakeCarService()
	_, _ = fake.Create(context.Background(), &models.Car{InventoryID: 1, Make: "BMW", Model: "M3", Year: 2021, Color: "Black", VIN: "VIN-API-2"})
	h := NewCarHandler(fake)

	req := httptest.NewRequest(http.MethodGet, "/api/cars", nil)
	rec := httptest.NewRecorder()
	h.HandleCars(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp jsendResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response error = %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("status = %q, want success", resp.Status)
	}
}

func TestUpdateAndDeleteHandlers(t *testing.T) {
	fake := newFakeCarService()
	created, _ := fake.Create(context.Background(), &models.Car{InventoryID: 1, Make: "Kia", Model: "Soul", Year: 2020, Color: "Yellow", VIN: "VIN-API-3"})
	h := NewCarHandler(fake)

	updateBody := []byte(`{"inventory_id":1,"make":"Kia","model":"Soul","year":2020,"color":"Orange","vin":"VIN-API-3"}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/cars/"+toString(created.ID), bytes.NewReader(updateBody))
	updateRec := httptest.NewRecorder()
	h.HandleCarByID(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d", updateRec.Code, http.StatusOK)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/cars/"+toString(created.ID), nil)
	deleteRec := httptest.NewRecorder()
	h.HandleCarByID(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want %d", deleteRec.Code, http.StatusOK)
	}

	if err := fake.Delete(context.Background(), created.ID); !errors.Is(err, service.ErrCarNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func toString(id int64) string {
	return strconv.FormatInt(id, 10)
}
