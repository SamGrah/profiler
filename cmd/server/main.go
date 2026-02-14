package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	"carsapi/internal/api"
	"carsapi/internal/repository"
	"carsapi/internal/service"
	_ "modernc.org/sqlite"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	dbPath := flag.String("db-path", "cars.db", "SQLite database path")
	schemaPath := flag.String("schema-path", "db/schema.sql", "Path to SQL schema file")
	flag.Parse()

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("enable foreign keys: %v", err)
	}

	if err := applySchema(db, *schemaPath); err != nil {
		log.Fatalf("apply schema: %v", err)
	}

	repo := repository.NewSQLiteCarRepository(repository.NewSQLDBAdapter(db))
	svc := service.NewCarService(repo)
	h := api.NewCarHandler(svc)

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, h)

	log.Printf("server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

func applySchema(db *sql.DB, schemaPath string) error {
	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(schemaSQL))
	return err
}
