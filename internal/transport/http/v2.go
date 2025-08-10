package httptransport

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/endpoints"
)

func RegisterV2Routes(r chi.Router, eps endpoints.Endpoints) {
	server := kithttp.NewServer(
		eps.Delivery,
		decodeDeliveryRequest,
		encodeDeliveryResponse,
	)

	r.Get("/v2/delivery", server.ServeHTTP)
}

func decodeDeliveryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	app := strings.TrimSpace(r.URL.Query().Get("app"))
	country := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("country")))
	os := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("os")))
	return endpoints.DeliveryRequest{App: app, Country: country, OS: os}, nil
}

func encodeDeliveryResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	resp := response.(endpoints.DeliveryResponse)
	// On empty campaigns, align with v1 behavior and return 204
	if resp.Err == "" && len(resp.Campaigns) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	if resp.Err != "" {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(map[string]string{"error": resp.Err})
	}
	return json.NewEncoder(w).Encode(resp.Campaigns)
}