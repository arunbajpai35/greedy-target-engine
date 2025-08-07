package delivery

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq" // <-- Add this line

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestMatchesRule(t *testing.T) {
	tests := []struct {
		name     string
		rule     models.TargetingRule
		req      models.DeliveryRequest
		expected bool
	}{
		{
			name: "Match with include country",
			rule: models.TargetingRule{
				IncludeCountry: []string{"us"},
			},
			req: models.DeliveryRequest{
				App:     "any",
				Country: "us",
				OS:      "android",
			},
			expected: true,
		},
		{
			name: "Mismatch with exclude country",
			rule: models.TargetingRule{
				ExcludeCountry: []string{"us"},
			},
			req: models.DeliveryRequest{
				App:     "any",
				Country: "us",
				OS:      "android",
			},
			expected: false,
		},
		{
			name: "Match with include app",
			rule: models.TargetingRule{
				IncludeApp: []string{"com.gametion.ludokinggame"},
			},
			req: models.DeliveryRequest{
				App:     "com.gametion.ludokinggame",
				Country: "in",
				OS:      "android",
			},
			expected: true,
		},
		{
			name: "Mismatch with exclude os",
			rule: models.TargetingRule{
				ExcludeOS: []string{"android"},
			},
			req: models.DeliveryRequest{
				App:     "com.test",
				Country: "in",
				OS:      "android",
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesRule(tc.rule, tc.req)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestHandleDeliveryRequest_MissingParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/delivery?country=us&os=android", nil)
	w := httptest.NewRecorder()

	// Use nil db for this test — it won’t reach DB layer
	handler := HandleDeliveryRequest(nil)
	handler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "missing app param")
}

func TestHandleDeliveryRequest_Match(t *testing.T) {
	// Open real DB (assumes Docker PostgreSQL is running and seeded)
	db, err := sql.Open("postgres", "postgres://chanvimanwani@localhost:5432/targeting_db?sslmode=disable")
	assert.NoError(t, err)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android", nil)
	w := httptest.NewRecorder()

	handler := HandleDeliveryRequest(db)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "spotify")
	assert.Contains(t, w.Body.String(), "subwaysurfer")
}
