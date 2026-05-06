package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/endpoints"
)

type badRequestError struct{ msg string }

func (e badRequestError) Error() string { return e.msg }

func RegisterV2Routes(r chi.Router, eps endpoints.Endpoints) {
	server := kithttp.NewServer(
		eps.Delivery,
		decodeDeliveryRequest,
		encodeDeliveryResponse,
		kithttp.ServerErrorEncoder(encodeError),
	)
	r.Get("/v2/delivery", server.ServeHTTP)
}

func decodeDeliveryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	app := strings.TrimSpace(r.URL.Query().Get("app"))
	country := strings.TrimSpace(r.URL.Query().Get("country"))
	os := strings.TrimSpace(r.URL.Query().Get("os"))

	switch {
	case app == "":
		return nil, badRequestError{"missing app param"}
	case country == "":
		return nil, badRequestError{"missing country param"}
	case os == "":
		return nil, badRequestError{"missing os param"}
	}

	return endpoints.DeliveryRequest{
		App:     strings.ToLower(app),
		Country: strings.ToLower(country),
		OS:      strings.ToLower(os),
	}, nil
}

func encodeDeliveryResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	resp := response.(endpoints.DeliveryResponse)
	if resp.Err != "" {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(map[string]string{"error": resp.Err})
	}
	if len(resp.Campaigns) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	return json.NewEncoder(w).Encode(resp.Campaigns)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	var br badRequestError
	if errors.As(err, &br) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": br.Error()})
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
}
