package optimization

import (
	"math"
	"testing"
)

func TestCalculateMean(t *testing.T) {
	analyzer := NewStatisticalAnalyzer()
	
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
			name:     "multiple values",
			values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 3.0,
		},
		{
			name:     "negative values",
			values:   []float64{-1.0, -2.0, -3.0, -4.0, -5.0},
			expected: -3.0,
		},
		{
			name:     "mixed values",
			values:   []float64{-10.0, 0.0, 10.0},
			expected: 0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CalculateMean(tt.values)
			if result != tt.expected {
				t.Errorf("CalculateMean() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateStandardDeviation(t *testing.T) {
	analyzer := NewStatisticalAnalyzer()
	
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
			expected: 0,
		},
		{
			name:     "identical values",
			values:   []float64{3.0, 3.0, 3.0},
			expected: 0,
		},
		{
			name:     "standard case",
			values:   []float64{2.0, 4.0, 6.0, 8.0, 10.0},
			expected: 2.83, // Rounded expected value
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.CalculateStandardDeviation(tt.values)
			
			// For the standard case, we'll round to 2 decimal places for comparison
			if tt.name == "standard case" {
				roundedResult := math.Round(result*100) / 100
				if roundedResult != tt.expected {
					t.Errorf("CalculateStandardDeviation() = %v, want %v", roundedResult, tt.expected)
				}
			} else {
				if result != tt.expected {
					t.Errorf("CalculateStandardDeviation() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}


func TestIsOutlier(t *testing.T) {
	analyzer := NewStatisticalAnalyzer()
	
	tests := []struct {
		name     string
		value    float64
		values   []float64
		expected bool
	}{
		{
			name:     "empty slice",
			value:    5.0,
			values:   []float64{},
			expected: false,
		},
		{
			name:     "single value",
			value:    5.0,
			values:   []float64{5.0},
			expected: false,
		},
		{
			name:     "not an outlier",
			value:    6.0,
			values:   []float64{4.0, 5.0, 6.0, 7.0, 8.0},
			expected: false,
		},
		{
			name:     "is an outlier (high)",
			value:    15.0,
			values:   []float64{4.0, 5.0, 6.0, 7.0, 8.0},
			expected: true,
		},
		{
			name:     "is an outlier (low)",
			value:    -5.0,
			values:   []float64{4.0, 5.0, 6.0, 7.0, 8.0},
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.IsOutlier(tt.value, tt.values)
			if result != tt.expected {
				t.Errorf("IsOutlier() = %v, want %v", result, tt.expected)
			}
		})
	}
}