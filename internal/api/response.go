package api

import (
	"encoding/json"
	"net/http"
)

type jsendResponse struct {
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func writeSuccess(w http.ResponseWriter, code int, data any) {
	writeJSON(w, code, jsendResponse{Status: "success", Data: data})
}

func writeFail(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, jsendResponse{Status: "fail", Message: message})
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, jsendResponse{Status: "error", Message: message})
}

func writeJSON(w http.ResponseWriter, code int, payload jsendResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
