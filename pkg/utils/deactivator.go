package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/user/fb-ads/pkg/auth"
)

// DeactivationRule represents a rule for deactivating campaigns
type DeactivationRule struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	MetricType         string  `json:"metric_type"` // CPA, ROAS, CTR, etc.
	Threshold          float64 `json:"threshold"`
	ComparisonOperator string  `json:"comparison_operator"` // >, <, =, >=, <=
	MinImpressions     int     `json:"min_impressions"`     // Minimum impressions before rule applies
	MinSpend           float64 `json:"min_spend"`          // Minimum spend before rule applies
	MinRuntime         int     `json:"min_runtime"`        // Minimum hours campaign should run before rule applies
}

// DeactivationEvent represents a campaign deactivation event
type DeactivationEvent struct {
	CampaignID  string    `json:"campaign_id"`
	AdSetID     string    `json:"adset_id,omitempty"`
	Name        string    `json:"name"`
	RuleID      string    `json:"rule_id"`
	RuleName    string    `json:"rule_name"`
	MetricValue float64   `json:"metric_value"`
	Threshold   float64   `json:"threshold"`
	Timestamp   time.Time `json:"timestamp"`
}

// Deactivator handles deactivation of underperforming campaigns
type Deactivator struct {
	httpClient *http.Client
	auth       *auth.FacebookAuth
	accountID  string
	rules      []DeactivationRule
}

// NewDeactivator creates a new campaign deactivator
func NewDeactivator(auth *auth.FacebookAuth, accountID string) *Deactivator {
	return &Deactivator{
		httpClient: &http.Client{},
		auth:       auth,
		accountID:  accountID,
		rules:      defaultRules(),
	}
}

// LoadRules loads deactivation rules from a file
func (d *Deactivator) LoadRules(filePath string) error {
	// TODO: Implement rule loading from a configuration file
	return nil
}

// defaultRules returns a set of default deactivation rules
func defaultRules() []DeactivationRule {
	return []DeactivationRule{
		{
			ID:                 "rule1",
			Name:               "High CPA Rule",
			MetricType:         "CPA",
			Threshold:          20.0, // $20 CPA threshold
			ComparisonOperator: ">",
			MinImpressions:     1000,
			MinSpend:           50.0, // $50 minimum spend
			MinRuntime:         24,   // 24 hours
		},
		{
			ID:                 "rule2",
			Name:               "Low CTR Rule",
			MetricType:         "CTR",
			Threshold:          0.5, // 0.5% CTR threshold
			ComparisonOperator: "<",
			MinImpressions:     3000,
			MinSpend:           30.0, // $30 minimum spend
			MinRuntime:         48,   // 48 hours
		},
		{
			ID:                 "rule3",
			Name:               "Low ROAS Rule",
			MetricType:         "ROAS",
			Threshold:          1.5, // 1.5x ROAS threshold
			ComparisonOperator: "<",
			MinImpressions:     2000,
			MinSpend:           100.0, // $100 minimum spend
			MinRuntime:         72,    // 72 hours
		},
	}
}

// CheckCampaigns checks all campaigns against deactivation rules
func (d *Deactivator) CheckCampaigns() ([]DeactivationEvent, error) {
	// Get campaign performance data
	optimizer := NewOptimizer(d.auth, d.accountID, 10.0) // Target CPA doesn't matter here
	performances, err := optimizer.GetCampaignPerformances()
	if err != nil {
		return nil, fmt.Errorf("error getting campaign performances: %w", err)
	}
	
	var events []DeactivationEvent
	
	for _, perf := range performances {
		// Check each rule
		for _, rule := range d.rules {
			// Skip if minimum requirements not met
			if perf.Impressions < rule.MinImpressions || perf.Spend < rule.MinSpend {
				continue
			}
			
			// Check campaign runtime
			campaignAge := time.Since(perf.LastUpdated).Hours()
			if int(campaignAge) < rule.MinRuntime {
				continue
			}
			
			// Get metric value based on rule type
			var metricValue float64
			switch rule.MetricType {
			case "CPA":
				if perf.Conversions == 0 {
					continue // Skip if no conversions
				}
				metricValue = perf.Spend / float64(perf.Conversions)
			case "CTR":
				if perf.Impressions == 0 {
					continue // Skip if no impressions
				}
				metricValue = float64(perf.Clicks) / float64(perf.Impressions) * 100
			case "ROAS":
				if perf.Spend == 0 {
					continue // Skip if no spend
				}
				metricValue = perf.ROAS
			default:
				continue // Skip unknown metric types
			}
			
			// Check if rule is triggered
			ruleTriggered := false
			switch rule.ComparisonOperator {
			case ">":
				ruleTriggered = metricValue > rule.Threshold
			case "<":
				ruleTriggered = metricValue < rule.Threshold
			case "=":
				ruleTriggered = metricValue == rule.Threshold
			case ">=":
				ruleTriggered = metricValue >= rule.Threshold
			case "<=":
				ruleTriggered = metricValue <= rule.Threshold
			}
			
			if ruleTriggered {
				events = append(events, DeactivationEvent{
					CampaignID:  perf.CampaignID,
					Name:        perf.Name,
					RuleID:      rule.ID,
					RuleName:    rule.Name,
					MetricValue: metricValue,
					Threshold:   rule.Threshold,
					Timestamp:   time.Now(),
				})
				
				// Deactivate the campaign
				if err := d.DeactivateCampaign(perf.CampaignID); err != nil {
					log.Printf("Error deactivating campaign %s: %v", perf.CampaignID, err)
				}
				
				// Break after first triggered rule
				break
			}
		}
	}
	
	return events, nil
}

// DeactivateCampaign deactivates a campaign by setting its status to PAUSED
func (d *Deactivator) DeactivateCampaign(campaignID string) error {
	params := url.Values{}
	params.Set("status", "PAUSED")
	
	// Create the endpoint URL with the campaign ID
	endpoint := fmt.Sprintf("%s/act_%s/campaigns/%s", d.auth.GetAPIBaseURL(), d.accountID, campaignID)

	// Create the request
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add authentication
	d.auth.AuthenticateRequest(req)

	// Send the request
	log.Printf("Deactivating campaign %s", campaignID)
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}
	
	return nil
}