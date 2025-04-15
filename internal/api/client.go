package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/user/fb-ads/pkg/auth"
	"github.com/user/fb-ads/pkg/models"
)

// Client is the Facebook Marketing API client
type Client struct {
	httpClient *http.Client
	auth       *auth.FacebookAuth
	accountID  string
}

// NewClient creates a new Facebook Marketing API client
func NewClient(auth *auth.FacebookAuth, accountID string) *Client {
	return &Client{
		httpClient: &http.Client{},
		auth:       auth,
		accountID:  accountID,
	}
}

// GetCampaigns retrieves all campaigns for the account
func (c *Client) GetCampaigns(limit int, after string) (*models.CampaignResponse, error) {
	params := url.Values{}
	params.Set("fields", "id,name,status,objective,spend_cap,daily_budget,lifetime_budget,bid_strategy,buying_type,created_time,updated_time,start_time,stop_time,special_ad_categories")

	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	if after != "" {
		params.Set("after", after)
	}

	endpoint := fmt.Sprintf("act_%s/campaigns", c.accountID)

	req, err := c.auth.GetAuthenticatedRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// First, decode raw response to handle date parsing issues
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Debugging - print raw response
	// fmt.Println("Raw API response:", string(body))

	// Create a map to hold the raw JSON
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling raw response: %w", err)
	}

	// Create the campaign response
	campaignResp := models.CampaignResponse{}

	// Process the data array if it exists
	if rawData, ok := rawResponse["data"].([]interface{}); ok {
		for _, rawCampaign := range rawData {
			campaignMap, ok := rawCampaign.(map[string]interface{})
			if !ok {
				continue
			}

			campaign := models.Campaign{
				ID:             getString(campaignMap, "id"),
				Name:           getString(campaignMap, "name"),
				Status:         getString(campaignMap, "status"),
				ObjectiveType:  getString(campaignMap, "objective"),
				SpendCap:       getFloat(campaignMap, "spend_cap"),
				DailyBudget:    getFloat(campaignMap, "daily_budget"),
				LifetimeBudget: getFloat(campaignMap, "lifetime_budget"),
				BidStrategy:    getString(campaignMap, "bid_strategy"),
				BuyingType:     getString(campaignMap, "buying_type"),
			}

			// Handle date fields with flexible parsing
			createdStr := getString(campaignMap, "created_time")
			if createdStr != "" {
				campaign.Created = parseTime(createdStr)
			}

			updatedStr := getString(campaignMap, "updated_time")
			if updatedStr != "" {
				campaign.Updated = parseTime(updatedStr)
			}

			startStr := getString(campaignMap, "start_time")
			if startStr != "" {
				campaign.StartTime = parseTime(startStr)
			}

			stopStr := getString(campaignMap, "stop_time")
			if stopStr != "" {
				campaign.StopTime = parseTime(stopStr)
			}

			// Parse special_ad_categories if it exists
			if rawCategories, ok := campaignMap["special_ad_categories"].([]interface{}); ok {
				for _, cat := range rawCategories {
					if catStr, ok := cat.(string); ok {
						campaign.SpecialAdCategories = append(campaign.SpecialAdCategories, catStr)
					}
				}
			}

			campaignResp.Data = append(campaignResp.Data, campaign)
		}
	}

	// Process paging info if it exists
	if rawPaging, ok := rawResponse["paging"].(map[string]interface{}); ok {
		if rawCursors, ok := rawPaging["cursors"].(map[string]interface{}); ok {
			campaignResp.Paging.Cursors.Before = getString(rawCursors, "before")
			campaignResp.Paging.Cursors.After = getString(rawCursors, "after")
		}
		campaignResp.Paging.Next = getString(rawPaging, "next")
		campaignResp.Paging.Previous = getString(rawPaging, "previous")
	}

	return &campaignResp, nil
}

// Helper functions for parsing the JSON response
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	return 0
}

func parseTime(timeStr string) time.Time {
	// Try multiple date formats
	formats := []string{
		time.RFC3339,                // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05-0700",  // Another common format
		"2006-01-02T15:04:05",       // Without timezone
		"2006-01-02T15:04:05-07:00", // With timezone
		"2006-01-02T15:04:05+0000",  // Yet another format
		"2006-01-02",                // Just date
		time.RFC1123,                // Mon, 02 Jan 2006 15:04:05 MST
		time.RFC1123Z,               // Mon, 02 Jan 2006 15:04:05 -0700
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}

	// If all parse attempts fail, try a custom approach
	// Handle the format like "2025-04-08T12:02:56+0100"
	if len(timeStr) > 20 {
		// Extract the timezone portion
		tzOffset := timeStr[len(timeStr)-5:]
		if len(tzOffset) == 5 && (tzOffset[0] == '+' || tzOffset[0] == '-') {
			// Convert +0100 to +01:00
			tzFormatted := tzOffset[:3] + ":" + tzOffset[3:]
			reformatted := timeStr[:len(timeStr)-5] + tzFormatted
			if t, err := time.Parse(time.RFC3339, reformatted); err == nil {
				return t
			}
		}
	}

	fmt.Printf("Warning: could not parse time string: %s\n", timeStr)
	return time.Time{} // Return zero time if parsing fails
}

// GetCampaignDetails retrieves detailed information about a specific campaign
func (c *Client) GetCampaignDetails(campaignID string) (*models.CampaignDetails, error) {
	// Create the fields list for all the information we need
	fields := []string{
		"id",
		"name",
		"status",
		"objective",
		"spend_cap",
		"daily_budget",
		"lifetime_budget",
		"bid_strategy",
		"buying_type",
		"created_time",
		"updated_time",
		"start_time",
		"stop_time",
		"special_ad_categories",
		// "targeting",  // Targeting is at the adset level, not campaign level
		"adlabels",
		"promoted_object",
		"source_campaign_id",
		"adsets{id,name,status,targeting,optimization_goal,billing_event,bid_amount,start_time,end_time}",
		"ads{id,name,status,creative{id,name,title,body,image_url,link_url,call_to_action_type,object_story_spec{page_id}}}",
	}

	// Create the parameters
	params := url.Values{}
	params.Set("fields", strings.Join(fields, ","))

	// Create the endpoint
	endpoint := campaignID

	// Create the request
	req, err := c.auth.GetAuthenticatedRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// For debugging
	// fmt.Println("Raw response:", string(body))

	// Parse the raw JSON response
	var rawData map[string]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Extract campaign details
	details := &models.CampaignDetails{
		ID:                  getString(rawData, "id"),
		Name:                getString(rawData, "name"),
		Status:              getString(rawData, "status"),
		ObjectiveType:       getString(rawData, "objective"),
		SpendCap:            getFloat(rawData, "spend_cap"),
		DailyBudget:         getFloat(rawData, "daily_budget"),
		LifetimeBudget:      getFloat(rawData, "lifetime_budget"),
		BidStrategy:         getString(rawData, "bid_strategy"),
		BuyingType:          getString(rawData, "buying_type"),
		SpecialAdCategories: []string{},
	}

	// Handle date fields
	createdStr := getString(rawData, "created_time")
	if createdStr != "" {
		details.Created = parseTime(createdStr)
	}

	updatedStr := getString(rawData, "updated_time")
	if updatedStr != "" {
		details.Updated = parseTime(updatedStr)
	}

	startStr := getString(rawData, "start_time")
	if startStr != "" {
		details.StartTime = parseTime(startStr)
	}

	stopStr := getString(rawData, "stop_time")
	if stopStr != "" {
		details.StopTime = parseTime(stopStr)
	}

	// Handle special ad categories
	if categories, ok := rawData["special_ad_categories"].([]interface{}); ok {
		for _, cat := range categories {
			if catStr, ok := cat.(string); ok {
				details.SpecialAdCategories = append(details.SpecialAdCategories, catStr)
			}
		}
	}

	// Extract targeting if available
	if targeting, ok := rawData["targeting"].(map[string]interface{}); ok {
		details.Targeting = targeting
	}

	// Extract adsets if available
	if adsets, ok := rawData["adsets"].(map[string]interface{}); ok {
		if data, ok := adsets["data"].([]interface{}); ok {
			for _, rawAdset := range data {
				if adsetMap, ok := rawAdset.(map[string]interface{}); ok {
					adset := models.AdSetDetails{
						ID:               getString(adsetMap, "id"),
						Name:             getString(adsetMap, "name"),
						Status:           getString(adsetMap, "status"),
						OptimizationGoal: getString(adsetMap, "optimization_goal"),
						BillingEvent:     getString(adsetMap, "billing_event"),
						BidAmount:        getFloat(adsetMap, "bid_amount"),
					}

					// Parse dates
					startStr := getString(adsetMap, "start_time")
					if startStr != "" {
						adset.StartTime = parseTime(startStr)
					}

					endStr := getString(adsetMap, "end_time")
					if endStr != "" {
						adset.EndTime = parseTime(endStr)
					}

					// Extract targeting if available
					if targeting, ok := adsetMap["targeting"].(map[string]interface{}); ok {
						adset.Targeting = targeting
					}

					details.AdSets = append(details.AdSets, adset)
				}
			}
		}
	}

	// Extract ads if available
	if ads, ok := rawData["ads"].(map[string]interface{}); ok {
		if data, ok := ads["data"].([]interface{}); ok {
			for _, rawAd := range data {
				if adMap, ok := rawAd.(map[string]interface{}); ok {
					ad := models.AdDetails{
						ID:     getString(adMap, "id"),
						Name:   getString(adMap, "name"),
						Status: getString(adMap, "status"),
					}

					// Extract creative if available
					if creative, ok := adMap["creative"].(map[string]interface{}); ok {
						creativeDetails := models.CreativeDetails{
							ID:               getString(creative, "id"),
							Name:             getString(creative, "name"),
							Title:            getString(creative, "title"),
							Body:             getString(creative, "body"),
							ImageURL:         getString(creative, "image_url"),
							LinkURL:          getString(creative, "link_url"),
							CallToActionType: getString(creative, "call_to_action_type"),
						}

						// Extract page_id from object_story_spec if available
						if objectStorySpec, ok := creative["object_story_spec"].(map[string]interface{}); ok {
							creativeDetails.PageID = getString(objectStorySpec, "page_id")
						}

						ad.Creative = creativeDetails
					}

					details.Ads = append(details.Ads, ad)
				}
			}
		}
	}

	return details, nil
}

// GetAllCampaigns retrieves all campaigns by handling pagination
func (c *Client) GetAllCampaigns() ([]models.Campaign, error) {
	// Check if we're in mock mode (no API credentials)
	// This is helpful for testing without real Facebook credentials
	if c.auth.AccessToken == "YOUR_FACEBOOK_ACCESS_TOKEN" || c.auth.AccessToken == "" {
		fmt.Println("[Using mock data] Configure real Facebook credentials with 'fbads config'")
		return getMockCampaigns(), nil
	}

	fmt.Println("[Using Facebook API] Fetching campaigns from account ID:", c.accountID)

	var allCampaigns []models.Campaign
	var nextCursor string

	for {
		resp, err := c.GetCampaigns(100, nextCursor)
		if err != nil {
			return nil, err
		}

		allCampaigns = append(allCampaigns, resp.Data...)
		fmt.Printf("[Using Facebook API] Retrieved %d campaigns\n", len(resp.Data))

		// Check if there are more pages
		if resp.Paging.Next == "" {
			break
		}

		// Extract the next cursor
		nextCursor = resp.Paging.Cursors.After
		if nextCursor == "" {
			break
		}
	}

	return allCampaigns, nil
}

// GetPages retrieves Facebook Pages available for the current access token
func (c *Client) GetPages() ([]models.Page, error) {
	// Create the parameters
	params := url.Values{}
	params.Set("fields", "id,name,category,picture")

	// Create the endpoint (no account ID needed as we're getting pages for the user token)
	endpoint := "me/accounts"

	// Create the request
	req, err := c.auth.GetAuthenticatedRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse the response
	var result struct {
		Data   []models.Page `json:"data"`
		Paging struct {
			Cursors struct {
				Before string `json:"before"`
				After  string `json:"after"`
			} `json:"cursors"`
			Next string `json:"next"`
		} `json:"paging"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return result.Data, nil
}

// getMockCampaigns returns mock campaign data for testing
func getMockCampaigns() []models.Campaign {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	return []models.Campaign{
		{
			ID:             "23847239847",
			Name:           "Summer Sale 2023",
			Status:         "ACTIVE",
			ObjectiveType:  "CONVERSIONS",
			SpendCap:       0,
			DailyBudget:    50.00,
			LifetimeBudget: 0,
			BidStrategy:    "LOWEST_COST_WITHOUT_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, 0, -5),
			Updated:        yesterday,
		},
		{
			ID:             "23847239848",
			Name:           "New Product Launch - Premium Widgets",
			Status:         "ACTIVE",
			ObjectiveType:  "CONVERSIONS",
			SpendCap:       1000.00,
			DailyBudget:    100.00,
			LifetimeBudget: 0,
			BidStrategy:    "LOWEST_COST_WITHOUT_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, 0, -10),
			Updated:        yesterday,
		},
		{
			ID:             "23847239849",
			Name:           "Brand Awareness Campaign",
			Status:         "PAUSED",
			ObjectiveType:  "BRAND_AWARENESS",
			SpendCap:       0,
			DailyBudget:    0,
			LifetimeBudget: 5000.00,
			BidStrategy:    "LOWEST_COST_WITH_BID_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, -1, 0),
			Updated:        yesterday.AddDate(0, 0, -5),
		},
		{
			ID:             "23847239850",
			Name:           "Retargeting Campaign - Cart Abandoners",
			Status:         "ACTIVE",
			ObjectiveType:  "CONVERSIONS",
			SpendCap:       0,
			DailyBudget:    75.00,
			LifetimeBudget: 0,
			BidStrategy:    "LOWEST_COST_WITH_BID_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, -2, 0),
			Updated:        yesterday,
		},
		{
			ID:             "23847239851",
			Name:           "Lead Generation - Newsletter Signup",
			Status:         "ACTIVE",
			ObjectiveType:  "LEAD_GENERATION",
			SpendCap:       500.00,
			DailyBudget:    25.00,
			LifetimeBudget: 0,
			BidStrategy:    "LOWEST_COST_WITHOUT_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, -1, -15),
			Updated:        yesterday.AddDate(0, 0, -3),
		},
		{
			ID:             "23847239852",
			Name:           "Holiday Special Promotion",
			Status:         "SCHEDULED",
			ObjectiveType:  "CONVERSIONS",
			SpendCap:       0,
			DailyBudget:    150.00,
			LifetimeBudget: 0,
			BidStrategy:    "LOWEST_COST_WITHOUT_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, 0, -2),
			Updated:        yesterday,
			StartTime:      now.AddDate(0, 0, 30), // 30 days in the future
			StopTime:       now.AddDate(0, 0, 45), // 45 days in the future
		},
		{
			ID:             "23847239853",
			Name:           "Winter Collection 2023",
			Status:         "SCHEDULED",
			ObjectiveType:  "CATALOG_SALES",
			SpendCap:       0,
			DailyBudget:    0,
			LifetimeBudget: 2000.00,
			BidStrategy:    "LOWEST_COST_WITHOUT_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, 0, -1),
			Updated:        yesterday,
			StartTime:      now.AddDate(0, 1, 0), // 1 month in the future
			StopTime:       now.AddDate(0, 2, 0), // 2 months in the future
		},
		{
			ID:             "23847239854",
			Name:           "App Install Campaign",
			Status:         "ACTIVE",
			ObjectiveType:  "APP_INSTALLS",
			SpendCap:       1500.00,
			DailyBudget:    50.00,
			LifetimeBudget: 0,
			BidStrategy:    "LOWEST_COST_WITH_BID_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, -3, 0),
			Updated:        yesterday.AddDate(0, 0, -1),
		},
		{
			ID:             "23847239855",
			Name:           "Video Views - Product Demo",
			Status:         "ACTIVE",
			ObjectiveType:  "VIDEO_VIEWS",
			SpendCap:       0,
			DailyBudget:    30.00,
			LifetimeBudget: 0,
			BidStrategy:    "LOWEST_COST_WITHOUT_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, -1, -10),
			Updated:        yesterday,
		},
		{
			ID:             "23847239856",
			Name:           "Store Traffic Campaign - New York",
			Status:         "PAUSED",
			ObjectiveType:  "STORE_TRAFFIC",
			SpendCap:       0,
			DailyBudget:    45.00,
			LifetimeBudget: 0,
			BidStrategy:    "LOWEST_COST_WITHOUT_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, -2, -15),
			Updated:        yesterday.AddDate(0, 0, -10),
		},
		{
			ID:             "23847239857",
			Name:           "Page Likes Campaign",
			Status:         "ARCHIVED",
			ObjectiveType:  "PAGE_LIKES",
			SpendCap:       0,
			DailyBudget:    0,
			LifetimeBudget: 300.00,
			BidStrategy:    "LOWEST_COST_WITHOUT_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, -6, 0),
			Updated:        yesterday.AddDate(0, -1, 0),
		},
		{
			ID:             "23847239858",
			Name:           "Messages Campaign - Customer Support",
			Status:         "ACTIVE",
			ObjectiveType:  "MESSAGES",
			SpendCap:       0,
			DailyBudget:    20.00,
			LifetimeBudget: 0,
			BidStrategy:    "LOWEST_COST_WITHOUT_CAP",
			BuyingType:     "AUCTION",
			Created:        yesterday.AddDate(0, -1, -5),
			Updated:        yesterday,
		},
	}
}

// UpdateCampaign updates an existing campaign with the provided parameters
func (c *Client) UpdateCampaign(campaignID string, params url.Values) error {
	// Create the endpoint URL with the campaign ID
	endpoint := fmt.Sprintf("%s/%s", c.auth.GetAPIBaseURL(), campaignID)

	// Create the request
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add authentication
	c.auth.AuthenticateRequest(req)

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// Parse the response
	var result struct {
		Success bool `json:"success"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("API did not return success")
	}

	return nil
}
