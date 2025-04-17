package optimization

import (
	"testing"
)

func TestNewBudgetCalculator(t *testing.T) {
	tests := []struct {
		name                 string
		totalBudget          float64
		testBudgetPercentage float64
		maxCPM               float64
		wantErr              bool
	}{
		{
			name:                 "Valid parameters",
			totalBudget:          1000,
			testBudgetPercentage: 20,
			maxCPM:               15,
			wantErr:              false,
		},
		{
			name:                 "Zero total budget",
			totalBudget:          0,
			testBudgetPercentage: 20,
			maxCPM:               15,
			wantErr:              true,
		},
		{
			name:                 "Negative total budget",
			totalBudget:          -1000,
			testBudgetPercentage: 20,
			maxCPM:               15,
			wantErr:              true,
		},
		{
			name:                 "Zero test budget percentage",
			totalBudget:          1000,
			testBudgetPercentage: 0,
			maxCPM:               15,
			wantErr:              true,
		},
		{
			name:                 "Negative test budget percentage",
			totalBudget:          1000,
			testBudgetPercentage: -20,
			maxCPM:               15,
			wantErr:              true,
		},
		{
			name:                 "Test budget percentage > 100",
			totalBudget:          1000,
			testBudgetPercentage: 120,
			maxCPM:               15,
			wantErr:              true,
		},
		{
			name:                 "Zero max CPM",
			totalBudget:          1000,
			testBudgetPercentage: 20,
			maxCPM:               0,
			wantErr:              true,
		},
		{
			name:                 "Negative max CPM",
			totalBudget:          1000,
			testBudgetPercentage: 20,
			maxCPM:               -15,
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBudgetCalculator(tt.totalBudget, tt.testBudgetPercentage, tt.maxCPM)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBudgetCalculator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestBudgetCalculator_GetTestBudget(t *testing.T) {
	bc, _ := NewBudgetCalculator(1000, 20, 15)
	expected := 200.0 // 20% of 1000
	if got := bc.GetTestBudget(); got != expected {
		t.Errorf("GetTestBudget() = %v, want %v", got, expected)
	}
}

func TestBudgetCalculator_GetMainBudget(t *testing.T) {
	bc, _ := NewBudgetCalculator(1000, 20, 15)
	expected := 800.0 // Total 1000 - test budget 200
	if got := bc.GetMainBudget(); got != expected {
		t.Errorf("GetMainBudget() = %v, want %v", got, expected)
	}
}

func TestBudgetCalculator_GetBudgetPerCampaign(t *testing.T) {
	bc, _ := NewBudgetCalculator(1000, 20, 15)
	
	tests := []struct {
		name         string
		numCampaigns int
		want         float64
		wantErr      bool
	}{
		{
			name:         "10 campaigns",
			numCampaigns: 10,
			want:         20.0, // 200 / 10
			wantErr:      false,
		},
		{
			name:         "7 campaigns",
			numCampaigns: 7,
			want:         28.57, // 200 / 7 rounded to 2 decimals
			wantErr:      false,
		},
		{
			name:         "Zero campaigns",
			numCampaigns: 0,
			want:         0,
			wantErr:      true,
		},
		{
			name:         "Negative campaigns",
			numCampaigns: -5,
			want:         0,
			wantErr:      true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bc.GetBudgetPerCampaign(tt.numCampaigns)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBudgetPerCampaign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetBudgetPerCampaign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBudgetCalculator_CalculateImpressions(t *testing.T) {
	bc, _ := NewBudgetCalculator(1000, 20, 15)
	
	tests := []struct {
		name    string
		budget  float64
		cpm     float64
		want    int
		wantErr bool
	}{
		{
			name:    "Valid parameters",
			budget:  100,
			cpm:     5,
			want:    20000, // (100 / 5) * 1000
			wantErr: false,
		},
		{
			name:    "Zero budget",
			budget:  0,
			cpm:     5,
			want:    0,
			wantErr: true,
		},
		{
			name:    "Negative budget",
			budget:  -100,
			cpm:     5,
			want:    0,
			wantErr: true,
		},
		{
			name:    "Zero CPM",
			budget:  100,
			cpm:     0,
			want:    0,
			wantErr: true,
		},
		{
			name:    "Negative CPM",
			budget:  100,
			cpm:     -5,
			want:    0,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bc.CalculateImpressions(tt.budget, tt.cpm)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateImpressions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CalculateImpressions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateOptimalCPM(t *testing.T) {
	tests := []struct {
		name    string
		cpms    []float64
		maxCPM  float64
		want    float64
		wantErr bool
	}{
		{
			name:    "One CPM",
			cpms:    []float64{5.0},
			maxCPM:  15.0,
			want:    5.0, // mean + stddev = 5 + 0
			wantErr: false,
		},
		{
			name:    "Multiple CPMs",
			cpms:    []float64{3.0, 5.0, 7.0},
			maxCPM:  15.0,
			want:    6.63, // mean = 5.0, stddev = 1.63, mean + stddev = 6.63
			wantErr: false,
		},
		{
			name:    "Exceeds max CPM",
			cpms:    []float64{10.0, 15.0, 20.0},
			maxCPM:  12.0,
			want:    12.0, // mean + stddev = 15 + 4.08 > 12, so capped at 12
			wantErr: false,
		},
		{
			name:    "Empty CPMs",
			cpms:    []float64{},
			maxCPM:  15.0,
			want:    0,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateOptimalCPM(tt.cpms, tt.maxCPM)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateOptimalCPM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CalculateOptimalCPM() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldTerminateCampaign(t *testing.T) {
	tests := []struct {
		name                  string
		campaignImpressions   int
		worstActiveImpressions int
		want                  bool
	}{
		{
			name:                  "Should terminate",
			campaignImpressions:   500,
			worstActiveImpressions: 1000,
			want:                  true,
		},
		{
			name:                  "Equal impressions",
			campaignImpressions:   1000,
			worstActiveImpressions: 1000,
			want:                  false,
		},
		{
			name:                  "Better performance",
			campaignImpressions:   1500,
			worstActiveImpressions: 1000,
			want:                  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldTerminateCampaign(tt.campaignImpressions, tt.worstActiveImpressions); got != tt.want {
				t.Errorf("ShouldTerminateCampaign() = %v, want %v", got, tt.want)
			}
		})
	}
}