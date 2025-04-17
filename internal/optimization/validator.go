package optimization

import (
	"fmt"
	"time"

	"github.com/user/fb-ads/pkg/utils"
)

// ValidationThresholds defines minimum thresholds for valid campaign performance data
type ValidationThresholds struct {
	MinImpressions   int           `json:"min_impressions"`
	MinClicks        int           `json:"min_clicks"`
	MinRunningTime   time.Duration `json:"min_running_time"`
	MinDataPoints    int           `json:"min_data_points"`
	MinSpend         float64       `json:"min_spend"`
	EvaluationPeriod time.Duration `json:"evaluation_period"`
}

// DefaultValidationThresholds returns the default validation thresholds
func DefaultValidationThresholds() ValidationThresholds {
	return ValidationThresholds{
		MinImpressions:   1000,
		MinClicks:        10,
		MinRunningTime:   24 * time.Hour,
		MinDataPoints:    2,
		MinSpend:         1.0,
		EvaluationPeriod: 48 * time.Hour,
	}
}

// PerformanceValidator handles validation of campaign performance data
type PerformanceValidator struct {
	thresholds ValidationThresholds
}

// NewPerformanceValidator creates a new performance validator with default thresholds
func NewPerformanceValidator() *PerformanceValidator {
	return &PerformanceValidator{
		thresholds: DefaultValidationThresholds(),
	}
}

// SetThresholds sets custom validation thresholds
func (v *PerformanceValidator) SetThresholds(thresholds ValidationThresholds) {
	v.thresholds = thresholds
}

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	IsValid        bool              `json:"is_valid"`
	CampaignID     string            `json:"campaign_id"`
	Reasons        []string          `json:"reasons,omitempty"`
	Metrics        ValidationMetrics `json:"metrics"`
	EarliestData   time.Time         `json:"earliest_data,omitempty"`
	LatestData     time.Time         `json:"latest_data,omitempty"`
	RunningTime    time.Duration     `json:"running_time,omitempty"`
	DataPoints     int               `json:"data_points"`
	RecommendWait  bool              `json:"recommend_wait"`
	WaitTimeNeeded time.Duration     `json:"wait_time_needed,omitempty"`
}

// ValidationMetrics contains the actual metrics values being validated
type ValidationMetrics struct {
	TotalImpressions int     `json:"total_impressions"`
	TotalClicks      int     `json:"total_clicks"`
	TotalSpend       float64 `json:"total_spend"`
	CTR              float64 `json:"ctr"`
	CPC              float64 `json:"cpc"`
	CPM              float64 `json:"cpm"`
}

// ValidateCampaignData checks if the campaign has enough data for reliable optimization
func (v *PerformanceValidator) ValidateCampaignData(
	campaignID string,
	performances []utils.CampaignPerformance,
) ValidationResult {
	result := ValidationResult{
		IsValid:    true,
		CampaignID: campaignID,
		Reasons:    []string{},
		DataPoints: len(performances),
		Metrics: ValidationMetrics{
			TotalImpressions: 0,
			TotalClicks:      0,
			TotalSpend:       0,
			CTR:              0,
			CPC:              0,
			CPM:              0,
		},
	}

	// If no data, campaign is invalid
	if len(performances) == 0 {
		result.IsValid = false
		result.Reasons = append(result.Reasons, "No performance data available")
		return result
	}

	// Find earliest and latest data points to calculate runtime
	earliestTime := performances[0].LastUpdated
	latestTime := performances[0].LastUpdated

	// Accumulate metrics
	for _, perf := range performances {
		result.Metrics.TotalImpressions += perf.Impressions
		result.Metrics.TotalClicks += perf.Clicks
		result.Metrics.TotalSpend += perf.Spend

		// Update earliest/latest timestamps
		if perf.LastUpdated.Before(earliestTime) {
			earliestTime = perf.LastUpdated
		}
		if perf.LastUpdated.After(latestTime) {
			latestTime = perf.LastUpdated
		}
	}

	result.EarliestData = earliestTime
	result.LatestData = latestTime
	result.RunningTime = latestTime.Sub(earliestTime)

	// Calculate averages if we have data
	if result.Metrics.TotalImpressions > 0 {
		result.Metrics.CTR = float64(result.Metrics.TotalClicks) / float64(result.Metrics.TotalImpressions) * 100
		result.Metrics.CPM = result.Metrics.TotalSpend / float64(result.Metrics.TotalImpressions) * 1000
	}

	if result.Metrics.TotalClicks > 0 {
		result.Metrics.CPC = result.Metrics.TotalSpend / float64(result.Metrics.TotalClicks)
	}

	// Check minimum impressions
	if result.Metrics.TotalImpressions < v.thresholds.MinImpressions {
		result.IsValid = false
		result.Reasons = append(
			result.Reasons,
			fmt.Sprintf(
				"Insufficient impressions: %d (minimum required: %d)",
				result.Metrics.TotalImpressions,
				v.thresholds.MinImpressions,
			),
		)
	}

	// Check minimum clicks
	if result.Metrics.TotalClicks < v.thresholds.MinClicks {
		result.IsValid = false
		result.Reasons = append(
			result.Reasons,
			fmt.Sprintf(
				"Insufficient clicks: %d (minimum required: %d)",
				result.Metrics.TotalClicks,
				v.thresholds.MinClicks,
			),
		)
	}

	// Check minimum running time
	if result.RunningTime < v.thresholds.MinRunningTime {
		result.IsValid = false
		result.Reasons = append(
			result.Reasons,
			fmt.Sprintf(
				"Insufficient running time: %s (minimum required: %s)",
				result.RunningTime.String(),
				v.thresholds.MinRunningTime.String(),
			),
		)
	}

	// Check minimum data points
	if len(performances) < v.thresholds.MinDataPoints {
		result.IsValid = false
		result.Reasons = append(
			result.Reasons,
			fmt.Sprintf(
				"Insufficient data points: %d (minimum required: %d)",
				len(performances),
				v.thresholds.MinDataPoints,
			),
		)
	}

	// Check minimum spend
	if result.Metrics.TotalSpend < v.thresholds.MinSpend {
		result.IsValid = false
		result.Reasons = append(
			result.Reasons,
			fmt.Sprintf(
				"Insufficient spend: $%.2f (minimum required: $%.2f)",
				result.Metrics.TotalSpend,
				v.thresholds.MinSpend,
			),
		)
	}

	// Calculate wait recommendation
	if !result.IsValid && result.RunningTime < v.thresholds.EvaluationPeriod {
		result.RecommendWait = true
		result.WaitTimeNeeded = v.thresholds.EvaluationPeriod - result.RunningTime
	}

	return result
}

// ValidateCampaignsData checks if multiple campaigns have enough data for optimization
func (v *PerformanceValidator) ValidateCampaignsData(
	campaignPerformances map[string][]utils.CampaignPerformance,
) map[string]ValidationResult {
	results := make(map[string]ValidationResult)

	for campaignID, performances := range campaignPerformances {
		results[campaignID] = v.ValidateCampaignData(campaignID, performances)
	}

	return results
}