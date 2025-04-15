package audience

import (
	"encoding/json"
	"fmt"
	"io"
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
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Type         string    `json:"type"` // interest, behavior, demographic
	Path         interface{} `json:"path,omitempty"` // Can be either string or array of strings
	Size         int64     `json:"audience_size,omitempty"`
	Performance  *SegmentPerformance `json:"performance,omitempty"`
	LastUpdated  time.Time `json:"last_updated,omitempty"`
}

// SegmentPerformance contains performance metrics for an audience segment
type SegmentPerformance struct {
	Impressions  int64   `json:"impressions"`
	Clicks       int64   `json:"clicks"`
	Conversions  int64   `json:"conversions"`
	Spend        float64 `json:"spend"`
	CPC          float64 `json:"cpc,omitempty"` // Cost per click
	CPM          float64 `json:"cpm,omitempty"` // Cost per 1000 impressions
	CTR          float64 `json:"ctr,omitempty"` // Click-through rate
	CVR          float64 `json:"cvr,omitempty"` // Conversion rate
	CPA          float64 `json:"cpa,omitempty"` // Cost per acquisition
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

// GetInterests retrieves interest targeting options
func (a *AudienceAnalyzer) GetInterests(query string) ([]AudienceSegment, error) {
	params := url.Values{}
	params.Set("type", "adinterest")
	params.Set("q", query)
	
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
	
	var audienceResp AudienceResponse
	if err := json.NewDecoder(resp.Body).Decode(&audienceResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	// Update our segments cache
	for _, segment := range audienceResp.Data {
		a.segments[segment.ID] = segment
	}
	
	return audienceResp.Data, nil
}

// GetBehaviors retrieves behavior targeting options
func (a *AudienceAnalyzer) GetBehaviors(query string) ([]AudienceSegment, error) {
	params := url.Values{}
	params.Set("type", "adsquizzedbehavior")
	params.Set("q", query)
	
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
	
	var audienceResp AudienceResponse
	if err := json.NewDecoder(resp.Body).Decode(&audienceResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	// Update our segments cache
	for _, segment := range audienceResp.Data {
		segment.Type = "behavior"
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
		if hasMinSize && segment.Size < minSize {
			continue
		}
		if hasMaxSize && segment.Size > maxSize {
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
		return filtered[i].Size > filtered[j].Size
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