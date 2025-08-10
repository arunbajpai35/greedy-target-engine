package delivery

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/campaigns"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/metrics"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

func HandleDeliveryRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Set response headers
		w.Header().Set("Content-Type", "application/json")

		// Validate request parameters
		req, errMsg := validateParams(r)
		if errMsg != "" {
			log.Printf("❌ Invalid request parameters: %s", errMsg)
			w.WriteHeader(http.StatusBadRequest)
			metrics.ObserveRequest("bad_request", time.Since(start).Seconds())
			json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
			return
		}

		// Get matching campaigns
		matched, err := campaigns.GetMatchingCampaigns(db, req.App, req.Country, req.OS)
		if err != nil {
			log.Printf("❌ Database query failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			metrics.ObserveRequest("error", time.Since(start).Seconds())
			json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			return
		}

		// Log request details
		log.Printf("📊 Request: app=%s, country=%s, os=%s, matches=%d, duration=%v",
			req.App, req.Country, req.OS, len(matched), time.Since(start))

		// Return appropriate response
		if len(matched) == 0 {
			w.WriteHeader(http.StatusNoContent)
			metrics.ObserveRequest("no_content", time.Since(start).Seconds())
			return
		}

		w.WriteHeader(http.StatusOK)
		metrics.ObserveRequest("ok", time.Since(start).Seconds())
		if err := json.NewEncoder(w).Encode(matched); err != nil {
			log.Printf("❌ Failed to encode response: %v", err)
		}
	}
}

// validateParams validates the required query parameters
func validateParams(r *http.Request) (models.DeliveryRequest, string) {
	app := strings.TrimSpace(r.URL.Query().Get("app"))
	country := strings.TrimSpace(r.URL.Query().Get("country"))
	os := strings.TrimSpace(r.URL.Query().Get("os"))

	if app == "" {
		return models.DeliveryRequest{}, "missing app param"
	}
	if country == "" {
		return models.DeliveryRequest{}, "missing country param"
	}
	if os == "" {
		return models.DeliveryRequest{}, "missing os param"
	}

	return models.DeliveryRequest{
		App:     app,
		Country: strings.ToLower(country),
		OS:      strings.ToLower(os),
	}, ""
}
