package optimization

import (
	"bytes"
	"testing"
	"time"

	"github.com/user/fb-ads/pkg/models"
)

// MockCampaignGetter implements the CampaignGetter interface for testing
type MockCampaignGetter struct {
	campaign *models.CampaignDetails
	err      error
}

func (m *MockCampaignGetter) GetCampaignDetails(id string) (*models.CampaignDetails, error) {
	return m.campaign, m.err
}

func TestExportCampaignToWriter(t *testing.T) {
	// Create a sample campaign
	campaign := createSampleCampaign()

	// Create exporter with default config
	exporter := NewExporter(nil)

	// Create buffer to capture output
	var buf bytes.Buffer

	// Export campaign to buffer
	err := exporter.ExportCampaignToWriter(campaign, &buf)
	if err != nil {
		t.Fatalf("Failed to export campaign: %v", err)
	}

	// Check if output contains expected fields
	output := buf.String()
	expectedFields := []string{
		"campaign:",
		"name:", 
		"total_budget:", 
		"test_budget_percentage:", 
		"max_cpm:",
		"creatives:",
		"id:",
		"title:",
		"description:",
		"image_url:",
		"targeting_options:",
		"audiences:",
		"parameters:",
		"placements:",
		"position:",
	}

	for _, field := range expectedFields {
		if !contains(output, field) {
			t.Errorf("Output missing expected field: %s", field)
		}
	}
}

func TestConvertToOptimizationConfig(t *testing.T) {
	// Create a sample campaign
	campaign := createSampleCampaign()

	// Create exporter with custom config
	config := &ExporterConfig{
		TotalBudget:          2000.0,
		TestBudgetPercentage: 15.0,
		MaxCPM:               10.0,
	}
	exporter := NewExporter(config)

	// Convert to optimization config
	result := exporter.convertToOptimizationConfig(campaign)

	// Check campaign section
	if result.Campaign.Name != "Test Campaign" {
		t.Errorf("Expected campaign name 'Test Campaign', got %s", result.Campaign.Name)
	}
	if result.Campaign.TotalBudget != 2000.0 {
		t.Errorf("Expected total budget 2000.0, got %.2f", result.Campaign.TotalBudget)
	}
	if result.Campaign.TestBudgetPercentage != 15.0 {
		t.Errorf("Expected test budget percentage 15.0, got %.2f", result.Campaign.TestBudgetPercentage)
	}
	if result.Campaign.MaxCPM != 10.0 {
		t.Errorf("Expected max CPM 10.0, got %.2f", result.Campaign.MaxCPM)
	}

	// Check creatives section
	if len(result.Creatives) != 1 {
		t.Fatalf("Expected 1 creative, got %d", len(result.Creatives))
	}
	if result.Creatives[0].Title != "Ad Title" {
		t.Errorf("Expected creative title 'Ad Title', got %s", result.Creatives[0].Title)
	}
	if result.Creatives[0].Description != "Ad Body" {
		t.Errorf("Expected creative description 'Ad Body', got %s", result.Creatives[0].Description)
	}
	if result.Creatives[0].ImageURL != "https://example.com/image.jpg" {
		t.Errorf("Expected creative image URL 'https://example.com/image.jpg', got %s", result.Creatives[0].ImageURL)
	}
	if result.Creatives[0].PageID != "123456789" {
		t.Errorf("Expected creative page ID '123456789', got %s", result.Creatives[0].PageID)
	}

	// Check targeting options
	if len(result.TargetingOptions.Audiences) != 1 {
		t.Fatalf("Expected 1 audience, got %d", len(result.TargetingOptions.Audiences))
	}
	if result.TargetingOptions.Audiences[0].Name != "Test AdSet" {
		t.Errorf("Expected audience name 'Test AdSet', got %s", result.TargetingOptions.Audiences[0].Name)
	}

	if len(result.TargetingOptions.Placements) < 1 {
		t.Fatalf("Expected at least 1 placement, got %d", len(result.TargetingOptions.Placements))
	}
}

// Helper function to create a sample campaign for testing
func createSampleCampaign() *models.CampaignDetails {
	targeting := make(map[string]interface{})
	targeting["age_min"] = 18
	targeting["age_max"] = 65
	targeting["publisher_platforms"] = []interface{}{"facebook"}
	targeting["facebook_positions"] = []interface{}{"feed"}

	return &models.CampaignDetails{
		ID:                "123456789",
		Name:              "Test Campaign",
		Status:            "ACTIVE",
		ObjectiveType:     "OUTCOME_AWARENESS",
		BuyingType:        "AUCTION",
		BidStrategy:       "LOWEST_COST_WITHOUT_CAP",
		DailyBudget:       5000, // In cents (50.00)
		LifetimeBudget:    0,
		SpendCap:          0,
		Created:           time.Now(),
		Updated:           time.Now(),
		SpecialAdCategories: []string{},
		AdSets: []models.AdSetDetails{
			{
				ID:               "987654321",
				Name:             "Test AdSet",
				Status:           "ACTIVE",
				OptimizationGoal: "REACH",
				BillingEvent:     "IMPRESSIONS",
				BidAmount:        1000, // In cents (10.00)
				Targeting:        targeting,
			},
		},
		Ads: []models.AdDetails{
			{
				ID:     "456789123",
				Name:   "Test Ad",
				Status: "ACTIVE",
				Creative: models.CreativeDetails{
					ID:               "789123456",
					Name:             "Test Creative",
					Title:            "Ad Title",
					Body:             "Ad Body",
					ImageURL:         "https://example.com/image.jpg",
					LinkURL:          "https://example.com",
					CallToActionType: "LEARN_MORE",
					PageID:           "123456789",
				},
			},
		},
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}