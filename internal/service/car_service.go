package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"carsapi/internal/models"
	"carsapi/internal/repository"
)

type CarService interface {
	Create(ctx context.Context, car *models.Car) (*models.Car, error)
	GetByID(ctx context.Context, id int64) (*models.Car, error)
	GetAll(ctx context.Context) ([]*models.Car, error)
	Update(ctx context.Context, car *models.Car) (*models.Car, error)
	Delete(ctx context.Context, id int64) error
}

type carService struct {
	repo repository.CarRepository
}

func NewCarService(repo repository.CarRepository) CarService {
	return &carService{repo: repo}
}

func (s *carService) Create(ctx context.Context, car *models.Car) (*models.Car, error) {
	if err := validateCar(car); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, car); err != nil {
		return nil, err
	}

	created, err := s.repo.GetByID(ctx, car.ID)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *carService) GetByID(ctx context.Context, id int64) (*models.Car, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrValidation)
	}

	car, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCarNotFound
	}
	if err != nil {
		return nil, err
	}

	return car, nil
}

func (s *carService) GetAll(ctx context.Context) ([]*models.Car, error) {
	return s.repo.GetAll(ctx)
}

func (s *carService) Update(ctx context.Context, car *models.Car) (*models.Car, error) {
	if car.ID <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrValidation)
	}

	if err := validateCar(car); err != nil {
		return nil, err
	}

	err := s.repo.Update(ctx, car)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCarNotFound
	}
	if err != nil {
		return nil, err
	}

	updated, err := s.repo.GetByID(ctx, car.ID)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *carService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrValidation)
	}

	err := s.repo.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCarNotFound
	}

	return err
}

func validateCar(car *models.Car) error {
	if car == nil {
		return fmt.Errorf("%w: car payload is required", ErrValidation)
	}
	if car.InventoryID <= 0 {
		return fmt.Errorf("%w: inventory_id must be positive", ErrValidation)
	}
	if strings.TrimSpace(car.Make) == "" {
		return fmt.Errorf("%w: make is required", ErrValidation)
	}
	if strings.TrimSpace(car.Model) == "" {
		return fmt.Errorf("%w: model is required", ErrValidation)
	}
	if car.Year < 1886 || car.Year > 2100 {
		return fmt.Errorf("%w: year must be between 1886 and 2100", ErrValidation)
	}
	if strings.TrimSpace(car.Color) == "" {
		return fmt.Errorf("%w: color is required", ErrValidation)
	}
	if strings.TrimSpace(car.VIN) == "" {
		return fmt.Errorf("%w: vin is required", ErrValidation)
	}

	return nil
}
