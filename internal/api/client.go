package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

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
	
	var campaignResp models.CampaignResponse
	if err := json.NewDecoder(resp.Body).Decode(&campaignResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	
	return &campaignResp, nil
}

// GetAllCampaigns retrieves all campaigns by handling pagination
func (c *Client) GetAllCampaigns() ([]models.Campaign, error) {
	var allCampaigns []models.Campaign
	var nextCursor string
	
	for {
		resp, err := c.GetCampaigns(100, nextCursor)
		if err != nil {
			return nil, err
		}
		
		allCampaigns = append(allCampaigns, resp.Data...)
		
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