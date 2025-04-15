package campaign

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/user/fb-ads/pkg/auth"
	"github.com/user/fb-ads/pkg/models"
)

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

// CreateFromConfig creates a full campaign structure from a configuration file
func (c *CampaignCreator) CreateFromConfig(config *models.CampaignConfig) error {
	// Create the campaign
	campaignID, err := c.CreateCampaign(config)
	if err != nil {
		return fmt.Errorf("error creating campaign: %w", err)
	}

	fmt.Printf("Campaign created with ID: %s\n", campaignID)
	
	// Store adSet IDs to link with ads later
	adSetIDs := make([]string, 0, len(config.AdSets))
	
	// Create ad sets
	for i, adSetConfig := range config.AdSets {
		fmt.Printf("Creating ad set %d/%d: %s\n", i+1, len(config.AdSets), adSetConfig.Name)
		adSetID, err := c.CreateAdSet(campaignID, &adSetConfig)
		if err != nil {
			return fmt.Errorf("error creating ad set: %w", err)
		}
		
		fmt.Printf("Ad set created with ID: %s\n", adSetID)
		adSetIDs = append(adSetIDs, adSetID)
	}
	
	// Create ads (link each ad to an ad set)
	for i, adConfig := range config.Ads {
		// Find the right ad set for this ad
		adSetIndex := i % len(adSetIDs) // Simple distribution - cycle through ad sets
		adSetID := adSetIDs[adSetIndex]
		
		fmt.Printf("Creating ad %d/%d: %s (in ad set: %s)\n", i+1, len(config.Ads), adConfig.Name, adSetID)
		adID, err := c.CreateAd(adSetID, &adConfig)
		if err != nil {
			return fmt.Errorf("error creating ad: %w", err)
		}
		
		fmt.Printf("Ad created with ID: %s\n", adID)
	}
	
	return nil
}

// CreateCampaign creates a new campaign
func (c *CampaignCreator) CreateCampaign(config *models.CampaignConfig) (string, error) {
	params := url.Values{}
	
	// Required parameters
	params.Set("name", config.Name)
	params.Set("objective", config.Objective)
	params.Set("status", getStatusOrDefault(config.Status, "PAUSED")) // Default to PAUSED for safety
	params.Set("buying_type", config.BuyingType)
	params.Set("special_ad_categories", "[]") // Default to empty list
	
	// Budget (convert to cents as required by the API)
	if config.DailyBudget > 0 {
		params.Set("daily_budget", fmt.Sprintf("%d", int64(config.DailyBudget*100)))
	}
	
	if config.LifetimeBudget > 0 {
		params.Set("lifetime_budget", fmt.Sprintf("%d", int64(config.LifetimeBudget*100)))
	}
	
	// Optional parameters
	if config.BidStrategy != "" {
		params.Set("bid_strategy", config.BidStrategy)
	}
	
	if len(config.SpecialAdCategories) > 0 {
		specialCatsJSON, _ := json.Marshal(config.SpecialAdCategories)
		params.Set("special_ad_categories", string(specialCatsJSON))
	}
	
	// Time parameters
	if config.StartTime != "" {
		params.Set("start_time", config.StartTime)
	}
	
	if config.EndTime != "" {
		params.Set("end_time", config.EndTime)
	}
	
	// Create the endpoint
	endpoint := fmt.Sprintf("act_%s/campaigns", c.accountID)
	
	// Make the API request
	return c.createEntity(endpoint, params)
}

// CreateAdSet creates a new ad set
func (c *CampaignCreator) CreateAdSet(campaignID string, config *models.AdSetConfig) (string, error) {
	params := url.Values{}
	
	// Required parameters
	params.Set("name", config.Name)
	params.Set("campaign_id", campaignID)
	params.Set("status", getStatusOrDefault(config.Status, "PAUSED")) // Default to PAUSED for safety
	params.Set("optimization_goal", config.OptimizationGoal)
	params.Set("billing_event", config.BillingEvent)
	
	// Bid amount (convert to cents as required by the API)
	if config.BidAmount > 0 {
		params.Set("bid_amount", fmt.Sprintf("%d", int64(config.BidAmount*100)))
	}
	
	// Targeting
	if len(config.Targeting) > 0 {
		targetingJSON, err := json.Marshal(config.Targeting)
		if err != nil {
			return "", fmt.Errorf("error marshaling targeting: %w", err)
		}
		params.Set("targeting", string(targetingJSON))
	}
	
	// Time parameters
	if config.StartTime != "" {
		params.Set("start_time", config.StartTime)
	}
	
	if config.EndTime != "" {
		params.Set("end_time", config.EndTime)
	}
	
	// Create the endpoint
	endpoint := fmt.Sprintf("act_%s/adsets", c.accountID)
	
	// Make the API request
	return c.createEntity(endpoint, params)
}

// CreateAd creates a new ad
func (c *CampaignCreator) CreateAd(adSetID string, config *models.AdConfig) (string, error) {
	// First, create the creative
	creativeID, err := c.CreateCreative(config.Creative)
	if err != nil {
		return "", fmt.Errorf("error creating creative: %w", err)
	}
	
	params := url.Values{}
	
	// Required parameters
	params.Set("name", config.Name)
	params.Set("adset_id", adSetID)
	params.Set("status", getStatusOrDefault(config.Status, "PAUSED")) // Default to PAUSED for safety
	params.Set("creative", fmt.Sprintf("{\"creative_id\":\"%s\"}", creativeID))
	
	// Create the endpoint
	endpoint := fmt.Sprintf("act_%s/ads", c.accountID)
	
	// Make the API request
	return c.createEntity(endpoint, params)
}

// CreateCreative creates a new creative
func (c *CampaignCreator) CreateCreative(config models.CreativeConfig) (string, error) {
	params := url.Values{}
	
	// Check for required page_id
	if config.PageID == "" {
		return "", fmt.Errorf("page_id is required for creating ad creatives")
	}
	
	// Create object_story_spec with page_id
	objectStorySpec := make(map[string]interface{})
	
	// Add page_id to the story spec
	objectStorySpec["page_id"] = config.PageID
	
	// Create link_data object
	linkData := make(map[string]interface{})
	
	// Validate that LinkURL is not empty, as it's required by the Facebook API
	if config.LinkURL == "" {
		return "", fmt.Errorf("link_url is required for ad creatives and cannot be empty")
	}
	
	linkData["link"] = config.LinkURL
	
	// Note: As per the API error, title is not supported directly in link_data
	// Instead, we'll use name for the title/name field
	titleValue := config.Title
	
	// If Title is empty but Name is set, use the Name field instead
	if titleValue == "" && config.Name != "" {
		titleValue = config.Name
	}
	
	// Set the name parameter for the link data
	if titleValue != "" {
		linkData["name"] = titleValue
	}
	
	if config.Body != "" {
		linkData["message"] = config.Body
	}
	
	// NOTE: ImageURL is no longer supported in link_data of object_story_spec per Facebook API
	// Images should be uploaded separately or referenced by ID
	// This code is commented out to prevent API errors
	/*
	if config.ImageURL != "" {
		linkData["image_url"] = config.ImageURL
	}
	*/
	
	if config.CallToAction != "" {
		callToAction := map[string]string{
			"type": config.CallToAction,
		}
		linkData["call_to_action"] = callToAction
	}
	
	// Add link_data to story spec
	objectStorySpec["link_data"] = linkData
	
	// Marshal the object_story_spec to JSON
	objectJSON, err := json.Marshal(objectStorySpec)
	if err != nil {
		return "", fmt.Errorf("error marshaling creative object: %w", err)
	}
	
	params.Set("object_story_spec", string(objectJSON))
	
	// Create the endpoint
	endpoint := fmt.Sprintf("act_%s/adcreatives", c.accountID)
	
	// Make the API request
	return c.createEntity(endpoint, params)
}

// createEntity is a helper function to create an entity and return its ID
func (c *CampaignCreator) createEntity(endpoint string, params url.Values) (string, error) {
	// Add access token to parameters
	params.Set("access_token", c.auth.AccessToken)
	
	// Build the request URL
	baseURL := fmt.Sprintf("https://graph.facebook.com/%s/%s", c.auth.APIVersion, endpoint)
	
	// Create the POST request
	req, err := http.NewRequest("POST", baseURL, strings.NewReader(params.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	
	// Set the content type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}
	
	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}
	
	// Parse the response
	var result struct {
		ID      string `json:"id"`
		Success bool   `json:"success"`
		Error   struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing response: %w - %s", err, string(body))
	}
	
	// Check for API-level errors
	if result.Error.Message != "" {
		return "", fmt.Errorf("API error: %s (code: %d, type: %s)", 
			result.Error.Message, result.Error.Code, result.Error.Type)
	}
	
	// Return the ID
	return result.ID, nil
}

// getStatusOrDefault returns the status if it's valid, or the default
func getStatusOrDefault(status, defaultStatus string) string {
	if status == "" {
		return defaultStatus
	}
	
	validStatuses := map[string]bool{
		"ACTIVE":    true,
		"PAUSED":    true,
		"DELETED":   true,
		"ARCHIVED":  true,
		"SCHEDULED": true,
	}
	
	upperStatus := strings.ToUpper(status)
	if validStatuses[upperStatus] {
		return upperStatus
	}
	
	return defaultStatus
}