package campaign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/user/fb-ads/pkg/auth"
)

// CampaignConfig represents a full campaign configuration including ad sets and ads
type CampaignConfig struct {
	Name                string       `json:"name"`
	Status              string       `json:"status"`
	Objective           string       `json:"objective"`
	BuyingType          string       `json:"buying_type"`
	SpecialAdCategories []string     `json:"special_ad_categories"`
	BidStrategy         string       `json:"bid_strategy"`
	DailyBudget         float64      `json:"daily_budget"`
	LifetimeBudget      float64      `json:"lifetime_budget,omitempty"`
	AdSets              []AdSetConfig `json:"adsets"`
	Ads                 []AdConfig    `json:"ads"`
}

// AdSetConfig represents configuration for an ad set
type AdSetConfig struct {
	Name             string                 `json:"name"`
	Status           string                 `json:"status,omitempty"`
	Targeting        map[string]interface{} `json:"targeting"`
	OptimizationGoal string                 `json:"optimization_goal"`
	BillingEvent     string                 `json:"billing_event"`
	BidAmount        float64                `json:"bid_amount"`
	StartTime        string                 `json:"start_time"`
	EndTime          string                 `json:"end_time,omitempty"`
}

// AdConfig represents configuration for an ad
type AdConfig struct {
	Name     string         `json:"name"`
	Status   string         `json:"status,omitempty"`
	Creative CreativeConfig `json:"creative"`
}

// CreativeConfig represents configuration for an ad creative
type CreativeConfig struct {
	Title        string `json:"title"`
	Body         string `json:"body"`
	ImageURL     string `json:"image_url"`
	LinkURL      string `json:"link_url"`
	CallToAction string `json:"call_to_action"`
}

// CampaignCreator handles creation of campaigns
type CampaignCreator struct {
	httpClient *http.Client
	auth       *auth.FacebookAuth
	accountID  string
}

// NewCampaignCreator creates a new campaign creator
func NewCampaignCreator(auth *auth.FacebookAuth, accountID string) *CampaignCreator {
	return &CampaignCreator{
		httpClient: &http.Client{},
		auth:       auth,
		accountID:  accountID,
	}
}

// CreateCampaign creates a campaign from a configuration
func (c *CampaignCreator) CreateCampaign(config *CampaignConfig) (string, error) {
	// Convert daily budget to cents as required by the API
	dailyBudget := int64(config.DailyBudget * 100)
	
	params := url.Values{}
	params.Set("name", config.Name)
	params.Set("status", config.Status)
	params.Set("objective", config.Objective)
	params.Set("buying_type", config.BuyingType)
	params.Set("bid_strategy", config.BidStrategy)
	params.Set("daily_budget", fmt.Sprintf("%d", dailyBudget))
	
	if len(config.SpecialAdCategories) > 0 {
		specialCats, _ := json.Marshal(config.SpecialAdCategories)
		params.Set("special_ad_categories", string(specialCats))
	}
	
	endpoint := fmt.Sprintf("act_%s/campaigns", c.accountID)
	
	req, err := c.createPostRequest(endpoint, params)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}
	
	var result struct {
		ID string `json:"id"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}
	
	return result.ID, nil
}

// CreateAdSet creates an ad set for a campaign
func (c *CampaignCreator) CreateAdSet(campaignID string, config *AdSetConfig) (string, error) {
	// TODO: Implement ad set creation
	return "", nil
}

// CreateAd creates an ad in an ad set
func (c *CampaignCreator) CreateAd(adSetID string, config *AdConfig) (string, error) {
	// TODO: Implement ad creation
	return "", nil
}

// CreateFromConfig creates a full campaign structure from a configuration file
func (c *CampaignCreator) CreateFromConfig(config *CampaignConfig) error {
	campaignID, err := c.CreateCampaign(config)
	if err != nil {
		return fmt.Errorf("error creating campaign: %w", err)
	}
	
	// Store adSet IDs to link with ads later
	adSetIDs := make([]string, 0, len(config.AdSets))
	
	// Create ad sets
	for _, adSetConfig := range config.AdSets {
		adSetID, err := c.CreateAdSet(campaignID, &adSetConfig)
		if err != nil {
			return fmt.Errorf("error creating ad set: %w", err)
		}
		adSetIDs = append(adSetIDs, adSetID)
	}
	
	// Create ads (simplified for now - assuming one ad per ad set)
	for i, adConfig := range config.Ads {
		if i < len(adSetIDs) {
			_, err := c.CreateAd(adSetIDs[i], &adConfig)
			if err != nil {
				return fmt.Errorf("error creating ad: %w", err)
			}
		}
	}
	
	return nil
}

// createPostRequest creates a POST request with authentication
func (c *CampaignCreator) createPostRequest(endpoint string, params url.Values) (*http.Request, error) {
	baseURL := fmt.Sprintf("https://graph.facebook.com/%s/%s", c.auth.APIVersion, endpoint)
	
	if params == nil {
		params = url.Values{}
	}
	
	params.Set("access_token", c.auth.AccessToken)
	
	req, err := http.NewRequest("POST", baseURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}