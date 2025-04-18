package optimization

import (
	"sort"
)

// CampaignPerformance represents the performance metrics of a campaign
type CampaignPerformance struct {
	CampaignID   string
	Impressions  int
	Clicks       int
	Conversions  int
	Cost         float64
	CPM          float64
	CTR          float64
	CPC          float64
}

// Terminator is responsible for determining which campaigns should be terminated
type Terminator struct {
	minImpressions int // Minimum number of impressions required for a valid campaign
}

// NewTerminator creates a new instance of Terminator
func NewTerminator(minImpressions int) *Terminator {
	return &Terminator{
		minImpressions: minImpressions,
	}
}

// GetCampaignsToTerminate identifies campaigns that should be terminated
// based on performance data and the termination criteria
func (t *Terminator) GetCampaignsToTerminate(campaigns []CampaignPerformance) []string {
	if len(campaigns) == 0 {
		return []string{}
	}

	// Filter out campaigns that don't meet the minimum impression threshold
	validCampaigns := t.filterValidCampaigns(campaigns)
	if len(validCampaigns) == 0 {
		return []string{}
	}

	// Find the worst performing active campaign among valid campaigns
	worstActive := t.findWorstActiveCampaign(validCampaigns)
	
	// Identify campaigns to terminate (those with fewer impressions than the worst active)
	campaignsToTerminate := []string{}
	for _, campaign := range campaigns {
		// If campaign has fewer impressions than the worst active campaign, terminate it
		if campaign.Impressions < worstActive.Impressions {
			campaignsToTerminate = append(campaignsToTerminate, campaign.CampaignID)
		}
	}
	
	return campaignsToTerminate
}

// filterValidCampaigns filters out campaigns that don't meet the minimum impression threshold
func (t *Terminator) filterValidCampaigns(campaigns []CampaignPerformance) []CampaignPerformance {
	validCampaigns := []CampaignPerformance{}
	
	for _, campaign := range campaigns {
		if campaign.Impressions >= t.minImpressions {
			validCampaigns = append(validCampaigns, campaign)
		}
	}
	
	return validCampaigns
}

// findWorstActiveCampaign finds the campaign with the lowest impressions
// among the campaigns that meet the minimum threshold
func (t *Terminator) findWorstActiveCampaign(validCampaigns []CampaignPerformance) CampaignPerformance {
	// Sort campaigns by impressions (ascending)
	sort.Slice(validCampaigns, func(i, j int) bool {
		return validCampaigns[i].Impressions < validCampaigns[j].Impressions
	})
	
	// Return the campaign with the lowest impressions that still meets the threshold
	return validCampaigns[0]
}

// GetUnderperformingCampaigns identifies campaigns that are underperforming
// Campaigns are considered underperforming if their CPC is significantly higher
// than the median CPC of all valid campaigns
func (t *Terminator) GetUnderperformingCampaigns(campaigns []CampaignPerformance, cpcThresholdFactor float64) []string {
	if len(campaigns) == 0 {
		return []string{}
	}

	// Filter campaigns that meet the minimum impression threshold
	validCampaigns := t.filterValidCampaigns(campaigns)
	if len(validCampaigns) <= 1 {
		return []string{}
	}

	// Calculate median CPC from valid campaigns
	cpcValues := make([]float64, len(validCampaigns))
	for i, campaign := range validCampaigns {
		cpcValues[i] = campaign.CPC
	}
	medianCPC := calculateMedian(cpcValues)
	
	// Identify campaigns with CPC exceeding threshold
	underperforming := []string{}
	for _, campaign := range validCampaigns {
		// If CPC is significantly higher than median, consider it underperforming
		if campaign.CPC > (medianCPC * cpcThresholdFactor) {
			underperforming = append(underperforming, campaign.CampaignID)
		}
	}
	
	return underperforming
}

// calculateMedian calculates the median value of a slice of float64 values
func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Create a copy to avoid modifying the original slice
	valuesCopy := make([]float64, len(values))
	copy(valuesCopy, values)
	
	// Sort the values
	sort.Float64s(valuesCopy)
	
	// Calculate median
	middle := len(valuesCopy) / 2
	if len(valuesCopy)%2 == 0 {
		// Even number of elements, average the two middle values
		return (valuesCopy[middle-1] + valuesCopy[middle]) / 2
	}
	// Odd number of elements, return the middle value
	return valuesCopy[middle]
}