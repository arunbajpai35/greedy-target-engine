package delivery

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/campaigns"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

func HandleDeliveryRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, errMsg := validateParams(r)
		if errMsg != "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
			return
		}

		matched, err := campaigns.GetMatchingCampaigns(db, req.App, req.Country, req.OS)
		if err != nil {
			log.Printf("DB query failed: %v\n", err) // Add this
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			return
		}

		if len(matched) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(matched)
	}
}

// Helpers...
func validateParams(r *http.Request) (models.DeliveryRequest, string) {
	app := r.URL.Query().Get("app")
	country := r.URL.Query().Get("country")
	os := r.URL.Query().Get("os")

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

func matchesRule(rule models.TargetingRule, req models.DeliveryRequest) bool {
	if len(rule.IncludeCountry) > 0 && !contains(rule.IncludeCountry, req.Country) {
		return false
	}
	if len(rule.IncludeOS) > 0 && !contains(rule.IncludeOS, req.OS) {
		return false
	}
	if len(rule.IncludeApp) > 0 && !contains(rule.IncludeApp, req.App) {
		return false
	}

	if len(rule.ExcludeCountry) > 0 && contains(rule.ExcludeCountry, req.Country) {
		return false
	}
	if len(rule.ExcludeOS) > 0 && contains(rule.ExcludeOS, req.OS) {
		return false
	}
	if len(rule.ExcludeApp) > 0 && contains(rule.ExcludeApp, req.App) {
		return false
	}

	return true
}

func contains(list []string, val string) bool {
	for _, v := range list {
		if strings.ToLower(v) == strings.ToLower(val) {
			return true
		}
	}
	return false
}
