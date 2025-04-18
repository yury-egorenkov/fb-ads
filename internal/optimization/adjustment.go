package optimization

import (
	"math"
	"time"
)

// CampaignAdjustment represents CPM adjustment data for a campaign
type CampaignAdjustment struct {
	CampaignID   string
	CurrentCPM   float64
	AdjustedCPM  float64
	AdjustmentTS time.Time
}

// Adjuster provides methods for adjusting campaign CPM bids
type Adjuster struct {
	statAnalyzer     *StatisticalAnalyzer
	maxCPM           float64
	minCPM           float64
	incrementPercent float64
	decrementPercent float64
	waitHours        int // Hours to wait before applying another adjustment
}

// NewAdjuster creates a new instance of Adjuster
func NewAdjuster(maxCPM, minCPM, incrementPercent, decrementPercent float64, waitHours int) *Adjuster {
	return &Adjuster{
		statAnalyzer:     NewStatisticalAnalyzer(),
		maxCPM:           maxCPM,
		minCPM:           minCPM,
		incrementPercent: incrementPercent,
		decrementPercent: decrementPercent,
		waitHours:        waitHours,
	}
}

// CalculateAdjustments calculates CPM adjustments for campaigns based on performance
func (a *Adjuster) CalculateAdjustments(
	campaigns []CampaignPerformance,
	previousAdjustments []CampaignAdjustment,
) []CampaignAdjustment {
	if len(campaigns) == 0 {
		return []CampaignAdjustment{}
	}

	// Get CPM values from all campaigns
	cpmValues := make([]float64, 0, len(campaigns))
	for _, campaign := range campaigns {
		cpmValues = append(cpmValues, campaign.CPM)
	}

	// Calculate optimal CPM using mean + 1 standard deviation approach
	optimalCPM := a.statAnalyzer.CalculateOptimalCPM(cpmValues, a.maxCPM)

	// Prepare adjustments map to track the last adjustment time for each campaign
	lastAdjustment := make(map[string]time.Time)
	for _, adj := range previousAdjustments {
		lastAdjustment[adj.CampaignID] = adj.AdjustmentTS
	}

	// Generate new adjustments
	adjustments := make([]CampaignAdjustment, 0, len(campaigns))
	now := time.Now()

	for _, campaign := range campaigns {
		// Skip campaigns that were adjusted recently (within waitHours)
		if lastTime, exists := lastAdjustment[campaign.CampaignID]; exists {
			hoursSinceLastAdjustment := now.Sub(lastTime).Hours()
			if hoursSinceLastAdjustment < float64(a.waitHours) {
				// Keep the current CPM if we can't adjust yet
				adjustments = append(adjustments, CampaignAdjustment{
					CampaignID:   campaign.CampaignID,
					CurrentCPM:   campaign.CPM,
					AdjustedCPM:  campaign.CPM,
					AdjustmentTS: lastTime,
				})
				continue
			}
		}

		// Calculate new CPM based on performance comparison
		newCPM := a.calculateNewCPM(campaign, optimalCPM)

		// Ensure new CPM is within acceptable range
		newCPM = math.Max(a.minCPM, math.Min(a.maxCPM, newCPM))

		// Add to adjustments
		adjustments = append(adjustments, CampaignAdjustment{
			CampaignID:   campaign.CampaignID,
			CurrentCPM:   campaign.CPM,
			AdjustedCPM:  newCPM,
			AdjustmentTS: now,
		})
	}

	return adjustments
}

// calculateNewCPM determines the new CPM for a campaign based on its performance
func (a *Adjuster) calculateNewCPM(campaign CampaignPerformance, optimalCPM float64) float64 {
	currentCPM := campaign.CPM

	// If current CPM is significantly below optimal, increase it
	if currentCPM < (optimalCPM * 0.8) {
		// Increase by incrementPercent
		return currentCPM * (1 + a.incrementPercent/100)
	}

	// If current CPM is significantly above optimal, decrease it
	if currentCPM > (optimalCPM * 1.2) {
		// Decrease by decrementPercent
		return currentCPM * (1 - a.decrementPercent/100)
	}

	// If CPM is within reasonable range of optimal, make smaller adjustment
	if currentCPM < optimalCPM {
		// Small increase (half of normal increment)
		return currentCPM * (1 + (a.incrementPercent/2)/100)
	} else if currentCPM > optimalCPM {
		// Small decrease (half of normal decrement)
		return currentCPM * (1 - (a.decrementPercent/2)/100)
	}

	// CPM is already optimal
	return currentCPM
}

// GetEligibleCampaigns returns campaigns eligible for adjustment (waited long enough)
func (a *Adjuster) GetEligibleCampaigns(
	campaignIDs []string,
	previousAdjustments []CampaignAdjustment,
) []string {
	if len(campaignIDs) == 0 {
		return []string{}
	}

	// Prepare map to track the last adjustment time for each campaign
	lastAdjustment := make(map[string]time.Time)
	for _, adj := range previousAdjustments {
		lastAdjustment[adj.CampaignID] = adj.AdjustmentTS
	}

	// Find eligible campaigns
	eligible := make([]string, 0)
	now := time.Now()

	for _, id := range campaignIDs {
		// If no previous adjustment, campaign is eligible
		if _, exists := lastAdjustment[id]; !exists {
			eligible = append(eligible, id)
			continue
		}

		// Check if enough time has passed since last adjustment
		hoursSinceLastAdjustment := now.Sub(lastAdjustment[id]).Hours()
		if hoursSinceLastAdjustment >= float64(a.waitHours) {
			eligible = append(eligible, id)
		}
	}

	return eligible
}