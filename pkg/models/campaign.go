package models

import (
	"time"
)

// Campaign represents a Facebook ad campaign
type Campaign struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Status               string    `json:"status"`
	ObjectiveType        string    `json:"objective_type"`
	SpendCap             float64   `json:"spend_cap,omitempty"`
	DailyBudget          float64   `json:"daily_budget,omitempty"`
	LifetimeBudget       float64   `json:"lifetime_budget,omitempty"`
	BidStrategy          string    `json:"bid_strategy,omitempty"`
	BuyingType           string    `json:"buying_type"`
	Created              time.Time `json:"created_time"`
	Updated              time.Time `json:"updated_time"`
	StartTime            time.Time `json:"start_time,omitempty"`
	StopTime             time.Time `json:"stop_time,omitempty"`
	SpecialAdCategories  []string  `json:"special_ad_categories,omitempty"`
	
	// Raw time strings for parsing flexibility
	CreatedTimeString    string    `json:"created_time_string,omitempty"`
	UpdatedTimeString    string    `json:"updated_time_string,omitempty"`
	StartTimeString      string    `json:"start_time_string,omitempty"`
	StopTimeString       string    `json:"stop_time_string,omitempty"`
}

// CampaignResponse represents the Facebook API response for campaigns
type CampaignResponse struct {
	Data    []Campaign `json:"data"`
	Paging  Paging     `json:"paging"`
	Summary Summary    `json:"summary,omitempty"`
}

// Paging represents pagination information from Facebook API responses
type Paging struct {
	Cursors Cursors `json:"cursors"`
	Next    string  `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
}

// Cursors represents pagination cursors
type Cursors struct {
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
}

// Summary represents summary information from Facebook API responses
type Summary struct {
	TotalCount int `json:"total_count"`
}