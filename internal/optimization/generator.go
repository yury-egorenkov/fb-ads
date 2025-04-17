// Package optimization provides campaign optimization functionality
package optimization

import (
	"fmt"
	"time"

	"github.com/user/fb-ads/pkg/models"
)

// CampaignCombination represents a single test campaign combination
type CampaignCombination struct {
	Name            string
	Creative        CreativeConfig
	AudienceID      string
	AudienceName    string
	AudienceParams  map[string]interface{}
	PlacementID     string
	PlacementName   string
	PlacementParams string
	Budget          float64
	BidAmount       float64
	TargetingType   string // "audience" or "placement"
}

// CampaignGenerator handles the generation of test campaign combinations
type CampaignGenerator struct {
	Config        *CampaignOptimizationConfig
	BudgetCalc    *BudgetCalculator
	Combinations  []CampaignCombination
	MaxBatchSize  int
	CurrentBatch  int
	Priority      string // "audience" or "placement" - which to prioritize
	Limit         int    // Maximum number of combinations to generate (0 = no limit)
}

// NewCampaignGenerator creates a new campaign generator
func NewCampaignGenerator(config *CampaignOptimizationConfig, budgetCalc *BudgetCalculator) *CampaignGenerator {
	return &CampaignGenerator{
		Config:       config,
		BudgetCalc:   budgetCalc,
		MaxBatchSize: 5, // Default max batch size
		Priority:     "audience", // Default priority
	}
}

// SetLimit sets the maximum number of combinations to generate
func (g *CampaignGenerator) SetLimit(limit int) {
	g.Limit = limit
}

// SetMaxBatchSize sets the maximum batch size
func (g *CampaignGenerator) SetMaxBatchSize(size int) {
	if size > 0 {
		g.MaxBatchSize = size
	}
}

// SetPriority sets the priority for combination generation
func (g *CampaignGenerator) SetPriority(priority string) {
	if priority == "audience" || priority == "placement" {
		g.Priority = priority
	}
}

// GenerateAllCombinations generates all possible combinations
func (g *CampaignGenerator) GenerateAllCombinations() error {
	// Reset combinations
	g.Combinations = []CampaignCombination{}
	
	// Calculate total number of combinations
	totalCombinations := len(g.Config.Creatives) * 
		(len(g.Config.TargetingOptions.Audiences) + len(g.Config.TargetingOptions.Placements))
	
	// If limit is set, adjust the total
	actualTotal := totalCombinations
	if g.Limit > 0 && g.Limit < totalCombinations {
		actualTotal = g.Limit
	}
	
	// Calculate budget per campaign
	budgetPerCampaign, err := g.BudgetCalc.GetBudgetPerCampaign(actualTotal)
	if err != nil {
		return fmt.Errorf("error calculating budget per campaign: %w", err)
	}
	
	// Generate creative + audience combinations
	audienceCombinations := []CampaignCombination{}
	for _, creative := range g.Config.Creatives {
		for _, audience := range g.Config.TargetingOptions.Audiences {
			combination := CampaignCombination{
				Name:           fmt.Sprintf("%s - %s", g.Config.Campaign.Name, audience.Name),
				Creative:       creative,
				AudienceID:     audience.ID,
				AudienceName:   audience.Name,
				AudienceParams: audience.Parameters,
				Budget:         budgetPerCampaign,
				BidAmount:      g.Config.Campaign.MaxCPM,
				TargetingType:  "audience",
			}
			audienceCombinations = append(audienceCombinations, combination)
		}
	}
	
	// Generate creative + placement combinations
	placementCombinations := []CampaignCombination{}
	for _, creative := range g.Config.Creatives {
		for _, placement := range g.Config.TargetingOptions.Placements {
			combination := CampaignCombination{
				Name:            fmt.Sprintf("%s - %s", g.Config.Campaign.Name, placement.Name),
				Creative:        creative,
				PlacementID:     placement.ID,
				PlacementName:   placement.Name,
				PlacementParams: placement.Position,
				Budget:          budgetPerCampaign,
				BidAmount:       g.Config.Campaign.MaxCPM,
				TargetingType:   "placement",
			}
			placementCombinations = append(placementCombinations, combination)
		}
	}
	
	// Combine based on priority
	if g.Priority == "audience" {
		g.Combinations = append(audienceCombinations, placementCombinations...)
	} else {
		g.Combinations = append(placementCombinations, audienceCombinations...)
	}
	
	// Apply limit if specified
	if g.Limit > 0 && len(g.Combinations) > g.Limit {
		g.Combinations = g.Combinations[:g.Limit]
	}
	
	return nil
}

// GetNextBatch returns the next batch of combinations
func (g *CampaignGenerator) GetNextBatch() []CampaignCombination {
	start := g.CurrentBatch * g.MaxBatchSize
	if start >= len(g.Combinations) {
		return []CampaignCombination{} // No more combinations
	}
	
	end := start + g.MaxBatchSize
	if end > len(g.Combinations) {
		end = len(g.Combinations)
	}
	
	batch := g.Combinations[start:end]
	g.CurrentBatch++
	
	return batch
}

// ResetBatch resets the batch counter
func (g *CampaignGenerator) ResetBatch() {
	g.CurrentBatch = 0
}

// TotalCombinations returns the total number of combinations
func (g *CampaignGenerator) TotalCombinations() int {
	return len(g.Combinations)
}

// TotalBatches returns the total number of batches
func (g *CampaignGenerator) TotalBatches() int {
	if len(g.Combinations) == 0 {
		return 0
	}
	
	batches := len(g.Combinations) / g.MaxBatchSize
	if len(g.Combinations) % g.MaxBatchSize > 0 {
		batches++
	}
	
	return batches
}

// ConvertToFacebookCampaign converts a combination to Facebook campaign config
func (g *CampaignGenerator) ConvertToFacebookCampaign(combination CampaignCombination) *models.CampaignConfig {
	// Generate a unique name with timestamp
	timestamp := time.Now().Format("20060102-150405")
	campaignName := fmt.Sprintf("%s (%s)", combination.Name, timestamp)
	
	// Calculate start and end times for lifetime budget
	startTime := time.Now()
	endTime := startTime.Add(7 * 24 * time.Hour) // End time is 7 days from now
	
	// Create base campaign config
	campaign := &models.CampaignConfig{
		Name:           campaignName,
		Status:         "PAUSED", // Always start paused for safety
		Objective:      "OUTCOME_AWARENESS", // Using awareness for test campaigns
		BuyingType:     "AUCTION",
		BidStrategy:    "LOWEST_COST_WITH_BID_CAP",
		LifetimeBudget: combination.Budget,
		StartTime:      startTime.Format(time.RFC3339),
		EndTime:        endTime.Format(time.RFC3339), // Required for lifetime budget
		AdSets:         []models.AdSetConfig{},
		Ads:            []models.AdConfig{},
	}
	
	// Create ad set based on targeting type
	adSet := models.AdSetConfig{
		Name:             fmt.Sprintf("AdSet - %s", campaignName),
		Status:           "PAUSED",
		OptimizationGoal: "REACH",
		BillingEvent:     "IMPRESSIONS",
		BidAmount:        combination.BidAmount,
		StartTime:        startTime.Format(time.RFC3339),
		EndTime:          endTime.Format(time.RFC3339), // Required for lifetime budget
		Targeting:        make(map[string]interface{}),
	}
	
	// Set up targeting based on type
	if combination.TargetingType == "audience" {
		// Copy all audience parameters to the targeting
		for key, value := range combination.AudienceParams {
			adSet.Targeting[key] = value
		}
	} else if combination.TargetingType == "placement" {
		// Set up placement targeting
		adSet.Targeting["publisher_platforms"] = []string{"facebook", "instagram"}
		
		// Add specific placement based on position
		switch combination.PlacementParams {
		case "feed":
			adSet.Targeting["facebook_positions"] = []string{"feed"}
		case "story":
			adSet.Targeting["instagram_positions"] = []string{"story"}
		case "right_hand_column":
			adSet.Targeting["facebook_positions"] = []string{"right_hand_column"}
		default:
			// Use all positions if not specified
			adSet.Targeting["facebook_positions"] = []string{"feed"}
		}
		
		// Add required location targeting (required by Facebook API)
		adSet.Targeting["geo_locations"] = map[string]interface{}{
			"countries": []string{"US"},
			"location_types": []string{"home", "recent"},
		}
		
		// Add minimal age targeting (required by Facebook API)
		adSet.Targeting["age_min"] = 18
		adSet.Targeting["age_max"] = 65
	}
	
	campaign.AdSets = append(campaign.AdSets, adSet)
	
	// Create ad with creative
	ad := models.AdConfig{
		Name:   fmt.Sprintf("Ad - %s", campaignName),
		Status: "PAUSED",
		Creative: models.CreativeConfig{
			Title:        combination.Creative.Title,
			Body:         combination.Creative.Description,
			ImageURL:     combination.Creative.ImageURL,
			LinkURL:      combination.Creative.LinkURL,
			CallToAction: combination.Creative.CallToAction,
			PageID:       combination.Creative.PageID,
		},
	}
	
	campaign.Ads = append(campaign.Ads, ad)
	
	return campaign
}