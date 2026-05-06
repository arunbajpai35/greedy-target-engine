package endpoints

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

type fakeService struct {
	out []models.Campaign
	err error
}

func (f fakeService) Deliver(_ context.Context, _, _, _ string) ([]models.Campaign, error) {
	return f.out, f.err
}

func TestDeliveryEndpoint_OK(t *testing.T) {
	svc := fakeService{out: []models.Campaign{{ID: "spotify", Name: "Spotify"}}}
	resp, err := MakeDeliveryEndpoint(svc)(context.Background(), DeliveryRequest{App: "x", Country: "us", OS: "android"})
	assert.NoError(t, err)
	r := resp.(DeliveryResponse)
	assert.Empty(t, r.Err)
	assert.Len(t, r.Campaigns, 1)
	assert.Equal(t, "spotify", r.Campaigns[0].ID)
}

func TestDeliveryEndpoint_Empty(t *testing.T) {
	svc := fakeService{out: nil}
	resp, _ := MakeDeliveryEndpoint(svc)(context.Background(), DeliveryRequest{})
	r := resp.(DeliveryResponse)
	assert.Empty(t, r.Err)
	assert.Empty(t, r.Campaigns)
}

func TestDeliveryEndpoint_Error(t *testing.T) {
	svc := fakeService{err: errors.New("boom")}
	resp, err := MakeDeliveryEndpoint(svc)(context.Background(), DeliveryRequest{})
	assert.NoError(t, err, "endpoint must not surface internal errors as transport errors")
	r := resp.(DeliveryResponse)
	assert.Equal(t, "internal server error", r.Err)
}
