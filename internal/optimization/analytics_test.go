package optimization

import (
	"reflect"
	"testing"
)

func TestCalculatePerformanceMetrics(t *testing.T) {
	analyzer := NewAnalyzer(1000, 2.0)

	tests := []struct {
		name            string
		campaigns       []CampaignPerformance
		expectedMetrics PerformanceMetrics
	}{
		{
			name:      "empty campaigns",
			campaigns: []CampaignPerformance{},
			expectedMetrics: PerformanceMetrics{
				TotalImpressions:  0,
				TotalClicks:       0,
				TotalConversions:  0,
				TotalCost:         0,
				AverageCPM:        0,
				AverageCPC:        0,
				AverageCTR:        0,
				MedianCPM:         0,
				MedianCPC:         0,
				BestCTR:           0,
				WorstCTR:          0,
				AnomalyCampaigns:  []string{},
			},
		},
		{
			name: "no campaigns meet threshold",
			campaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 500, Clicks: 10, CPM: 5.0, CPC: 2.5, CTR: 2.0},
				{CampaignID: "2", Impressions: 800, Clicks: 20, CPM: 6.0, CPC: 2.4, CTR: 2.5},
			},
			expectedMetrics: PerformanceMetrics{
				TotalImpressions:  0,
				TotalClicks:       0,
				TotalConversions:  0,
				TotalCost:         0,
				AverageCPM:        0,
				AverageCPC:        0,
				AverageCTR:        0,
				MedianCPM:         0,
				MedianCPC:         0,
				BestCTR:           0,
				WorstCTR:          0,
				AnomalyCampaigns:  []string{},
			},
		},
		{
			name: "valid campaigns",
			campaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, Clicks: 24, CPM: 5.0, CPC: 2.5, CTR: 2.0, Cost: 6.0},
				{CampaignID: "2", Impressions: 1800, Clicks: 45, CPM: 6.0, CPC: 2.4, CTR: 2.5, Cost: 10.8},
				{CampaignID: "3", Impressions: 2400, Clicks: 60, CPM: 7.0, CPC: 2.8, CTR: 2.5, Cost: 16.8},
				{CampaignID: "4", Impressions: 800, Clicks: 16, CPM: 4.0, CPC: 2.0, CTR: 2.0, Cost: 3.2}, // Below threshold
			},
			expectedMetrics: PerformanceMetrics{
				TotalImpressions:  5400,  // Sum of valid campaigns only
				TotalClicks:       129,
				TotalConversions:  0,
				TotalCost:         33.6,
				AverageCPM:        6.0,   // (5.0 + 6.0 + 7.0) / 3
				AverageCPC:        2.57,  // (2.5 + 2.4 + 2.8) / 3
				AverageCTR:        2.33,  // (2.0 + 2.5 + 2.5) / 3
				MedianCPM:         6.0,   // Middle value of [5.0, 6.0, 7.0]
				MedianCPC:         2.5,   // Middle value of [2.4, 2.5, 2.8]
				BestCTR:           2.5,
				WorstCTR:          2.0,
				AnomalyCampaigns:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := analyzer.CalculatePerformanceMetrics(tt.campaigns)
			
			// Ignore TimeStamp in comparison
			metrics.TimeStamp = tt.expectedMetrics.TimeStamp
			
			// For floating point values, round to 2 decimal places
			metrics.AverageCPM = round(metrics.AverageCPM, 2)
			metrics.AverageCPC = round(metrics.AverageCPC, 2)
			metrics.AverageCTR = round(metrics.AverageCTR, 2)
			metrics.MedianCPM = round(metrics.MedianCPM, 2)
			metrics.MedianCPC = round(metrics.MedianCPC, 2)
			
			if !reflect.DeepEqual(metrics, tt.expectedMetrics) {
				t.Errorf("CalculatePerformanceMetrics() = %+v, want %+v", metrics, tt.expectedMetrics)
			}
		})
	}
}

func TestAnalyzeCampaign(t *testing.T) {
	analyzer := NewAnalyzer(1000, 2.0)

	tests := []struct {
		name           string
		campaign       CampaignPerformance
		allCampaigns   []CampaignPerformance
		expectedAction string
	}{
		{
			name: "single campaign",
			campaign: CampaignPerformance{
				CampaignID:  "1",
				Impressions: 1200,
				Clicks:      24,
				CPM:         5.0,
				CPC:         2.5,
				CTR:         2.0,
			},
			allCampaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, Clicks: 24, CPM: 5.0, CPC: 2.5, CTR: 2.0},
			},
			expectedAction: "maintain",
		},
		{
			name: "below impression threshold",
			campaign: CampaignPerformance{
				CampaignID:  "1",
				Impressions: 500, // Below threshold
				Clicks:      10,
				CPM:         5.0,
				CPC:         2.5,
				CTR:         2.0,
			},
			allCampaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 500, Clicks: 10, CPM: 5.0, CPC: 2.5, CTR: 2.0},
				{CampaignID: "2", Impressions: 1200, Clicks: 24, CPM: 6.0, CPC: 3.0, CTR: 2.0},
			},
			expectedAction: "wait_for_data",
		},
		{
			name: "top performer",
			campaign: CampaignPerformance{
				CampaignID:  "1",
				Impressions: 1200,
				Clicks:      30,
				CPM:         5.0,
				CPC:         2.0, // Lowest CPC in group
				CTR:         2.5,
			},
			allCampaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, Clicks: 30, CPM: 5.0, CPC: 2.0, CTR: 2.5},
				{CampaignID: "2", Impressions: 1200, Clicks: 24, CPM: 6.0, CPC: 3.0, CTR: 2.0},
				{CampaignID: "3", Impressions: 1200, Clicks: 20, CPM: 7.0, CPC: 4.2, CTR: 1.7},
				{CampaignID: "4", Impressions: 1200, Clicks: 18, CPM: 6.5, CPC: 4.3, CTR: 1.5},
			},
			expectedAction: "increase_budget",
		},
		{
			name: "poor performer",
			campaign: CampaignPerformance{
				CampaignID:  "4",
				Impressions: 1200,
				Clicks:      18,
				CPM:         6.5,
				CPC:         4.3, // Highest CPC in group
				CTR:         1.5,
			},
			allCampaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, Clicks: 30, CPM: 5.0, CPC: 2.0, CTR: 2.5},
				{CampaignID: "2", Impressions: 1200, Clicks: 24, CPM: 6.0, CPC: 3.0, CTR: 2.0},
				{CampaignID: "3", Impressions: 1200, Clicks: 20, CPM: 7.0, CPC: 4.2, CTR: 1.7},
				{CampaignID: "4", Impressions: 1200, Clicks: 18, CPM: 6.5, CPC: 4.3, CTR: 1.5},
			},
			expectedAction: "terminate",
		},
		{
			name: "above reference CPC",
			campaign: CampaignPerformance{
				CampaignID:  "3",
				Impressions: 1200,
				Clicks:      20,
				CPM:         7.0,
				CPC:         2.5, // Above reference CPC of 2.0
				CTR:         1.7,
			},
			allCampaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, Clicks: 30, CPM: 5.0, CPC: 1.5, CTR: 2.5},
				{CampaignID: "2", Impressions: 1200, Clicks: 24, CPM: 6.0, CPC: 1.8, CTR: 2.0},
				{CampaignID: "3", Impressions: 1200, Clicks: 20, CPM: 7.0, CPC: 2.5, CTR: 1.7},
			},
			expectedAction: "optimize_creative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analytics := analyzer.AnalyzeCampaign(tt.campaign, tt.allCampaigns)
			
			if analytics.RecommendedAction != tt.expectedAction {
				t.Errorf("AnalyzeCampaign() recommended action = %v, want %v", 
					analytics.RecommendedAction, tt.expectedAction)
			}
		})
	}
}

func TestSortCampaignsByPerformance(t *testing.T) {
	analyzer := NewAnalyzer(1000, 2.0)

	campaigns := []CampaignPerformance{
		{CampaignID: "1", Impressions: 1200, CPC: 2.5},
		{CampaignID: "2", Impressions: 800, CPC: 2.0},  // Below threshold
		{CampaignID: "3", Impressions: 1500, CPC: 1.8}, // Best valid CPC
		{CampaignID: "4", Impressions: 1300, CPC: 3.0}, // Worst valid CPC
	}

	expected := []string{"3", "1", "4", "2"}

	sortedCampaigns := analyzer.SortCampaignsByPerformance(campaigns)
	
	// Extract campaign IDs
	sortedIDs := make([]string, len(sortedCampaigns))
	for i, campaign := range sortedCampaigns {
		sortedIDs[i] = campaign.CampaignID
	}
	
	if !reflect.DeepEqual(sortedIDs, expected) {
		t.Errorf("SortCampaignsByPerformance() = %v, want %v", sortedIDs, expected)
	}
}

// Helper function to round float64 to specified decimal places
func round(val float64, places int) float64 {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= 0.5 {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	return round / pow
}