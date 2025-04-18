package optimization

import (
	"math"
	"reflect"
	"testing"
)

func TestGetCampaignsToTerminate(t *testing.T) {
	tests := []struct {
		name           string
		minImpressions int
		campaigns      []CampaignPerformance
		expected       []string
	}{
		{
			name:           "empty campaigns",
			minImpressions: 1000,
			campaigns:      []CampaignPerformance{},
			expected:       []string{},
		},
		{
			name:           "no campaigns meet threshold",
			minImpressions: 1000,
			campaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 500, CPC: 2.0},
				{CampaignID: "2", Impressions: 800, CPC: 3.0},
			},
			expected: []string{},
		},
		{
			name:           "some campaigns below threshold",
			minImpressions: 1000,
			campaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, CPC: 2.0},
				{CampaignID: "2", Impressions: 500, CPC: 3.0}, // Below threshold
				{CampaignID: "3", Impressions: 1500, CPC: 2.5},
			},
			expected: []string{"2"},
		},
		{
			name:           "all campaigns meet threshold",
			minImpressions: 1000,
			campaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, CPC: 2.0},
				{CampaignID: "2", Impressions: 1500, CPC: 3.0},
				{CampaignID: "3", Impressions: 1800, CPC: 2.5},
			},
			expected: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terminator := NewTerminator(tt.minImpressions)
			result := terminator.GetCampaignsToTerminate(tt.campaigns)
			
			// Sort the slices to ensure consistent comparison
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetCampaignsToTerminate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetUnderperformingCampaigns(t *testing.T) {
	tests := []struct {
		name               string
		minImpressions     int
		cpcThresholdFactor float64
		campaigns          []CampaignPerformance
		expected           []string
	}{
		{
			name:               "empty campaigns",
			minImpressions:     1000,
			cpcThresholdFactor: 1.5,
			campaigns:          []CampaignPerformance{},
			expected:           []string{},
		},
		{
			name:               "single campaign",
			minImpressions:     1000,
			cpcThresholdFactor: 1.5,
			campaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, CPC: 2.0},
			},
			expected: []string{},
		},
		{
			name:               "no underperforming campaigns",
			minImpressions:     1000,
			cpcThresholdFactor: 1.5,
			campaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, CPC: 2.0},
				{CampaignID: "2", Impressions: 1500, CPC: 3.0},
			},
			expected: []string{},
		},
		{
			name:               "some underperforming campaigns",
			minImpressions:     1000,
			cpcThresholdFactor: 1.5,
			campaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, CPC: 2.0},
				{CampaignID: "2", Impressions: 1500, CPC: 3.0},
				{CampaignID: "3", Impressions: 1800, CPC: 5.0}, // This is underperforming
				{CampaignID: "4", Impressions: 1600, CPC: 6.0}, // This is underperforming
			},
			expected: []string{"4"}, // Only Campaign 4 has CPC > 5.25 (median * 1.5)
		},
		{
			name:               "campaigns below threshold ignored",
			minImpressions:     1000,
			cpcThresholdFactor: 1.5,
			campaigns: []CampaignPerformance{
				{CampaignID: "1", Impressions: 1200, CPC: 2.0},
				{CampaignID: "2", Impressions: 1500, CPC: 3.0},
				{CampaignID: "3", Impressions: 800, CPC: 10.0}, // Below threshold, should be ignored
				{CampaignID: "4", Impressions: 1600, CPC: 6.0}, // This is underperforming
			},
			expected: []string{"4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terminator := NewTerminator(tt.minImpressions)
			result := terminator.GetUnderperformingCampaigns(tt.campaigns, tt.cpcThresholdFactor)
			
			// Debug information
			if tt.name == "some underperforming campaigns" {
				// Filter valid campaigns
				validCampaigns := []CampaignPerformance{}
				for _, campaign := range tt.campaigns {
					if campaign.Impressions >= tt.minImpressions {
						validCampaigns = append(validCampaigns, campaign)
					}
				}
				
				// Calculate median CPC
				cpcValues := make([]float64, len(validCampaigns))
				for i, campaign := range validCampaigns {
					cpcValues[i] = campaign.CPC
				}
				medianCPC := calculateMedian(cpcValues)
				threshold := medianCPC * tt.cpcThresholdFactor
				
				t.Logf("Valid campaigns: %+v", validCampaigns)
				t.Logf("CPC values: %+v", cpcValues)
				t.Logf("Median CPC: %v", medianCPC)
				t.Logf("Threshold (median * factor): %v", threshold)
				for _, campaign := range validCampaigns {
					t.Logf("Campaign %s: CPC %v > threshold %v? %v", 
						campaign.CampaignID, campaign.CPC, threshold, campaign.CPC > threshold)
				}
			}
			
			// Sort the slices to ensure consistent comparison
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetUnderperformingCampaigns() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "empty slice",
			values:   []float64{},
			expected: 0,
		},
		{
			name:     "single value",
			values:   []float64{5.0},
			expected: 5.0,
		},
		{
			name:     "odd number of values",
			values:   []float64{1.0, 3.0, 5.0, 7.0, 9.0},
			expected: 5.0,
		},
		{
			name:     "even number of values",
			values:   []float64{1.0, 3.0, 5.0, 7.0},
			expected: 4.0,
		},
		{
			name:     "unsorted values",
			values:   []float64{9.0, 3.0, 7.0, 1.0, 5.0},
			expected: 5.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMedian(tt.values)
			
			if math.Abs(result - tt.expected) > 0.0001 {
				t.Errorf("calculateMedian() = %v, want %v", result, tt.expected)
			}
		})
	}
}
