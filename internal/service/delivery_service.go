package service

import (
	"context"
	"strings"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/cache"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

type DeliveryService interface {
	Deliver(ctx context.Context, app, country, os string) ([]models.Campaign, error)
}

type deliveryService struct {
	cache *cache.Cache
}

func NewDeliveryService(c *cache.Cache) DeliveryService {
	return &deliveryService{cache: c}
}

func (s *deliveryService) Deliver(_ context.Context, app, country, os string) ([]models.Campaign, error) {
	return s.cache.Match(strings.ToLower(app), strings.ToLower(country), strings.ToLower(os)), nil
}
