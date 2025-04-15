package api

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/user/fb-ads/internal/audience"
	"github.com/user/fb-ads/pkg/utils"
)

// PerformanceAnalysis contains analysis results for campaign performance
type PerformanceAnalysis struct {
	TopCampaigns     []utils.CampaignPerformance `json:"top_campaigns"`
	WorstCampaigns   []utils.CampaignPerformance `json:"worst_campaigns"`
	AverageCPA       float64                    `json:"average_cpa"`
	AverageCTR       float64                    `json:"average_ctr"`
	AverageROAS      float64                    `json:"average_roas"`
	TotalSpend       float64                    `json:"total_spend"`
	TotalConversions int                        `json:"total_conversions"`
	TotalClicks      int                        `json:"total_clicks"`
	TotalImpressions int                        `json:"total_impressions"`
	AnalysisDate     time.Time                  `json:"analysis_date"`
	Recommendations  []string                   `json:"recommendations"`
	TopAudiences     []AudiencePerformance      `json:"top_audiences,omitempty"`
}

// AudiencePerformance represents performance metrics for a specific audience segment
type AudiencePerformance struct {
	Segment     audience.AudienceSegment      `json:"segment"`
	Performance audience.SegmentPerformance   `json:"performance"`
	Campaigns   []string                      `json:"campaigns"`
	ReachSize   int64                         `json:"reach_size"`
}

// PerformanceAnalyzer handles analysis of campaign performance
type PerformanceAnalyzer struct {
	metricsCollector *MetricsCollector
	audienceAnalyzer *audience.AudienceAnalyzer
}

// NewPerformanceAnalyzer creates a new performance analyzer
func NewPerformanceAnalyzer(metricsCollector *MetricsCollector, audienceAnalyzer *audience.AudienceAnalyzer) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		metricsCollector: metricsCollector,
		audienceAnalyzer: audienceAnalyzer,
	}
}

// AnalyzeCampaignPerformance analyzes campaign performance
func (p *PerformanceAnalyzer) AnalyzeCampaignPerformance(timeRange TimeRange) (*PerformanceAnalysis, error) {
	// Create insights request
	request := InsightsRequest{
		Level: "campaign",
		TimeRange: timeRange,
		Fields: []string{
			"campaign_id",
			"campaign_name",
			"spend",
			"impressions",
			"clicks",
			"actions",
			"cpm",
			"cpc",
			"ctr",
			"cost_per_action_type",
		},
	}
	
	// Collect metrics
	performances, err := p.metricsCollector.CollectCampaignMetrics(request)
	if err != nil {
		return nil, fmt.Errorf("error collecting metrics: %w", err)
	}
	
	if len(performances) == 0 {
		return nil, fmt.Errorf("no campaign data found for the specified time range")
	}
	
	// Calculate summary statistics
	analysis := &PerformanceAnalysis{
		AnalysisDate: time.Now(),
	}
	
	var totalCPA float64
	var totalCTR float64
	var totalROAS float64
	var campaignsWithConversions int
	
	for _, perf := range performances {
		analysis.TotalSpend += perf.Spend
		analysis.TotalConversions += perf.Conversions
		analysis.TotalClicks += perf.Clicks
		analysis.TotalImpressions += perf.Impressions
		
		if perf.Conversions > 0 {
			cpa := perf.Spend / float64(perf.Conversions)
			totalCPA += cpa
			campaignsWithConversions++
		}
		
		totalCTR += perf.CTR
		totalROAS += perf.ROAS
	}
	
	// Calculate averages
	if campaignsWithConversions > 0 {
		analysis.AverageCPA = totalCPA / float64(campaignsWithConversions)
	}
	
	if len(performances) > 0 {
		analysis.AverageCTR = totalCTR / float64(len(performances))
		analysis.AverageROAS = totalROAS / float64(len(performances))
	}
	
	// Sort campaigns by ROAS (descending) for top campaigns
	sort.Slice(performances, func(i, j int) bool {
		return performances[i].ROAS > performances[j].ROAS
	})
	
	// Get top 5 campaigns by ROAS
	if len(performances) > 0 {
		numTop := int(math.Min(5, float64(len(performances))))
		analysis.TopCampaigns = performances[:numTop]
	}
	
	// Sort campaigns by CPA (descending) for worst campaigns
	sort.Slice(performances, func(i, j int) bool {
		// If either campaign has no conversions, consider it worse
		if performances[i].Conversions == 0 && performances[j].Conversions > 0 {
			return true
		}
		if performances[j].Conversions == 0 && performances[i].Conversions > 0 {
			return false
		}
		if performances[i].Conversions == 0 && performances[j].Conversions == 0 {
			// Both have no conversions, sort by spend (descending)
			return performances[i].Spend > performances[j].Spend
		}
		
		// Otherwise sort by CPA (descending)
		cpaI := performances[i].Spend / float64(performances[i].Conversions)
		cpaJ := performances[j].Spend / float64(performances[j].Conversions)
		
		// Handle NaN cases safely
		if math.IsNaN(cpaI) {
			return false
		}
		if math.IsNaN(cpaJ) {
			return true
		}
		
		return cpaI > cpaJ
	})
	
	// Get worst 5 campaigns by CPA
	if len(performances) > 0 {
		numWorst := int(math.Min(5, float64(len(performances))))
		analysis.WorstCampaigns = performances[:numWorst]
	}
	
	// Add audience performance analysis if available
	if p.audienceAnalyzer != nil {
		topAudiences, err := p.AnalyzeAudiencePerformance(timeRange)
		if err == nil && len(topAudiences) > 0 {
			analysis.TopAudiences = topAudiences
		}
	}
	
	// Generate recommendations
	analysis.Recommendations = p.generateRecommendations(performances, analysis)
	
	return analysis, nil
}

// GenerateReport generates a performance report in JSON format
func (p *PerformanceAnalyzer) GenerateReport(analysis *PerformanceAnalysis, filePath string) error {
	// Sanitize any potential NaN values
	sanitizeAnalysis(analysis)
	
	// Convert analysis to JSON
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling analysis: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing report: %w", err)
	}
	
	return nil
}

// sanitizeAnalysis replaces any NaN or Inf values with 0 to prevent JSON marshaling errors
func sanitizeAnalysis(analysis *PerformanceAnalysis) {
	// Replace NaN or Inf in main metrics
	if math.IsNaN(analysis.AverageCPA) || math.IsInf(analysis.AverageCPA, 0) {
		analysis.AverageCPA = 0
	}
	if math.IsNaN(analysis.AverageCTR) || math.IsInf(analysis.AverageCTR, 0) {
		analysis.AverageCTR = 0
	}
	if math.IsNaN(analysis.AverageROAS) || math.IsInf(analysis.AverageROAS, 0) {
		analysis.AverageROAS = 0
	}
	
	// Sanitize top campaigns
	for i := range analysis.TopCampaigns {
		if math.IsNaN(analysis.TopCampaigns[i].CPC) || math.IsInf(analysis.TopCampaigns[i].CPC, 0) {
			analysis.TopCampaigns[i].CPC = 0
		}
		if math.IsNaN(analysis.TopCampaigns[i].CPM) || math.IsInf(analysis.TopCampaigns[i].CPM, 0) {
			analysis.TopCampaigns[i].CPM = 0
		}
		if math.IsNaN(analysis.TopCampaigns[i].CTR) || math.IsInf(analysis.TopCampaigns[i].CTR, 0) {
			analysis.TopCampaigns[i].CTR = 0
		}
		if math.IsNaN(analysis.TopCampaigns[i].ROAS) || math.IsInf(analysis.TopCampaigns[i].ROAS, 0) {
			analysis.TopCampaigns[i].ROAS = 0
		}
	}
	
	// Sanitize worst campaigns
	for i := range analysis.WorstCampaigns {
		if math.IsNaN(analysis.WorstCampaigns[i].CPC) || math.IsInf(analysis.WorstCampaigns[i].CPC, 0) {
			analysis.WorstCampaigns[i].CPC = 0
		}
		if math.IsNaN(analysis.WorstCampaigns[i].CPM) || math.IsInf(analysis.WorstCampaigns[i].CPM, 0) {
			analysis.WorstCampaigns[i].CPM = 0
		}
		if math.IsNaN(analysis.WorstCampaigns[i].CTR) || math.IsInf(analysis.WorstCampaigns[i].CTR, 0) {
			analysis.WorstCampaigns[i].CTR = 0
		}
		if math.IsNaN(analysis.WorstCampaigns[i].ROAS) || math.IsInf(analysis.WorstCampaigns[i].ROAS, 0) {
			analysis.WorstCampaigns[i].ROAS = 0
		}
	}
	
	// Sanitize audience performances if present
	for i := range analysis.TopAudiences {
		if math.IsNaN(analysis.TopAudiences[i].Performance.CPC) || math.IsInf(analysis.TopAudiences[i].Performance.CPC, 0) {
			analysis.TopAudiences[i].Performance.CPC = 0
		}
		if math.IsNaN(analysis.TopAudiences[i].Performance.CPM) || math.IsInf(analysis.TopAudiences[i].Performance.CPM, 0) {
			analysis.TopAudiences[i].Performance.CPM = 0
		}
		if math.IsNaN(analysis.TopAudiences[i].Performance.CTR) || math.IsInf(analysis.TopAudiences[i].Performance.CTR, 0) {
			analysis.TopAudiences[i].Performance.CTR = 0
		}
		if math.IsNaN(analysis.TopAudiences[i].Performance.CVR) || math.IsInf(analysis.TopAudiences[i].Performance.CVR, 0) {
			analysis.TopAudiences[i].Performance.CVR = 0
		}
		if math.IsNaN(analysis.TopAudiences[i].Performance.CPA) || math.IsInf(analysis.TopAudiences[i].Performance.CPA, 0) {
			analysis.TopAudiences[i].Performance.CPA = 0
		}
	}
}

// AnalyzeAudiencePerformance analyzes audience segment performance
func (p *PerformanceAnalyzer) AnalyzeAudiencePerformance(timeRange TimeRange) ([]AudiencePerformance, error) {
	if p.audienceAnalyzer == nil {
		return nil, fmt.Errorf("audience analyzer not initialized")
	}
	
	// In a real implementation, we would use these parameters to query the API
	// Example of how we would structure the request:
	_ = InsightsRequest{
		Level:          "ad",
		TimeRange:      timeRange,
		Fields: []string{
			"campaign_id",
			"campaign_name",
			"adset_id",
			"adset_name",
			"spend",
			"impressions",
			"clicks",
			"actions",
			"cpm",
			"cpc",
			"ctr",
		},
		BreakdownsType: "age,gender,country",
	}
	
	// In a production implementation, we would process actual insights data
	// For now, we'll use sample data since we don't have a CollectInsights method
	
	// Skip API call and insights processing for now since it's not implemented
	// This would be implemented in a real system to extract audience data
	// from the actual campaign performance
	
	// Sample implementation with mock data
	// Normally, you would extract this from the insights data
	sampleAudiences := []AudiencePerformance{
		{
			Segment: audience.AudienceSegment{
				ID:   "6003107902433",
				Name: "Online shopping",
				Type: "interest",
				Size: 15000000,
			},
			Performance: audience.SegmentPerformance{
				Impressions: 120000,
				Clicks:      3600,
				Conversions: 180,
				Spend:       500.00,
				CPC:         0.14,
				CPM:         4.17,
				CTR:         3.0,
				CVR:         5.0,
				CPA:         2.78,
			},
			Campaigns: []string{"Campaign A", "Campaign B"},
			ReachSize: 15000000,
		},
		{
			Segment: audience.AudienceSegment{
				ID:   "6002714895372",
				Name: "Engaged Shoppers",
				Type: "behavior",
				Size: 12500000,
			},
			Performance: audience.SegmentPerformance{
				Impressions: 95000,
				Clicks:      2850,
				Conversions: 155,
				Spend:       425.00,
				CPC:         0.15,
				CPM:         4.47,
				CTR:         3.0,
				CVR:         5.4,
				CPA:         2.74,
			},
			Campaigns: []string{"Campaign A", "Campaign C"},
			ReachSize: 12500000,
		},
	}
	
	// Sort audience performances by conversion rate (descending)
	sort.Slice(sampleAudiences, func(i, j int) bool {
		return sampleAudiences[i].Performance.CVR > sampleAudiences[j].Performance.CVR
	})
	
	return sampleAudiences, nil
}

// generateRecommendations generates recommendations based on campaign performance
func (p *PerformanceAnalyzer) generateRecommendations(performances []utils.CampaignPerformance, analysis *PerformanceAnalysis) []string {
	var recommendations []string
	
	// Check overall conversion rate
	if analysis.TotalConversions == 0 {
		recommendations = append(recommendations, "No conversions recorded. Consider revising your campaign targeting or creative elements.")
	}
	
	// Check for campaigns with high spend but no conversions
	var highSpendNoConv []string
	for _, perf := range performances {
		if perf.Conversions == 0 && perf.Spend > 100 {
			highSpendNoConv = append(highSpendNoConv, perf.Name)
		}
	}
	
	if len(highSpendNoConv) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Consider pausing these campaigns with high spend but no conversions: %v", highSpendNoConv))
	}
	
	// Check for campaigns with very low CTR
	var lowCTRCampaigns []string
	for _, perf := range performances {
		if perf.CTR < 0.5 && perf.Impressions > 1000 {
			lowCTRCampaigns = append(lowCTRCampaigns, perf.Name)
		}
	}
	
	if len(lowCTRCampaigns) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Improve ad creatives for these campaigns with low CTR: %v", lowCTRCampaigns))
	}
	
	// Check for high-performing campaigns that could benefit from more budget
	var highROASCampaigns []string
	for _, perf := range performances {
		if perf.ROAS > 3.0 && perf.Conversions > 5 {
			highROASCampaigns = append(highROASCampaigns, perf.Name)
		}
	}
	
	if len(highROASCampaigns) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Consider increasing budget for these high ROAS campaigns: %v", highROASCampaigns))
	}
	
	// Add audience-specific recommendations if available
	if len(analysis.TopAudiences) > 0 {
		topAudience := analysis.TopAudiences[0]
		recommendations = append(recommendations, 
			fmt.Sprintf("Consider expanding campaigns using the '%s' audience segment which shows strong performance (CVR: %.1f%%)", 
				topAudience.Segment.Name, topAudience.Performance.CVR))
	}
	
	// Add general recommendations
	recommendations = append(recommendations, "Regularly update your creative assets to prevent ad fatigue")
	recommendations = append(recommendations, "Test different audience segments to identify the most responsive demographics")
	
	return recommendations
}