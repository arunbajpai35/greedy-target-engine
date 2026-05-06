package service

import (
	"context"
	"database/sql"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/campaigns"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

type DeliveryService interface {
	Deliver(ctx context.Context, app, country, os string) ([]models.Campaign, error)
}

type deliveryService struct {
	db *sql.DB
}

func NewDeliveryService(db *sql.DB) DeliveryService {
	return &deliveryService{db: db}
}

func (s *deliveryService) Deliver(ctx context.Context, app, country, os string) ([]models.Campaign, error) {
	return campaigns.GetMatchingCampaigns(ctx, s.db, app, country, os)
}
