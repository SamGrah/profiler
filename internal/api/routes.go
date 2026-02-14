package api

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *CarHandler) {
	mux.HandleFunc("/api/cars", handler.HandleCars)
	mux.HandleFunc("/api/cars/", handler.HandleCarByID)
}
