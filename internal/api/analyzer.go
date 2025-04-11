package api

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

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
}

// PerformanceAnalyzer handles analysis of campaign performance
type PerformanceAnalyzer struct {
	metricsCollector *MetricsCollector
}

// NewPerformanceAnalyzer creates a new performance analyzer
func NewPerformanceAnalyzer(metricsCollector *MetricsCollector) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		metricsCollector: metricsCollector,
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
		return (performances[i].Spend / float64(performances[i].Conversions)) > 
		       (performances[j].Spend / float64(performances[j].Conversions))
	})
	
	// Get worst 5 campaigns by CPA
	if len(performances) > 0 {
		numWorst := int(math.Min(5, float64(len(performances))))
		analysis.WorstCampaigns = performances[:numWorst]
	}
	
	// Generate recommendations
	analysis.Recommendations = p.generateRecommendations(performances, analysis)
	
	return analysis, nil
}

// GenerateReport generates a performance report in JSON format
func (p *PerformanceAnalyzer) GenerateReport(analysis *PerformanceAnalysis, filePath string) error {
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
	
	// Add general recommendations
	recommendations = append(recommendations, "Regularly update your creative assets to prevent ad fatigue")
	recommendations = append(recommendations, "Test different audience segments to identify the most responsive demographics")
	
	return recommendations
}