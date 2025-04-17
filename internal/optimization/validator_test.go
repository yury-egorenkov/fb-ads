package optimization

import (
	"testing"
	"time"

	"github.com/user/fb-ads/pkg/utils"
)

func TestValidateCampaignData(t *testing.T) {
	// Create a validator with default thresholds
	validator := NewPerformanceValidator()

	// Test case 1: Empty data
	t.Run("EmptyData", func(t *testing.T) {
		result := validator.ValidateCampaignData("test-campaign", []utils.CampaignPerformance{})
		if result.IsValid {
			t.Errorf("Expected validation to fail with empty data, but got valid")
		}
		if len(result.Reasons) != 1 || result.Reasons[0] != "No performance data available" {
			t.Errorf("Expected 'No performance data available' reason, got: %v", result.Reasons)
		}
	})

	// Test case 2: Sufficient data
	t.Run("SufficientData", func(t *testing.T) {
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		performances := []utils.CampaignPerformance{
			{
				CampaignID:  "test-campaign",
				Name:        "Test Campaign",
				Impressions: 1500,
				Clicks:      30,
				Spend:       10.0,
				LastUpdated: yesterday,
			},
			{
				CampaignID:  "test-campaign",
				Name:        "Test Campaign",
				Impressions: 1500,
				Clicks:      30,
				Spend:       10.0,
				LastUpdated: now,
			},
		}

		result := validator.ValidateCampaignData("test-campaign", performances)
		if !result.IsValid {
			t.Errorf("Expected validation to succeed with sufficient data, but got invalid: %v", result.Reasons)
		}
		if result.Metrics.TotalImpressions != 3000 {
			t.Errorf("Expected 3000 total impressions, got: %d", result.Metrics.TotalImpressions)
		}
		if result.Metrics.TotalClicks != 60 {
			t.Errorf("Expected 60 total clicks, got: %d", result.Metrics.TotalClicks)
		}
		if result.Metrics.TotalSpend != 20.0 {
			t.Errorf("Expected $20.0 total spend, got: $%.2f", result.Metrics.TotalSpend)
		}
	})

	// Test case 3: Insufficient impressions
	t.Run("InsufficientImpressions", func(t *testing.T) {
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		performances := []utils.CampaignPerformance{
			{
				CampaignID:  "test-campaign",
				Name:        "Test Campaign",
				Impressions: 400,
				Clicks:      30,
				Spend:       10.0,
				LastUpdated: yesterday,
			},
			{
				CampaignID:  "test-campaign",
				Name:        "Test Campaign",
				Impressions: 400,
				Clicks:      30,
				Spend:       10.0,
				LastUpdated: now,
			},
		}

		result := validator.ValidateCampaignData("test-campaign", performances)
		if result.IsValid {
			t.Errorf("Expected validation to fail with insufficient impressions, but got valid")
		}
		// Check if the reason contains information about insufficient impressions
		found := false
		for _, reason := range result.Reasons {
			if containsString(reason, "Insufficient impressions") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected reason about insufficient impressions, got: %v", result.Reasons)
		}
	})

	// Test case 4: Insufficient running time
	t.Run("InsufficientRunningTime", func(t *testing.T) {
		now := time.Now()
		twoHoursAgo := now.Add(-2 * time.Hour)
		performances := []utils.CampaignPerformance{
			{
				CampaignID:  "test-campaign",
				Name:        "Test Campaign",
				Impressions: 1500,
				Clicks:      30,
				Spend:       10.0,
				LastUpdated: twoHoursAgo,
			},
			{
				CampaignID:  "test-campaign",
				Name:        "Test Campaign",
				Impressions: 1500,
				Clicks:      30,
				Spend:       10.0,
				LastUpdated: now,
			},
		}

		result := validator.ValidateCampaignData("test-campaign", performances)
		if result.IsValid {
			t.Errorf("Expected validation to fail with insufficient running time, but got valid")
		}
		// Check if the reason contains information about insufficient running time
		found := false
		for _, reason := range result.Reasons {
			if containsString(reason, "Insufficient running time") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected reason about insufficient running time, got: %v", result.Reasons)
		}
	})

	// Test case 5: Wait recommendation
	t.Run("WaitRecommendation", func(t *testing.T) {
		now := time.Now()
		twoHoursAgo := now.Add(-2 * time.Hour)
		performances := []utils.CampaignPerformance{
			{
				CampaignID:  "test-campaign",
				Name:        "Test Campaign",
				Impressions: 400,
				Clicks:      5,
				Spend:       5.0,
				LastUpdated: twoHoursAgo,
			},
			{
				CampaignID:  "test-campaign",
				Name:        "Test Campaign",
				Impressions: 400,
				Clicks:      5,
				Spend:       5.0,
				LastUpdated: now,
			},
		}

		result := validator.ValidateCampaignData("test-campaign", performances)
		if result.IsValid {
			t.Errorf("Expected validation to fail, but got valid")
		}
		if !result.RecommendWait {
			t.Errorf("Expected wait recommendation, but got no wait recommendation")
		}
		if result.WaitTimeNeeded == 0 {
			t.Errorf("Expected non-zero wait time needed, but got zero")
		}
	})

	// Test case 6: Custom thresholds
	t.Run("CustomThresholds", func(t *testing.T) {
		customValidator := NewPerformanceValidator()
		customValidator.SetThresholds(ValidationThresholds{
			MinImpressions:   500,
			MinClicks:        5,
			MinRunningTime:   1 * time.Hour,
			MinDataPoints:    2,
			MinSpend:         1.0,
			EvaluationPeriod: 24 * time.Hour,
		})

		now := time.Now()
		twoHoursAgo := now.Add(-2 * time.Hour)
		performances := []utils.CampaignPerformance{
			{
				CampaignID:  "test-campaign",
				Name:        "Test Campaign",
				Impressions: 300,
				Clicks:      10,
				Spend:       10.0,
				LastUpdated: twoHoursAgo,
			},
			{
				CampaignID:  "test-campaign",
				Name:        "Test Campaign",
				Impressions: 300,
				Clicks:      10,
				Spend:       10.0,
				LastUpdated: now,
			},
		}

		result := customValidator.ValidateCampaignData("test-campaign", performances)
		if result.IsValid {
			t.Errorf("Expected validation to fail with custom thresholds, but got valid")
		}
		// Check if the reason contains information about insufficient impressions but not clicks
		impressionsFound := false
		clicksFound := false
		for _, reason := range result.Reasons {
			if containsString(reason, "Insufficient impressions") {
				impressionsFound = true
			}
			if containsString(reason, "Insufficient clicks") {
				clicksFound = true
			}
		}
		if !impressionsFound {
			t.Errorf("Expected reason about insufficient impressions, got: %v", result.Reasons)
		}
		if clicksFound {
			t.Errorf("Did not expect reason about insufficient clicks, got: %v", result.Reasons)
		}
	})
}

func TestValidateCampaignsData(t *testing.T) {
	validator := NewPerformanceValidator()
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// Create test data for multiple campaigns
	campaignPerformances := map[string][]utils.CampaignPerformance{
		"campaign-1": {
			{
				CampaignID:  "campaign-1",
				Name:        "Campaign 1",
				Impressions: 1500,
				Clicks:      30,
				Spend:       10.0,
				LastUpdated: yesterday,
			},
			{
				CampaignID:  "campaign-1",
				Name:        "Campaign 1",
				Impressions: 1500,
				Clicks:      30,
				Spend:       10.0,
				LastUpdated: now,
			},
		},
		"campaign-2": {
			{
				CampaignID:  "campaign-2",
				Name:        "Campaign 2",
				Impressions: 400,
				Clicks:      5,
				Spend:       5.0,
				LastUpdated: yesterday,
			},
			{
				CampaignID:  "campaign-2",
				Name:        "Campaign 2",
				Impressions: 400,
				Clicks:      5,
				Spend:       5.0,
				LastUpdated: now,
			},
		},
	}

	results := validator.ValidateCampaignsData(campaignPerformances)

	if len(results) != 2 {
		t.Errorf("Expected 2 validation results, got: %d", len(results))
	}

	if !results["campaign-1"].IsValid {
		t.Errorf("Expected campaign-1 to be valid, but got invalid: %v", results["campaign-1"].Reasons)
	}

	if results["campaign-2"].IsValid {
		t.Errorf("Expected campaign-2 to be invalid, but got valid")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}