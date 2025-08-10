package campaigns

import (
	"database/sql"
	"strings"
	"time"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/metrics"
	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

func GetMatchingCampaigns(db *sql.DB, app, country, os string) ([]models.Campaign, error) {
	// Convert to lowercase for case-insensitive matching
	app = strings.ToLower(app)
	country = strings.ToLower(country)
	os = strings.ToLower(os)

	query := `
	SELECT DISTINCT c.cid, c.name, c.img, c.cta, c.status
	FROM campaigns c
	JOIN targeting_rules tr ON c.cid = tr.cid
	WHERE c.status = 'ACTIVE'
	  AND (
		-- Check include rules
		(tr.include_country IS NULL OR $2 = ANY(tr.include_country))
		AND (tr.include_os IS NULL OR $3 = ANY(tr.include_os))
		AND (tr.include_app IS NULL OR $1 = ANY(tr.include_app))
		-- Check exclude rules
		AND (tr.exclude_country IS NULL OR NOT ($2 = ANY(tr.exclude_country)))
		AND (tr.exclude_os IS NULL OR NOT ($3 = ANY(tr.exclude_os)))
		AND (tr.exclude_app IS NULL OR NOT ($1 = ANY(tr.exclude_app)))
	  )
	ORDER BY c.cid
	`

	start := time.Now()
	rows, err := db.Query(query, app, country, os)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	metrics.ObserveDBQuery(time.Since(start).Seconds())

	var campaigns []models.Campaign

	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(&c.ID, &c.Name, &c.Img, &c.CTA, &c.Status); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return campaigns, nil
}

// GetCampaignByID retrieves a single campaign by ID
func GetCampaignByID(db *sql.DB, campaignID string) (*models.Campaign, error) {
	query := `SELECT cid, name, img, cta, status FROM campaigns WHERE cid = $1`

	var c models.Campaign
	err := db.QueryRow(query, campaignID).Scan(&c.ID, &c.Name, &c.Img, &c.CTA, &c.Status)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// GetAllActiveCampaigns retrieves all active campaigns
func GetAllActiveCampaigns(db *sql.DB) ([]models.Campaign, error) {
	query := `SELECT cid, name, img, cta, status FROM campaigns WHERE status = 'ACTIVE' ORDER BY cid`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(&c.ID, &c.Name, &c.Img, &c.CTA, &c.Status); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return campaigns, nil
}
