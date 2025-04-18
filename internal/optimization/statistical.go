package optimization

import (
	"math"
)

// StatisticalAnalyzer provides methods for statistical analysis of campaign data
type StatisticalAnalyzer struct{}

// NewStatisticalAnalyzer creates a new instance of StatisticalAnalyzer
func NewStatisticalAnalyzer() *StatisticalAnalyzer {
	return &StatisticalAnalyzer{}
}

// CalculateMean calculates the arithmetic mean of a slice of float64 values
func (s *StatisticalAnalyzer) CalculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// CalculateStandardDeviation calculates the standard deviation of a slice of float64 values
func (s *StatisticalAnalyzer) CalculateStandardDeviation(values []float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	mean := s.CalculateMean(values)
	var sumSquaredDiff float64
	
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	
	// Using population standard deviation formula (divide by n)
	// Use (n-1) instead if calculating sample standard deviation
	variance := sumSquaredDiff / float64(len(values))
	return math.Sqrt(variance)
}

// CalculateOptimalCPM calculates the optimal CPM value based on existing CPMs
// It uses the mean + 1 standard deviation approach as specified in the requirements
func (s *StatisticalAnalyzer) CalculateOptimalCPM(cpmValues []float64, maxCPM float64) float64 {
	if len(cpmValues) == 0 {
		return maxCPM
	}

	mean := s.CalculateMean(cpmValues)
	stdDev := s.CalculateStandardDeviation(cpmValues)
	
	// Calculate optimal CPM as mean + 1 standard deviation
	optimalCPM := mean + stdDev
	
	// Honor the user-defined maximum CPM threshold
	if optimalCPM > maxCPM {
		optimalCPM = maxCPM
	}
	
	return optimalCPM
}

// IsOutlier determines if a value is an outlier (>2 std dev from mean)
// Used for anomaly detection in campaign performance
func (s *StatisticalAnalyzer) IsOutlier(value float64, values []float64) bool {
	if len(values) <= 1 {
		return false
	}
	
	mean := s.CalculateMean(values)
	stdDev := s.CalculateStandardDeviation(values)
	
	// If the value is more than 2 standard deviations from the mean, it's an outlier
	return math.Abs(value-mean) > (2 * stdDev)
}