package optimization

import (
	"strings"
	"testing"
)

func TestCampaignGenerator_GenerateAllCombinations(t *testing.T) {
	// Create a test YAML configuration
	yamlConfig := `
campaign:
  name: "Test Campaign"
  total_budget: 1000.00
  test_budget_percentage: 20
  max_cpm: 15.00

creatives:
  - id: "creative1"
    title: "Creative 1"
    description: "Description 1"
    image_url: "https://example.com/image1.jpg"
    page_id: "123456789"
  - id: "creative2"
    title: "Creative 2"
    description: "Description 2"
    image_url: "https://example.com/image2.jpg"
    page_id: "123456789"

targeting_options:
  audiences:
    - id: "audience1"
      name: "Audience 1"
      parameters:
        age_min: 18
        age_max: 24
    - id: "audience2"
      name: "Audience 2"
      parameters:
        age_min: 25
        age_max: 34
  placements:
    - id: "placement1"
      name: "Placement 1"
      position: "feed"
    - id: "placement2"
      name: "Placement 2"
      position: "story"
`
	
	// Parse the configuration
	reader := strings.NewReader(yamlConfig)
	config, err := ParseYAMLReader(reader)
	if err != nil {
		t.Fatalf("Error parsing YAML: %v", err)
	}

	// Create budget calculator
	budgetCalc, err := NewBudgetCalculator(
		config.Campaign.TotalBudget,
		config.Campaign.TestBudgetPercentage,
		config.Campaign.MaxCPM,
	)
	if err != nil {
		t.Fatalf("Error creating budget calculator: %v", err)
	}

	// Create generator
	generator := NewCampaignGenerator(config, budgetCalc)
	if err := generator.GenerateAllCombinations(); err != nil {
		t.Fatalf("Error generating combinations: %v", err)
	}

	// Check total combinations
	// 2 creatives * (2 audiences + 2 placements) = 8 combinations
	if expected, got := 8, generator.TotalCombinations(); expected != got {
		t.Errorf("Expected %d combinations, got %d", expected, got)
	}

	// Test with limit
	generator.SetLimit(5)
	if err := generator.GenerateAllCombinations(); err != nil {
		t.Fatalf("Error generating combinations with limit: %v", err)
	}

	if expected, got := 5, generator.TotalCombinations(); expected != got {
		t.Errorf("Expected %d combinations with limit, got %d", expected, got)
	}

	// Test batch calculation
	generator.SetMaxBatchSize(3)
	if expected, got := 2, generator.TotalBatches(); expected != got {
		t.Errorf("Expected %d batches, got %d", expected, got)
	}

	// Test batch retrieval
	batch1 := generator.GetNextBatch()
	if expected, got := 3, len(batch1); expected != got {
		t.Errorf("Expected batch size %d, got %d", expected, got)
	}

	batch2 := generator.GetNextBatch()
	if expected, got := 2, len(batch2); expected != got {
		t.Errorf("Expected batch size %d, got %d", expected, got)
	}

	// Test empty batch after all combinations
	batch3 := generator.GetNextBatch()
	if expected, got := 0, len(batch3); expected != got {
		t.Errorf("Expected empty batch, got size %d", got)
	}

	// Test reset
	generator.ResetBatch()
	batchAfterReset := generator.GetNextBatch()
	if expected, got := 3, len(batchAfterReset); expected != got {
		t.Errorf("Expected batch size after reset %d, got %d", expected, got)
	}
}

func TestCampaignGenerator_ConvertToFacebookCampaign(t *testing.T) {
	// Create a test configuration
	config := &CampaignOptimizationConfig{
		Campaign: CampaignConfig{
			Name:                 "Test Campaign",
			TotalBudget:          1000.00,
			TestBudgetPercentage: 20,
			MaxCPM:               15.00,
		},
	}

	// Create budget calculator
	budgetCalc, _ := NewBudgetCalculator(
		config.Campaign.TotalBudget,
		config.Campaign.TestBudgetPercentage,
		config.Campaign.MaxCPM,
	)

	// Create generator
	generator := NewCampaignGenerator(config, budgetCalc)

	// Create a test combination
	combination := CampaignCombination{
		Name: "Test Combination",
		Creative: CreativeConfig{
			ID:          "creative1",
			Title:       "Test Creative",
			Description: "Test Description",
			ImageURL:    "https://example.com/image.jpg",
			PageID:      "123456789",
		},
		AudienceID:    "audience1",
		AudienceName:  "Test Audience",
		AudienceParams: map[string]interface{}{
			"age_min": float64(18),
			"age_max": float64(24),
		},
		Budget:       100.00,
		BidAmount:    10.00,
		TargetingType: "audience",
	}

	// Convert to Facebook campaign
	campaign := generator.ConvertToFacebookCampaign(combination)

	// Validate campaign fields
	if !strings.Contains(campaign.Name, "Test Combination") {
		t.Errorf("Expected campaign name to contain %q, got %q", "Test Combination", campaign.Name)
	}

	if expected, got := "PAUSED", campaign.Status; expected != got {
		t.Errorf("Expected campaign status %q, got %q", expected, got)
	}

	if expected, got := "OUTCOME_AWARENESS", campaign.Objective; expected != got {
		t.Errorf("Expected campaign objective %q, got %q", expected, got)
	}

	if expected, got := float64(100.00), campaign.LifetimeBudget; expected != got {
		t.Errorf("Expected campaign budget %.2f, got %.2f", expected, got)
	}

	// Validate ad set
	if len(campaign.AdSets) != 1 {
		t.Fatalf("Expected 1 ad set, got %d", len(campaign.AdSets))
	}

	adSet := campaign.AdSets[0]
	if !strings.Contains(adSet.Name, "AdSet") {
		t.Errorf("Expected ad set name to contain %q, got %q", "AdSet", adSet.Name)
	}

	if expected, got := "PAUSED", adSet.Status; expected != got {
		t.Errorf("Expected ad set status %q, got %q", expected, got)
	}

	if expected, got := "REACH", adSet.OptimizationGoal; expected != got {
		t.Errorf("Expected ad set optimization goal %q, got %q", expected, got)
	}

	if expected, got := "IMPRESSIONS", adSet.BillingEvent; expected != got {
		t.Errorf("Expected ad set billing event %q, got %q", expected, got)
	}

	if expected, got := float64(10.00), adSet.BidAmount; expected != got {
		t.Errorf("Expected ad set bid amount %.2f, got %.2f", expected, got)
	}

	// Validate audience targeting
	ageMin, hasAgeMin := adSet.Targeting["age_min"]
	if !hasAgeMin {
		t.Errorf("Expected targeting to have age_min")
	} else if ageMin != float64(18) {
		t.Errorf("Expected age_min %v, got %v", float64(18), ageMin)
	}

	ageMax, hasAgeMax := adSet.Targeting["age_max"]
	if !hasAgeMax {
		t.Errorf("Expected targeting to have age_max")
	} else if ageMax != float64(24) {
		t.Errorf("Expected age_max %v, got %v", float64(24), ageMax)
	}

	// Validate ad
	if len(campaign.Ads) != 1 {
		t.Fatalf("Expected 1 ad, got %d", len(campaign.Ads))
	}

	ad := campaign.Ads[0]
	if !strings.Contains(ad.Name, "Ad") {
		t.Errorf("Expected ad name to contain %q, got %q", "Ad", ad.Name)
	}

	if expected, got := "PAUSED", ad.Status; expected != got {
		t.Errorf("Expected ad status %q, got %q", expected, got)
	}

	// Validate creative
	if expected, got := "Test Creative", ad.Creative.Title; expected != got {
		t.Errorf("Expected creative title %q, got %q", expected, got)
	}

	if expected, got := "Test Description", ad.Creative.Body; expected != got {
		t.Errorf("Expected creative body %q, got %q", expected, got)
	}

	if expected, got := "https://example.com/image.jpg", ad.Creative.ImageURL; expected != got {
		t.Errorf("Expected creative image URL %q, got %q", expected, got)
	}

	if expected, got := "123456789", ad.Creative.PageID; expected != got {
		t.Errorf("Expected creative page ID %q, got %q", expected, got)
	}
}