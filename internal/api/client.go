package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/user/fb-ads/pkg/auth"
	"github.com/user/fb-ads/pkg/models"
)

// Client is the Facebook Marketing API client
type Client struct {
	httpClient  *http.Client
	auth        *auth.FacebookAuth
	accountID   string
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
				ID:            getString(campaignMap, "id"),
				Name:          getString(campaignMap, "name"),
				Status:        getString(campaignMap, "status"),
				ObjectiveType: getString(campaignMap, "objective"),
				SpendCap:      getFloat(campaignMap, "spend_cap"),
				DailyBudget:   getFloat(campaignMap, "daily_budget"),
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
		time.RFC3339,                      // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05-0700",        // Another common format
		"2006-01-02T15:04:05",             // Without timezone
		"2006-01-02T15:04:05-07:00",       // With timezone
		"2006-01-02T15:04:05+0000",        // Yet another format
		"2006-01-02",                      // Just date
		time.RFC1123,                      // Mon, 02 Jan 2006 15:04:05 MST
		time.RFC1123Z,                     // Mon, 02 Jan 2006 15:04:05 -0700
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
			StartTime:      now.AddDate(0, 0, 30),  // 30 days in the future
			StopTime:       now.AddDate(0, 0, 45),  // 45 days in the future
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
			StartTime:      now.AddDate(0, 1, 0),  // 1 month in the future
			StopTime:       now.AddDate(0, 2, 0),  // 2 months in the future
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