package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/user/fb-ads/pkg/auth"
	"github.com/user/fb-ads/pkg/utils"
)

// TimeRange represents a time range for metrics query
type TimeRange struct {
	Since string `json:"since"`
	Until string `json:"until"`
}

// InsightsRequest represents a request for campaign insights
type InsightsRequest struct {
	Level          string    `json:"level"`           // campaign, adset, ad
	IDs            []string  `json:"ids,omitempty"`   // specific IDs to filter
	TimeRange      TimeRange `json:"time_range"`
	Fields         []string  `json:"fields"`
	Filtering      []Filter  `json:"filtering,omitempty"`
	BreakdownsType string    `json:"breakdowns_type,omitempty"` // age, gender, country, etc.
}

// Filter represents a filter for insights query
type Filter struct {
	Field    string        `json:"field"`
	Operator string        `json:"operator"`
	Value    interface{}   `json:"value"`
}

// MetricsCollector handles collection of campaign metrics
type MetricsCollector struct {
	httpClient *http.Client
	auth       *auth.FacebookAuth
	accountID  string
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(auth *auth.FacebookAuth, accountID string) *MetricsCollector {
	return &MetricsCollector{
		httpClient: &http.Client{},
		auth:       auth,
		accountID:  accountID,
	}
}

// CollectCampaignMetrics collects metrics for campaigns
func (m *MetricsCollector) CollectCampaignMetrics(request InsightsRequest) ([]utils.CampaignPerformance, error) {
	// Set default fields if not provided
	if len(request.Fields) == 0 {
		request.Fields = []string{
			"campaign_name",
			"spend",
			"impressions",
			"clicks",
			"actions",
			"cpm",
			"cpc",
			"ctr",
			"cost_per_action_type",
		}
	}
	
	params := url.Values{}
	params.Set("level", request.Level)
	params.Set("fields", strings.Join(request.Fields, ","))
	
	// Add time range
	timeRangeJSON, _ := json.Marshal(request.TimeRange)
	params.Set("time_range", string(timeRangeJSON))
	
	// Add filtering if present
	if len(request.Filtering) > 0 {
		filteringJSON, _ := json.Marshal(request.Filtering)
		params.Set("filtering", string(filteringJSON))
	}
	
	// Add breakdown if present
	if request.BreakdownsType != "" {
		params.Set("breakdowns", request.BreakdownsType)
	}
	
	endpoint := fmt.Sprintf("act_%s/insights", m.accountID)
	
	req, err := m.auth.GetAuthenticatedRequest(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}
	
	// Parse the response into a raw map first
	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	// Extract the data array
	dataArray, ok := rawResponse["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	
	// Process the data into campaign performances
	var performances []utils.CampaignPerformance
	
	for _, item := range dataArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Extract campaign ID from the response
		campaignID, _ := itemMap["campaign_id"].(string)
		
		// Extract campaign name
		campaignName, _ := itemMap["campaign_name"].(string)
		
		// Extract metrics
		spend, _ := itemMap["spend"].(float64)
		impressions, _ := itemMap["impressions"].(float64)
		clicks, _ := itemMap["clicks"].(float64)
		ctr, _ := itemMap["ctr"].(float64)
		cpm, _ := itemMap["cpm"].(float64)
		
		// Calculate conversions from actions
		var conversions int
		if actions, ok := itemMap["actions"].([]interface{}); ok {
			for _, action := range actions {
				actionMap, ok := action.(map[string]interface{})
				if !ok {
					continue
				}
				
				actionType, _ := actionMap["action_type"].(string)
				if actionType == "offsite_conversion" {
					value, _ := actionMap["value"].(float64)
					conversions += int(value)
				}
			}
		}
		
		// Calculate ROAS
		var roas float64 = 0
		if spend > 0 && conversions > 0 {
			// This is a simplified ROAS calculation
			// In a real implementation, you would need to get the actual conversion value
			averageOrderValue := 50.0 // Example: average order is worth $50
			roas = float64(conversions) * averageOrderValue / spend
		}
		
		// Create campaign performance object
		performance := utils.CampaignPerformance{
			CampaignID:  campaignID,
			Name:        campaignName,
			Spend:       spend,
			Impressions: int(impressions),
			Clicks:      int(clicks),
			Conversions: conversions,
			CPC:         calculateSafeCPC(spend, clicks),
			CPM:         cpm,
			CTR:         ctr * 100, // Convert to percentage
			ROAS:        roas,
			LastUpdated: time.Now(),
		}
		
		performances = append(performances, performance)
	}
	
	return performances, nil
}

// StoreMetrics stores collected metrics to a file or database
func (m *MetricsCollector) StoreMetrics(performances []utils.CampaignPerformance, filePath string) error {
	// Create a statistics manager with file storage
	statsManager := NewStatisticsManager(m, StorageTypeFile, filepath.Dir(filePath))
	
	// Store the metrics
	if err := statsManager.StoreStatistics(performances); err != nil {
		return fmt.Errorf("error storing metrics: %w", err)
	}
	
	return nil
}

// calculateSafeCPC calculates CPC (Cost Per Click) safely by avoiding division by zero
func calculateSafeCPC(spend, clicks float64) float64 {
	if clicks <= 0 {
		return 0
	}
	return spend / clicks
}