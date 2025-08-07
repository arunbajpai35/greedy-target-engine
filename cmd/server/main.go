package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/delivery"
)

func main() {
	// Connect to PostgreSQL
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/targeting_db?sslmode=disable")
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	defer db.Close()

	r := chi.NewRouter()

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Main delivery handler with DB injected
	r.Get("/v1/delivery", delivery.HandleDeliveryRequest(db))

	log.Println("✅ Server started on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
