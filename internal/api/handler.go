package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"carsapi/internal/models"
	"carsapi/internal/service"
)

type CarHandler struct {
	service service.CarService
}

func NewCarHandler(svc service.CarService) *CarHandler {
	return &CarHandler{service: svc}
}

func (h *CarHandler) HandleCars(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listCars(w, r)
	case http.MethodPost:
		h.createCar(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *CarHandler) HandleCarByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path, "/api/cars/")
	if err != nil {
		writeFail(w, http.StatusBadRequest, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getCarByID(w, r, id)
	case http.MethodPut:
		h.updateCar(w, r, id)
	case http.MethodDelete:
		h.deleteCar(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *CarHandler) listCars(w http.ResponseWriter, r *http.Request) {
	cars, err := h.service.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch cars")
		return
	}

	writeSuccess(w, http.StatusOK, cars)
}

func (h *CarHandler) getCarByID(w http.ResponseWriter, r *http.Request, id int64) {
	car, err := h.service.GetByID(r.Context(), id)
	if errors.Is(err, service.ErrCarNotFound) {
		writeError(w, http.StatusNotFound, "car not found")
		return
	}
	if errors.Is(err, service.ErrValidation) {
		writeFail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch car")
		return
	}

	writeSuccess(w, http.StatusOK, car)
}

func (h *CarHandler) createCar(w http.ResponseWriter, r *http.Request) {
	var in models.Car
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeFail(w, http.StatusBadRequest, "invalid json body")
		return
	}

	created, err := h.service.Create(r.Context(), &in)
	if errors.Is(err, service.ErrValidation) {
		writeFail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create car")
		return
	}

	writeSuccess(w, http.StatusCreated, created)
}

func (h *CarHandler) updateCar(w http.ResponseWriter, r *http.Request, id int64) {
	var in models.Car
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeFail(w, http.StatusBadRequest, "invalid json body")
		return
	}
	in.ID = id

	updated, err := h.service.Update(r.Context(), &in)
	if errors.Is(err, service.ErrValidation) {
		writeFail(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrCarNotFound) {
		writeError(w, http.StatusNotFound, "car not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update car")
		return
	}

	writeSuccess(w, http.StatusOK, updated)
}

func (h *CarHandler) deleteCar(w http.ResponseWriter, r *http.Request, id int64) {
	err := h.service.Delete(r.Context(), id)
	if errors.Is(err, service.ErrValidation) {
		writeFail(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, service.ErrCarNotFound) {
		writeError(w, http.StatusNotFound, "car not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete car")
		return
	}

	writeSuccess(w, http.StatusOK, map[string]bool{"deleted": true})
}

func parseID(path, prefix string) (int64, error) {
	value := strings.TrimPrefix(path, prefix)
	if value == "" || strings.Contains(value, "/") {
		return 0, fmt.Errorf("invalid id")
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}

	return id, nil
}
