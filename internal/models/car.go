package models

type Car struct {
	ID          int64  `json:"id"`
	InventoryID int64  `json:"inventory_id"`
	Make        string `json:"make"`
	Model       string `json:"model"`
	Year        int    `json:"year"`
	Color       string `json:"color"`
	VIN         string `json:"vin"`
}
