package delivery

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test database connection string - adjust as needed
const testDBConnStr = "postgres://postgres:password@localhost:5432/targeting_db?sslmode=disable"

func TestValidateParams(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected models.DeliveryRequest
		hasError bool
		errorMsg string
	}{
		{
			name:     "Valid parameters",
			query:    "?app=com.test&country=us&os=android",
			expected: models.DeliveryRequest{App: "com.test", Country: "us", OS: "android"},
			hasError: false,
		},
		{
			name:     "Missing app parameter",
			query:    "?country=us&os=android",
			hasError: true,
			errorMsg: "missing app param",
		},
		{
			name:     "Missing country parameter",
			query:    "?app=com.test&os=android",
			hasError: true,
			errorMsg: "missing country param",
		},
		{
			name:     "Missing os parameter",
			query:    "?app=com.test&country=us",
			hasError: true,
			errorMsg: "missing os param",
		},
		{
			name:     "Case insensitive parameters",
			query:    "?app=COM.TEST&country=US&os=ANDROID",
			expected: models.DeliveryRequest{App: "COM.TEST", Country: "us", OS: "android"},
			hasError: false,
		},
		{
			name:     "Empty parameters",
			query:    "?app=&country=&os=",
			hasError: true,
			errorMsg: "missing app param",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/delivery"+tc.query, nil)
			result, errMsg := validateParams(req)

			if tc.hasError {
				assert.NotEmpty(t, errMsg)
				assert.Contains(t, errMsg, tc.errorMsg)
			} else {
				assert.Empty(t, errMsg)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestHandleDeliveryRequest_Integration(t *testing.T) {
	// Skip if database is not available
	db, err := sql.Open("postgres", testDBConnStr)
	if err != nil {
		t.Skip("Database not available, skipping integration tests")
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Skip("Cannot connect to database, skipping integration tests")
	}

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedBody   string
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "Successful match - spotify and subwaysurfer",
			query:          "?app=com.gametion.ludokinggame&country=us&os=android",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "spotify")
				assert.Contains(t, body, "subwaysurfer")
				assert.Contains(t, body, "https://somelink")
				assert.Contains(t, body, "https://somelink3")
			},
		},
		{
			name:           "Successful match - duolingo only",
			query:          "?app=com.test&country=germany&os=android",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "duolingo")
				assert.NotContains(t, body, "spotify")
				assert.NotContains(t, body, "subwaysurfer")
			},
		},
		{
			name:           "No matches - should return 204",
			query:          "?app=com.test&country=germany&os=web",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Missing app parameter",
			query:          "?country=us&os=android",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"missing app param"}`,
		},
		{
			name:           "Missing country parameter",
			query:          "?app=com.test&os=android",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"missing country param"}`,
		},
		{
			name:           "Missing os parameter",
			query:          "?app=com.test&country=us",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"missing os param"}`,
		},
		{
			name:           "Case insensitive matching",
			query:          "?app=COM.GAMETION.LUDOKINGGAME&country=US&os=ANDROID",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				assert.Contains(t, body, "spotify")
				assert.Contains(t, body, "subwaysurfer")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/delivery"+tc.query, nil)
			w := httptest.NewRecorder()

			handler := HandleDeliveryRequest(db)
			handler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, w.Body.String())
			}

			if tc.checkResponse != nil {
				tc.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestHandleDeliveryRequest_ResponseFormat(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr)
	if err != nil {
		t.Skip("Database not available, skipping response format tests")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Cannot connect to database, skipping response format tests")
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android", nil)
	w := httptest.NewRecorder()

	handler := HandleDeliveryRequest(db)
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response to verify structure
	var campaigns []models.Campaign
	err = json.Unmarshal(w.Body.Bytes(), &campaigns)
	require.NoError(t, err)

	// Verify campaign structure
	for _, campaign := range campaigns {
		assert.NotEmpty(t, campaign.ID)
		assert.NotEmpty(t, campaign.Name)
		assert.NotEmpty(t, campaign.Img)
		assert.NotEmpty(t, campaign.CTA)
		assert.Equal(t, "ACTIVE", campaign.Status)
	}
}

func TestHandleDeliveryRequest_Performance(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr)
	if err != nil {
		t.Skip("Database not available, skipping performance tests")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Cannot connect to database, skipping performance tests")
	}

	// Test multiple concurrent requests
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/v1/delivery?app=com.gametion.ludokinggame&country=us&os=android", nil)
			w := httptest.NewRecorder()

			handler := HandleDeliveryRequest(db)
			handler(w, req)

			results <- w.Code
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		statusCode := <-results
		if statusCode == http.StatusOK {
			successCount++
		}
	}

	// All requests should succeed
	assert.Equal(t, numRequests, successCount)
}
