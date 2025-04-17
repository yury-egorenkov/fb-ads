package audience

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/user/fb-ads/pkg/auth"
)

// AudienceSegment represents a Facebook audience segment
type AudienceSegment struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Type        string              `json:"type"`           // interest, behavior, demographic
	Path        interface{}         `json:"path,omitempty"` // Can be either string or array of strings
	LowerBound  int64               `json:"audience_size_lower_bound,omitempty"`
	UpperBound  int64               `json:"audience_size_upper_bound,omitempty"`
	Performance *SegmentPerformance `json:"performance,omitempty"`
	LastUpdated time.Time           `json:"last_updated,omitempty"`
}

// SegmentPerformance contains performance metrics for an audience segment
type SegmentPerformance struct {
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	Conversions int64   `json:"conversions"`
	Spend       float64 `json:"spend"`
	CPC         float64 `json:"cpc,omitempty"` // Cost per click
	CPM         float64 `json:"cpm,omitempty"` // Cost per 1000 impressions
	CTR         float64 `json:"ctr,omitempty"` // Click-through rate
	CVR         float64 `json:"cvr,omitempty"` // Conversion rate
	CPA         float64 `json:"cpa,omitempty"` // Cost per acquisition
}

// AudienceResponse represents the Facebook API response for audience data
type AudienceResponse struct {
	Data   []AudienceSegment `json:"data"`
	Paging struct {
		Cursors struct {
			Before string `json:"before,omitempty"`
			After  string `json:"after,omitempty"`
		} `json:"cursors"`
		Next string `json:"next,omitempty"`
	} `json:"paging"`
}

// AudienceAnalyzer handles audience data extraction and analysis
type AudienceAnalyzer struct {
	httpClient *http.Client
	auth       *auth.FacebookAuth
	accountID  string
	segments   map[string]AudienceSegment // Cache for audience segments
}

// NewAudienceAnalyzer creates a new audience analyzer
func NewAudienceAnalyzer(auth *auth.FacebookAuth, accountID string) *AudienceAnalyzer {
	return &AudienceAnalyzer{
		httpClient: &http.Client{},
		auth:       auth,
		accountID:  accountID,
		segments:   make(map[string]AudienceSegment),
	}
}

// Search retrieves targeting options
func (a *AudienceAnalyzer) Search(searchType string, class string, query string) ([]AudienceSegment, error) {
	params := url.Values{}
	params.Set("type", searchType)
	if len(class) > 0 {
		params.Set("class", class)
	}
	if len(query) > 0 {
		params.Set("q", query)
	}

	req, err := a.auth.GetAuthenticatedRequest("search", params)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// Read response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Print raw response for debugging (uncomment if needed)
	//fmt.Printf("Raw API response: %s\n", string(body))

	// Decode the JSON response
	var audienceResp AudienceResponse
	if err := json.Unmarshal(body, &audienceResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Update our segments cache
	for _, segment := range audienceResp.Data {
		a.segments[segment.ID] = segment
	}

	return audienceResp.Data, nil
}

// CollectSegmentStatistics gathers performance statistics for audience segments
func (a *AudienceAnalyzer) CollectSegmentStatistics(campaignID string, days int) error {
	// Set up endpoint and parameters for insights API call
	endpoint := fmt.Sprintf("/%s/insights", campaignID)
	params := url.Values{}

	// Get data from last N days
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	params.Set("time_range", fmt.Sprintf(`{"since":"%s","until":"%s"}`, startDate, endDate))

	// Try a simplified approach with a single demographic breakdown
	// This avoids potential conflicts with action_type that cause API errors
	params.Set("breakdowns", "age")

	// Explicitly request only standard metrics that don't require action_type
	params.Set("fields", "impressions,clicks,spend,cpm,ctr")

	req, err := a.auth.GetAuthenticatedRequest(endpoint, params)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// Process the response and update segment statistics
	// This is a simplified implementation; in a real system,
	// you would parse the response and update the appropriate segments

	return nil
}

// FilterAudiences filters audience segments based on criteria
func (a *AudienceAnalyzer) FilterAudiences(options map[string]interface{}) ([]AudienceSegment, error) {
	var filtered []AudienceSegment

	// Extract filter criteria
	minSize, hasMinSize := options["min_size"].(int64)
	maxSize, hasMaxSize := options["max_size"].(int64)
	types, hasTypes := options["types"].([]string)
	keywords, hasKeywords := options["keywords"].([]string)

	// Apply filters to all segments
	for _, segment := range a.segments {
		// Filter by size
		if hasMinSize && segment.LowerBound < minSize {
			continue
		}
		if hasMaxSize && segment.UpperBound > maxSize {
			continue
		}

		// Filter by type
		if hasTypes {
			typeMatch := false
			for _, t := range types {
				if segment.Type == t {
					typeMatch = true
					break
				}
			}
			if !typeMatch {
				continue
			}
		}

		// Filter by keywords
		if hasKeywords {
			keywordMatch := false
			for _, keyword := range keywords {
				if strings.Contains(strings.ToLower(segment.Name), strings.ToLower(keyword)) ||
					strings.Contains(strings.ToLower(segment.Description), strings.ToLower(keyword)) {
					keywordMatch = true
					break
				}
			}
			if !keywordMatch {
				continue
			}
		}

		// Segment passed all filters
		filtered = append(filtered, segment)
	}

	// Sort results by audience size (descending)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].UpperBound > filtered[j].UpperBound
	})

	return filtered, nil
}

// ExportAudienceData exports audience data to a file
func (a *AudienceAnalyzer) ExportAudienceData(filePath string, data []AudienceSegment) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling audience data: %w", err)
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing audience data to file: %w", err)
	}

	return nil
}

// ReachEstimateResponse represents the API response from the reach_estimate endpoint
type ReachEstimateResponse struct {
	Data []struct {
		EstimateReady bool  `json:"estimate_ready"`
		Users         int64 `json:"users"`
		LowerBound    int64 `json:"lower_bound"`
		UpperBound    int64 `json:"upper_bound"`
	} `json:"data"`
}

// FormatNumberReadable formats a number to a human-readable string (e.g., 1.2M, 450K)
func FormatNumberReadable(num int64) string {
	if num == 0 {
		return "0"
	}

	abs := math.Abs(float64(num))

	if abs >= 1e9 {
		// Billions
		value := float64(num) / 1000000000
		return fmt.Sprintf("%.0fb", value)
	}

	if abs >= 1e6 {
		// Millions
		value := float64(num) / 1000000
		return fmt.Sprintf("%.0fm", value)
	}

	if abs >= 1e3 {
		// Thousands
		value := float64(num) / 1000
		return fmt.Sprintf("%.0fk", value)
	}

	return fmt.Sprintf("%d", num)
}

// FormatAudienceRange formats audience range in a human-readable format
func FormatAudienceRange(lower, upper int64) string {
	if lower == 0 && upper == 0 {
		return "Unknown"
	}

	if lower == upper {
		return FormatNumberReadable(lower)
	}

	return fmt.Sprintf("%s - %s", FormatNumberReadable(lower), FormatNumberReadable(upper))
}

// GetAudienceSize retrieves the estimated audience size for a specific interest
func (a *AudienceAnalyzer) GetAudienceSize(interestID string) (int64, error) {
	// Construct the targeting spec for the interest
	targetingSpec := map[string]interface{}{
		"geo_locations": map[string]interface{}{
			"countries": []string{"US"}, // Default to US, could be made configurable
		},
		"interests": []map[string]string{
			{"id": interestID},
		},
	}

	// Marshal to JSON
	targetingJSON, err := json.Marshal(targetingSpec)
	if err != nil {
		return 0, fmt.Errorf("error marshaling targeting spec: %w", err)
	}

	// Set up the parameters for the reach_estimate endpoint
	params := url.Values{}
	params.Set("targeting_spec", string(targetingJSON))
	params.Set("optimization_goal", "REACH") // Required parameter for delivery_estimate

	// Build the endpoint with account ID
	endpoint := fmt.Sprintf("act_%s/delivery_estimate", a.accountID)

	req, err := a.auth.GetAuthenticatedRequest(endpoint, params)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %w", err)
	}
	fmt.Printf("Raw `delivery_estimate` response: %s\n", string(body))

	// Decode the JSON response
	var estimateResp ReachEstimateResponse
	if err := json.Unmarshal(body, &estimateResp); err != nil {
		return 0, fmt.Errorf("error decoding response: %w", err)
	}

	// Check if we have data
	if len(estimateResp.Data) == 0 {
		return 0, fmt.Errorf("no reach estimate data returned")
	}

	fmt.Printf("Audience size for %s: %s\n", interestID, FormatAudienceRange(estimateResp.Data[0].LowerBound, estimateResp.Data[0].UpperBound))

	// Return the estimated audience size
	return estimateResp.Data[0].Users, nil
}
