package optimization

import (
	"reflect"
	"testing"
	"time"
)

func TestCalculateNewCPM(t *testing.T) {
	// Create adjuster with 10% increment and 5% decrement
	adjuster := NewAdjuster(15.0, 1.0, 10.0, 5.0, 48)

	tests := []struct {
		name       string
		campaign   CampaignPerformance
		optimalCPM float64
		expected   float64
	}{
		{
			name: "significantly below optimal",
			campaign: CampaignPerformance{
				CampaignID:  "1",
				CPM:         5.0, // Well below 10.0 * 0.8 = 8.0
				Impressions: 1000,
			},
			optimalCPM: 10.0,
			expected:   5.5, // 5.0 * (1 + 10/100) = 5.5
		},
		{
			name: "significantly above optimal",
			campaign: CampaignPerformance{
				CampaignID:  "2",
				CPM:         13.0, // Above 10.0 * 1.2 = 12.0
				Impressions: 1000,
			},
			optimalCPM: 10.0,
			expected:   12.35, // 13.0 * (1 - 5/100) = 12.35
		},
		{
			name: "slightly below optimal",
			campaign: CampaignPerformance{
				CampaignID:  "3",
				CPM:         9.0, // Between 8.0 and 10.0
				Impressions: 1000,
			},
			optimalCPM: 10.0,
			expected:   9.225, // 9.0 * (1 + (10/2)/100) = 9.225
		},
		{
			name: "slightly above optimal",
			campaign: CampaignPerformance{
				CampaignID:  "4",
				CPM:         11.0, // Between 10.0 and 12.0
				Impressions: 1000,
			},
			optimalCPM: 10.0,
			expected:   10.725, // 11.0 * (1 - (5/2)/100) = 10.725
		},
		{
			name: "at optimal",
			campaign: CampaignPerformance{
				CampaignID:  "5",
				CPM:         10.0, // Equal to optimal
				Impressions: 1000,
			},
			optimalCPM: 10.0,
			expected:   10.0, // No change
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adjuster.calculateNewCPM(tt.campaign, tt.optimalCPM)
			
			// Compare with a small delta for floating point precision
			if !almostEqual(result, tt.expected, 0.001) {
				t.Errorf("calculateNewCPM() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateAdjustments(t *testing.T) {
	// Create adjuster with max CPM 15.0, min CPM 1.0, 10% increment, 5% decrement, 48h wait
	adjuster := NewAdjuster(15.0, 1.0, 10.0, 5.0, 48)
	
	// Current time for testing
	now := time.Now()
	
	// Time in the past, beyond 48h threshold
	pastTime := now.Add(-72 * time.Hour)
	
	// Recent time, within 48h threshold
	recentTime := now.Add(-24 * time.Hour)

	tests := []struct {
		name                string
		campaigns           []CampaignPerformance
		previousAdjustments []CampaignAdjustment
		expectedCount       int      // Number of adjustments expected
		expectedIDs         []string // Campaign IDs expected to be adjusted
	}{
		{
			name:                "empty campaigns",
			campaigns:           []CampaignPerformance{},
			previousAdjustments: []CampaignAdjustment{},
			expectedCount:       0,
			expectedIDs:         []string{},
		},
		{
			name: "no previous adjustments",
			campaigns: []CampaignPerformance{
				{CampaignID: "1", CPM: 5.0, Impressions: 1000},
				{CampaignID: "2", CPM: 8.0, Impressions: 1000},
			},
			previousAdjustments: []CampaignAdjustment{},
			expectedCount:       2,
			expectedIDs:         []string{"1", "2"},
		},
		{
			name: "some campaigns recently adjusted",
			campaigns: []CampaignPerformance{
				{CampaignID: "1", CPM: 5.0, Impressions: 1000},
				{CampaignID: "2", CPM: 8.0, Impressions: 1000},
				{CampaignID: "3", CPM: 10.0, Impressions: 1000},
			},
			previousAdjustments: []CampaignAdjustment{
				{CampaignID: "1", CurrentCPM: 4.5, AdjustedCPM: 5.0, AdjustmentTS: recentTime}, // Too recent
				{CampaignID: "2", CurrentCPM: 7.5, AdjustedCPM: 8.0, AdjustmentTS: pastTime},   // Old enough
			},
			expectedCount: 3,
			expectedIDs:   []string{"1", "2", "3"}, // All should be present, but ID 1 should not be adjusted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adjuster.CalculateAdjustments(tt.campaigns, tt.previousAdjustments)
			
			// Check count
			if len(result) != tt.expectedCount {
				t.Errorf("CalculateAdjustments() returned %v adjustments, want %v", len(result), tt.expectedCount)
			}
			
			// Extract campaign IDs from result
			resultIDs := make([]string, len(result))
			for i, adj := range result {
				resultIDs[i] = adj.CampaignID
			}
			
			// Check campaign IDs
			if !reflect.DeepEqual(resultIDs, tt.expectedIDs) {
				t.Errorf("CalculateAdjustments() returned campaign IDs %v, want %v", resultIDs, tt.expectedIDs)
			}
			
			// For campaigns that were recently adjusted, check that the CPM didn't change
			if len(tt.previousAdjustments) > 0 {
				for _, adj := range result {
					for _, prevAdj := range tt.previousAdjustments {
						// If it's a campaign that was recently adjusted
						if adj.CampaignID == prevAdj.CampaignID && prevAdj.AdjustmentTS == recentTime {
							// Make sure the CPM didn't change
							if adj.CurrentCPM != adj.AdjustedCPM {
								t.Errorf("Recently adjusted campaign %s had CPM changed from %v to %v", 
									adj.CampaignID, adj.CurrentCPM, adj.AdjustedCPM)
							}
							
							// Make sure the adjustment timestamp didn't change
							if adj.AdjustmentTS != prevAdj.AdjustmentTS {
								t.Errorf("Recently adjusted campaign %s had timestamp changed", adj.CampaignID)
							}
						}
					}
				}
			}
		})
	}
}

func TestGetEligibleCampaigns(t *testing.T) {
	// Create adjuster with 48h wait time
	adjuster := NewAdjuster(15.0, 1.0, 10.0, 5.0, 48)
	
	// Current time for testing
	now := time.Now()
	
	// Time in the past, beyond 48h threshold
	pastTime := now.Add(-72 * time.Hour)
	
	// Recent time, within 48h threshold
	recentTime := now.Add(-24 * time.Hour)

	tests := []struct {
		name                string
		campaignIDs         []string
		previousAdjustments []CampaignAdjustment
		expected            []string
	}{
		{
			name:                "empty campaigns",
			campaignIDs:         []string{},
			previousAdjustments: []CampaignAdjustment{},
			expected:            []string{},
		},
		{
			name:                "no previous adjustments",
			campaignIDs:         []string{"1", "2", "3"},
			previousAdjustments: []CampaignAdjustment{},
			expected:            []string{"1", "2", "3"}, // All eligible
		},
		{
			name:        "mixed eligibility",
			campaignIDs: []string{"1", "2", "3", "4"},
			previousAdjustments: []CampaignAdjustment{
				{CampaignID: "1", AdjustmentTS: recentTime}, // Too recent
				{CampaignID: "2", AdjustmentTS: pastTime},   // Old enough
				// Campaign 3 has no previous adjustment
				{CampaignID: "4", AdjustmentTS: recentTime}, // Too recent
			},
			expected: []string{"2", "3"}, // Only 2 and 3 are eligible
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adjuster.GetEligibleCampaigns(tt.campaignIDs, tt.previousAdjustments)
			
			// Check if the result matches expected
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetEligibleCampaigns() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper function to compare floating point values with a tolerance
func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}