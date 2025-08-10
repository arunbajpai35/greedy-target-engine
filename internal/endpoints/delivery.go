package endpoints

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/metrics"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/service"
)

// Request and Response models for the endpoint
 type DeliveryRequest struct {
	App     string `json:"app"`
	Country string `json:"country"`
	OS      string `json:"os"`
 }

 type DeliveryResponse struct {
	Campaigns []models.Campaign `json:"campaigns,omitempty"`
	Err       string            `json:"error,omitempty"`
 }

 type Endpoints struct {
	Delivery endpoint.Endpoint
 }

 func MakeDeliveryEndpoint(svc service.DeliveryService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		start := time.Now()
		req := request.(DeliveryRequest)
		campaigns, err := svc.Deliver(req.App, req.Country, req.OS)
		status := "ok"
		if err != nil {
			status = "error"
		}
		metrics.ObserveRequest(status, time.Since(start).Seconds())

		if err != nil {
			return DeliveryResponse{Err: "internal server error"}, nil
		}
		if len(campaigns) == 0 {
			return DeliveryResponse{Campaigns: []models.Campaign{}}, nil
		}
		return DeliveryResponse{Campaigns: campaigns}, nil
	}
 }