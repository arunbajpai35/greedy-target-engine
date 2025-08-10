package campaigns

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDBConnStr = "postgres://postgres:password@localhost:5432/targeting_db?sslmode=disable"

func TestGetMatchingCampaigns(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr)
	if err != nil {
		t.Skip("Database not available, skipping campaign store tests")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Cannot connect to database, skipping campaign store tests")
	}

	tests := []struct {
		name     string
		app      string
		country  string
		os       string
		expected []string // expected campaign IDs
	}{
		{
			name:     "Match spotify and subwaysurfer",
			app:      "com.gametion.ludokinggame",
			country:  "us",
			os:       "android",
			expected: []string{"spotify", "subwaysurfer"},
		},
		{
			name:     "Match duolingo only",
			app:      "com.test",
			country:  "germany",
			os:       "android",
			expected: []string{"duolingo"},
		},
		{
			name:     "Match duolingo on iOS",
			app:      "com.test",
			country:  "germany",
			os:       "ios",
			expected: []string{"duolingo"},
		},
		{
			name:     "No matches for web",
			app:      "com.test",
			country:  "us",
			os:       "web",
			expected: []string{},
		},
		{
			name:     "Case insensitive matching",
			app:      "COM.GAMETION.LUDOKINGGAME",
			country:  "US",
			os:       "ANDROID",
			expected: []string{"spotify", "subwaysurfer"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			campaigns, err := GetMatchingCampaigns(db, tc.app, tc.country, tc.os)
			require.NoError(t, err)

			// Extract campaign IDs
			var campaignIDs []string
			for _, c := range campaigns {
				campaignIDs = append(campaignIDs, c.ID)
			}

			assert.ElementsMatch(t, tc.expected, campaignIDs)

			// Verify all returned campaigns are active
			for _, c := range campaigns {
				assert.Equal(t, "ACTIVE", c.Status)
				assert.NotEmpty(t, c.Name)
				assert.NotEmpty(t, c.Img)
				assert.NotEmpty(t, c.CTA)
			}
		})
	}
}

func TestGetCampaignByID(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr)
	if err != nil {
		t.Skip("Database not available, skipping campaign store tests")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Cannot connect to database, skipping campaign store tests")
	}

	tests := []struct {
		name        string
		campaignID  string
		shouldExist bool
	}{
		{
			name:        "Existing campaign",
			campaignID:  "spotify",
			shouldExist: true,
		},
		{
			name:        "Non-existing campaign",
			campaignID:  "nonexistent",
			shouldExist: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			campaign, err := GetCampaignByID(db, tc.campaignID)

			if tc.shouldExist {
				require.NoError(t, err)
				assert.Equal(t, tc.campaignID, campaign.ID)
				assert.NotEmpty(t, campaign.Name)
				assert.NotEmpty(t, campaign.Img)
				assert.NotEmpty(t, campaign.CTA)
				assert.NotEmpty(t, campaign.Status)
			} else {
				assert.Error(t, err)
				assert.Nil(t, campaign)
			}
		})
	}
}

func TestGetAllActiveCampaigns(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr)
	if err != nil {
		t.Skip("Database not available, skipping campaign store tests")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Cannot connect to database, skipping campaign store tests")
	}

	campaigns, err := GetAllActiveCampaigns(db)
	require.NoError(t, err)

	// Should have at least the seeded campaigns
	assert.GreaterOrEqual(t, len(campaigns), 3)

	// All campaigns should be active
	for _, c := range campaigns {
		assert.Equal(t, "ACTIVE", c.Status)
		assert.NotEmpty(t, c.ID)
		assert.NotEmpty(t, c.Name)
		assert.NotEmpty(t, c.Img)
		assert.NotEmpty(t, c.CTA)
	}

	// Check for expected campaign IDs
	campaignIDs := make(map[string]bool)
	for _, c := range campaigns {
		campaignIDs[c.ID] = true
	}

	assert.True(t, campaignIDs["spotify"])
	assert.True(t, campaignIDs["duolingo"])
	assert.True(t, campaignIDs["subwaysurfer"])
}
