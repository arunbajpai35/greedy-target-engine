package service

import (
	"database/sql"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/campaigns"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

// DeliveryService defines the business logic for campaign delivery
 type DeliveryService interface {
	Deliver(app, country, os string) ([]models.Campaign, error)
 }

 type deliveryService struct {
	db *sql.DB
 }

 func NewDeliveryService(db *sql.DB) DeliveryService {
	return &deliveryService{db: db}
 }

 func (s *deliveryService) Deliver(app, country, os string) ([]models.Campaign, error) {
	return campaigns.GetMatchingCampaigns(s.db, app, country, os)
 }