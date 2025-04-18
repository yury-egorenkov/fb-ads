package optimization

import (
	"math"
	"sort"
	"time"
)

// PerformanceMetrics represents the aggregated performance metrics for a set of campaigns
type PerformanceMetrics struct {
	TotalImpressions  int
	TotalClicks       int
	TotalConversions  int
	TotalCost         float64
	AverageCPM        float64
	AverageCPC        float64
	AverageCTR        float64
	MedianCPM         float64
	MedianCPC         float64
	BestCTR           float64
	WorstCTR          float64
	AnomalyCampaigns  []string    // Campaigns with abnormal performance
	TimeStamp         time.Time   // When the metrics were calculated
}

// CampaignAnalytics represents the analytics data for a specific campaign
type CampaignAnalytics struct {
	CampaignID         string
	Impressions        int
	Clicks             int
	Conversions        int
	Cost               float64
	CPM                float64
	CPC                float64
	CTR                float64
	PerformanceScore   float64   // Normalized score (0-100) comparing to other campaigns
	RecommendedAction  string    // "increase_budget", "decrease_budget", "terminate", "maintain"
	AnomalyScore       float64   // How much this campaign deviates from the norm
	IsAnomaly          bool      // Whether this campaign is considered an anomaly
}

// Analyzer provides methods for analyzing campaign performance
type Analyzer struct {
	statAnalyzer  *StatisticalAnalyzer
	minImpressions int
	referenceCPC  float64   // Benchmark CPC to compare against
}

// NewAnalyzer creates a new instance of Analyzer
func NewAnalyzer(minImpressions int, referenceCPC float64) *Analyzer {
	return &Analyzer{
		statAnalyzer:  NewStatisticalAnalyzer(),
		minImpressions: minImpressions,
		referenceCPC:  referenceCPC,
	}
}

// CalculatePerformanceMetrics calculates aggregated performance metrics
func (a *Analyzer) CalculatePerformanceMetrics(campaigns []CampaignPerformance) PerformanceMetrics {
	if len(campaigns) == 0 {
		return PerformanceMetrics{
			TimeStamp: time.Now(),
		}
	}

	// Filter campaigns with minimum required impressions
	validCampaigns := []CampaignPerformance{}
	for _, campaign := range campaigns {
		if campaign.Impressions >= a.minImpressions {
			validCampaigns = append(validCampaigns, campaign)
		}
	}
	
	if len(validCampaigns) == 0 {
		return PerformanceMetrics{
			TimeStamp: time.Now(),
		}
	}

	// Initialize metrics
	metrics := PerformanceMetrics{
		TimeStamp: time.Now(),
	}

	// Collect values for calculation
	var totalImpressions, totalClicks, totalConversions int
	var totalCost, totalCPM, totalCPC, totalCTR float64
	cpmValues := make([]float64, len(validCampaigns))
	cpcValues := make([]float64, len(validCampaigns))
	ctrValues := make([]float64, len(validCampaigns))
	
	// Track best and worst CTR
	bestCTR := -1.0
	worstCTR := 101.0  // CTR is percentage, so this is well above any valid value
	
	for i, campaign := range validCampaigns {
		totalImpressions += campaign.Impressions
		totalClicks += campaign.Clicks
		totalConversions += campaign.Conversions
		totalCost += campaign.Cost
		totalCPM += campaign.CPM
		totalCPC += campaign.CPC
		totalCTR += campaign.CTR
		
		cpmValues[i] = campaign.CPM
		cpcValues[i] = campaign.CPC
		ctrValues[i] = campaign.CTR
		
		// Track best and worst CTR
		if campaign.CTR > bestCTR {
			bestCTR = campaign.CTR
		}
		if campaign.CTR < worstCTR {
			worstCTR = campaign.CTR
		}
	}
	
	// Calculate averages
	campaignCount := float64(len(validCampaigns))
	metrics.TotalImpressions = totalImpressions
	metrics.TotalClicks = totalClicks
	metrics.TotalConversions = totalConversions
	metrics.TotalCost = totalCost
	metrics.AverageCPM = totalCPM / campaignCount
	metrics.AverageCPC = totalCPC / campaignCount
	metrics.AverageCTR = totalCTR / campaignCount
	metrics.BestCTR = bestCTR
	metrics.WorstCTR = worstCTR
	
	// Calculate medians
	metrics.MedianCPM = calculateMedian(cpmValues)
	metrics.MedianCPC = calculateMedian(cpcValues)
	
	// Find anomalies
	metrics.AnomalyCampaigns = a.findAnomalies(validCampaigns)
	
	return metrics
}

// findAnomalies identifies campaigns with abnormal performance
func (a *Analyzer) findAnomalies(campaigns []CampaignPerformance) []string {
	if len(campaigns) <= 1 {
		return []string{}
	}
	
	// Extract CPC values to check for outliers
	cpcValues := make([]float64, len(campaigns))
	for i, campaign := range campaigns {
		cpcValues[i] = campaign.CPC
	}
	
	// Find campaigns with CPC outliers (> 2 standard deviations from mean)
	anomalies := []string{}
	for _, campaign := range campaigns {
		if a.statAnalyzer.IsOutlier(campaign.CPC, cpcValues) {
			anomalies = append(anomalies, campaign.CampaignID)
		}
	}
	
	return anomalies
}

// AnalyzeCampaign generates analytics for a specific campaign
func (a *Analyzer) AnalyzeCampaign(
	campaign CampaignPerformance,
	allCampaigns []CampaignPerformance,
) CampaignAnalytics {
	// Basic analytics data from campaign performance
	analytics := CampaignAnalytics{
		CampaignID:  campaign.CampaignID,
		Impressions: campaign.Impressions,
		Clicks:      campaign.Clicks,
		Conversions: campaign.Conversions,
		Cost:        campaign.Cost,
		CPM:         campaign.CPM,
		CPC:         campaign.CPC,
		CTR:         campaign.CTR,
	}
	
	// If there are no other campaigns to compare with, return basic analytics
	if len(allCampaigns) <= 1 {
		analytics.PerformanceScore = 50.0 // Neutral score
		analytics.RecommendedAction = "maintain"
		return analytics
	}
	
	// Extract CPC values for comparison
	cpcValues := make([]float64, 0, len(allCampaigns))
	for _, c := range allCampaigns {
		// Only include campaigns with sufficient impressions
		if c.Impressions >= a.minImpressions {
			cpcValues = append(cpcValues, c.CPC)
		}
	}
	
	// Check if campaign CPC is an outlier
	analytics.IsAnomaly = a.statAnalyzer.IsOutlier(campaign.CPC, cpcValues)
	
	// Calculate anomaly score (how many standard deviations from mean)
	mean := a.statAnalyzer.CalculateMean(cpcValues)
	stdDev := a.statAnalyzer.CalculateStandardDeviation(cpcValues)
	
	if stdDev > 0 {
		analytics.AnomalyScore = math.Abs(campaign.CPC - mean) / stdDev
	} else {
		analytics.AnomalyScore = 0
	}
	
	// Calculate performance score (0-100)
	// Lower CPC is better, so invert the relationship
	lowestCPC := math.MaxFloat64
	highestCPC := 0.0
	
	for _, cpc := range cpcValues {
		if cpc < lowestCPC {
			lowestCPC = cpc
		}
		if cpc > highestCPC {
			highestCPC = cpc
		}
	}
	
	cpcRange := highestCPC - lowestCPC
	if cpcRange > 0 {
		// Invert so lower CPC = higher score
		analytics.PerformanceScore = 100 * (1 - ((campaign.CPC - lowestCPC) / cpcRange))
	} else {
		analytics.PerformanceScore = 50.0 // Default to neutral if all CPCs are identical
	}
	
	// Determine recommended action
	analytics.RecommendedAction = a.determineRecommendedAction(analytics, mean)
	
	return analytics
}

// determineRecommendedAction recommends an action based on campaign analytics
func (a *Analyzer) determineRecommendedAction(
	analytics CampaignAnalytics,
	averageCPC float64,
) string {
	// Check if impressions are too low
	if analytics.Impressions < a.minImpressions {
		return "wait_for_data"
	}
	
	// If the campaign is performing exceptionally well (top 10% score)
	if analytics.PerformanceScore >= 90 {
		return "increase_budget"
	}
	
	// If the campaign is performing poorly (bottom 20% score)
	if analytics.PerformanceScore <= 20 {
		return "terminate"
	}
	
	// If the campaign is an anomaly with high CPC
	if analytics.IsAnomaly && analytics.CPC > averageCPC {
		return "decrease_budget"
	}
	
	// If the CPC is higher than reference CPC (benchmark)
	if analytics.CPC > a.referenceCPC * 1.2 {
		return "optimize_creative"
	}
	
	// Default recommendation for average performing campaigns
	return "maintain"
}

// SortCampaignsByPerformance sorts campaigns by their performance (best to worst)
func (a *Analyzer) SortCampaignsByPerformance(campaigns []CampaignPerformance) []CampaignPerformance {
	// Create a copy to avoid modifying the original
	sortedCampaigns := make([]CampaignPerformance, len(campaigns))
	copy(sortedCampaigns, campaigns)
	
	// Sort by CPC (ascending, lower is better)
	sort.Slice(sortedCampaigns, func(i, j int) bool {
		// Only consider campaigns with sufficient impressions
		iValid := sortedCampaigns[i].Impressions >= a.minImpressions
		jValid := sortedCampaigns[j].Impressions >= a.minImpressions
		
		// Invalid campaigns are worse than valid ones
		if iValid && !jValid {
			return true
		}
		if !iValid && jValid {
			return false
		}
		if !iValid && !jValid {
			// If both invalid, sort by impressions (higher is better)
			return sortedCampaigns[i].Impressions > sortedCampaigns[j].Impressions
		}
		
		// Both valid, sort by CPC (lower is better)
		return sortedCampaigns[i].CPC < sortedCampaigns[j].CPC
	})
	
	return sortedCampaigns
}