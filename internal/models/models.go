package models

type Campaign struct {
	ID     string `json:"cid"`
	Name   string `json:"name"`
	Img    string `json:"img"`
	CTA    string `json:"cta"`
	Status string `json:"status"`
}

type TargetingRule struct {
	CampaignID     string   `json:"campaign_id"`
	IncludeCountry []string `json:"include_country"`
	ExcludeCountry []string `json:"exclude_country"`
	IncludeOS      []string `json:"include_os"`
	ExcludeOS      []string `json:"exclude_os"`
	IncludeApp     []string `json:"include_app"`
	ExcludeApp     []string `json:"exclude_app"`
}

type DeliveryRequest struct {
	App     string `json:"app"`
	Country string `json:"country"`
	OS      string `json:"os"`
}
