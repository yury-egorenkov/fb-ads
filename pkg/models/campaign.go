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

// CampaignDetails represents detailed information about a campaign
type CampaignDetails struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Status              string                 `json:"status"`
	ObjectiveType       string                 `json:"objective_type"`
	SpendCap            float64                `json:"spend_cap,omitempty"`
	DailyBudget         float64                `json:"daily_budget,omitempty"`
	LifetimeBudget      float64                `json:"lifetime_budget,omitempty"`
	BidStrategy         string                 `json:"bid_strategy,omitempty"`
	BuyingType          string                 `json:"buying_type"`
	Created             time.Time              `json:"created_time"`
	Updated             time.Time              `json:"updated_time"`
	StartTime           time.Time              `json:"start_time,omitempty"`
	StopTime            time.Time              `json:"stop_time,omitempty"`
	SpecialAdCategories []string               `json:"special_ad_categories,omitempty"`
	Targeting           map[string]interface{} `json:"targeting,omitempty"`
	AdSets              []AdSetDetails         `json:"adsets,omitempty"`
	Ads                 []AdDetails            `json:"ads,omitempty"`
}

// AdSetDetails represents detailed information about an ad set
type AdSetDetails struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Status           string                 `json:"status"`
	OptimizationGoal string                 `json:"optimization_goal"`
	BillingEvent     string                 `json:"billing_event"`
	BidAmount        float64                `json:"bid_amount"`
	StartTime        time.Time              `json:"start_time,omitempty"`
	EndTime          time.Time              `json:"end_time,omitempty"`
	Targeting        map[string]interface{} `json:"targeting,omitempty"`
}

// AdDetails represents detailed information about an ad
type AdDetails struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Status   string          `json:"status"`
	Creative CreativeDetails `json:"creative,omitempty"`
}

// CreativeDetails represents detailed information about an ad creative
type CreativeDetails struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Title            string `json:"title,omitempty"`
	Body             string `json:"body,omitempty"`
	ImageURL         string `json:"image_url,omitempty"`
	LinkURL          string `json:"link_url,omitempty"`
	CallToActionType string `json:"call_to_action_type,omitempty"`
	PageID           string `json:"page_id,omitempty"`
}

// CampaignConfig represents a campaign configuration for creating or exporting campaigns
type CampaignConfig struct {
	Name                string          `json:"name"`
	Status              string          `json:"status"`
	Objective           string          `json:"objective"`
	BuyingType          string          `json:"buying_type"`
	SpecialAdCategories []string        `json:"special_ad_categories,omitempty"`
	BidStrategy         string          `json:"bid_strategy"`
	DailyBudget         float64         `json:"daily_budget,omitempty"`
	LifetimeBudget      float64         `json:"lifetime_budget,omitempty"`
	StartTime           string          `json:"start_time,omitempty"`
	EndTime             string          `json:"end_time,omitempty"`
	AdSets              []AdSetConfig   `json:"adsets"`
	Ads                 []AdConfig      `json:"ads"`
}

// AdSetConfig represents configuration for an ad set
type AdSetConfig struct {
	Name             string                 `json:"name"`
	Status           string                 `json:"status,omitempty"`
	Targeting        map[string]interface{} `json:"targeting"`
	OptimizationGoal string                 `json:"optimization_goal"`
	BillingEvent     string                 `json:"billing_event"`
	BidAmount        float64                `json:"bid_amount"`
	StartTime        string                 `json:"start_time,omitempty"`
	EndTime          string                 `json:"end_time,omitempty"`
}

// AdConfig represents configuration for an ad
type AdConfig struct {
	Name     string          `json:"name"`
	Status   string          `json:"status,omitempty"`
	Creative CreativeConfig  `json:"creative"`
}

// CreativeConfig represents configuration for an ad creative
type CreativeConfig struct {
	Title            string `json:"title,omitempty"`
	Name             string `json:"name,omitempty"`  // Added to support templates using name instead of title
	Body             string `json:"body,omitempty"`
	ImageURL         string `json:"image_url,omitempty"`
	LinkURL          string `json:"link_url,omitempty"`
	CallToAction     string `json:"call_to_action,omitempty"`
	PageID           string `json:"page_id"`
}

// Page represents a Facebook Page
type Page struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Picture  struct {
		Data struct {
			Height       int    `json:"height"`
			Width        int    `json:"width"`
			IsSilhouette bool   `json:"is_silhouette"`
			URL          string `json:"url"`
		} `json:"data"`
	} `json:"picture,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}