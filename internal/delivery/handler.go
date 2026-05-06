package delivery

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/campaigns"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/metrics"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

func HandleDeliveryRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := middleware.GetReqID(r.Context())
		w.Header().Set("Content-Type", "application/json")

		req, errMsg := validateParams(r)
		if errMsg != "" {
			w.WriteHeader(http.StatusBadRequest)
			metrics.ObserveRequest("bad_request", time.Since(start).Seconds())
			_ = json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
			return
		}

		matched, err := campaigns.GetMatchingCampaigns(r.Context(), db, req.App, req.Country, req.OS)
		if err != nil {
			slog.ErrorContext(r.Context(), "delivery query failed", "req_id", reqID, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			metrics.ObserveRequest("error", time.Since(start).Seconds())
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			return
		}

		slog.InfoContext(r.Context(), "delivery",
			"req_id", reqID,
			"app", req.App, "country", req.Country, "os", req.OS,
			"matches", len(matched), "duration_ms", time.Since(start).Milliseconds(),
		)

		if len(matched) == 0 {
			w.WriteHeader(http.StatusNoContent)
			metrics.ObserveRequest("no_content", time.Since(start).Seconds())
			return
		}

		w.WriteHeader(http.StatusOK)
		metrics.ObserveRequest("ok", time.Since(start).Seconds())
		if err := json.NewEncoder(w).Encode(matched); err != nil {
			slog.ErrorContext(r.Context(), "encode response", "req_id", reqID, "err", err)
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
		App:     strings.ToLower(app),
		Country: strings.ToLower(country),
		OS:      strings.ToLower(os),
	}, ""
}
