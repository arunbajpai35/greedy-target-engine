package campaigns

import (
	"database/sql"

	"github.com/arunbajpai35/greedygame-targeting-engine/internal/models"
)

func GetMatchingCampaigns(db *sql.DB, app, country, os string) ([]models.Campaign, error) {
	query := `
	SELECT c.cid, c.name, c.img, c.cta, c.status
	FROM campaigns c
	JOIN targeting_rules tr ON c.cid = tr.cid
	WHERE c.status = 'ACTIVE'
	  AND tr.app = $1
	  AND tr.country = $2
	  AND tr.os = $3
	`

	rows, err := db.Query(query, app, country, os)
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

	return campaigns, nil
}
