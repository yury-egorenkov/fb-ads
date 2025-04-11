package audience

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/user/fb-ads/pkg/auth"
)

// AudienceSegment represents a Facebook audience segment
type AudienceSegment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"` // interest, behavior, demographic
	Path        string `json:"path,omitempty"`
	Size        int64  `json:"audience_size,omitempty"`
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
}

// NewAudienceAnalyzer creates a new audience analyzer
func NewAudienceAnalyzer(auth *auth.FacebookAuth, accountID string) *AudienceAnalyzer {
	return &AudienceAnalyzer{
		httpClient: &http.Client{},
		auth:       auth,
		accountID:  accountID,
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
	
	return audienceResp.Data, nil
}

// ExportAudienceData exports audience data to a file
func (a *AudienceAnalyzer) ExportAudienceData(filePath string, data []AudienceSegment) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling audience data: %w", err)
	}
	
	return nil // TODO: Implement file writing
}