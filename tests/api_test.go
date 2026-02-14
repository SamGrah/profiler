package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

type jsendResp struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

type car struct {
	ID          int64  `json:"id"`
	InventoryID int64  `json:"inventory_id"`
	Make        string `json:"make"`
	Model       string `json:"model"`
	Year        int    `json:"year"`
	Color       string `json:"color"`
	VIN         string `json:"vin"`
}

func TestCarsAPI_CRUDAgainstRunningServer(t *testing.T) {
	if os.Getenv("RUN_API_TESTS") != "1" {
		t.Skip("set RUN_API_TESTS=1 to run integration API tests")
	}

	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	vin := fmt.Sprintf("VIN-IT-%d", time.Now().UnixNano())

	created := createCar(t, baseURL, car{
		InventoryID: 1,
		Make:        "Toyota",
		Model:       "Prius",
		Year:        2024,
		Color:       "Silver",
		VIN:         vin,
	})

	fetched := getCar(t, baseURL, created.ID)
	if fetched.VIN != vin {
		t.Fatalf("GET /api/cars/{id} VIN=%q want %q", fetched.VIN, vin)
	}

	updatedVIN := fmt.Sprintf("VIN-IT-UPDATED-%d", time.Now().UnixNano())
	updated := updateCar(t, baseURL, created.ID, car{
		InventoryID: 1,
		Make:        "Toyota",
		Model:       "Prius",
		Year:        2025,
		Color:       "Blue",
		VIN:         updatedVIN,
	})
	if updated.Color != "Blue" {
		t.Fatalf("PUT /api/cars/{id} color=%q want Blue", updated.Color)
	}

	list := listCars(t, baseURL)
	found := false
	for _, c := range list {
		if c.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("GET /api/cars did not include created ID %d", created.ID)
	}

	deleteCar(t, baseURL, created.ID)

	resp, err := http.Get(fmt.Sprintf("%s/api/cars/%d", baseURL, created.ID))
	if err != nil {
		t.Fatalf("GET deleted car error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET deleted car status=%d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func createCar(t *testing.T, baseURL string, in car) car {
	t.Helper()
	body, _ := json.Marshal(in)
	resp, err := http.Post(baseURL+"/api/cars", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/cars error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/cars status=%d want %d", resp.StatusCode, http.StatusCreated)
	}

	var js jsendResp
	if err := json.NewDecoder(resp.Body).Decode(&js); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	var out car
	if err := json.Unmarshal(js.Data, &out); err != nil {
		t.Fatalf("unmarshal create data: %v", err)
	}
	return out
}

func getCar(t *testing.T, baseURL string, id int64) car {
	t.Helper()
	resp, err := http.Get(fmt.Sprintf("%s/api/cars/%d", baseURL, id))
	if err != nil {
		t.Fatalf("GET /api/cars/{id} error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/cars/{id} status=%d want %d", resp.StatusCode, http.StatusOK)
	}

	var js jsendResp
	if err := json.NewDecoder(resp.Body).Decode(&js); err != nil {
		t.Fatalf("decode get response: %v", err)
	}

	var out car
	if err := json.Unmarshal(js.Data, &out); err != nil {
		t.Fatalf("unmarshal get data: %v", err)
	}
	return out
}

func updateCar(t *testing.T, baseURL string, id int64, in car) car {
	t.Helper()
	body, _ := json.Marshal(in)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/cars/%d", baseURL, id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /api/cars/{id} error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT /api/cars/{id} status=%d want %d", resp.StatusCode, http.StatusOK)
	}

	var js jsendResp
	if err := json.NewDecoder(resp.Body).Decode(&js); err != nil {
		t.Fatalf("decode update response: %v", err)
	}

	var out car
	if err := json.Unmarshal(js.Data, &out); err != nil {
		t.Fatalf("unmarshal update data: %v", err)
	}
	return out
}

func listCars(t *testing.T, baseURL string) []car {
	t.Helper()
	resp, err := http.Get(baseURL + "/api/cars")
	if err != nil {
		t.Fatalf("GET /api/cars error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/cars status=%d want %d", resp.StatusCode, http.StatusOK)
	}

	var js jsendResp
	if err := json.NewDecoder(resp.Body).Decode(&js); err != nil {
		t.Fatalf("decode list response: %v", err)
	}

	var out []car
	if err := json.Unmarshal(js.Data, &out); err != nil {
		t.Fatalf("unmarshal list data: %v", err)
	}
	return out
}

func deleteCar(t *testing.T, baseURL string, id int64) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/cars/%d", baseURL, id), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /api/cars/{id} error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE /api/cars/{id} status=%d want %d", resp.StatusCode, http.StatusOK)
	}
}
