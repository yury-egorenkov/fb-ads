package optimization

import (
	"fmt"
	"math"
)

// BudgetCalculator handles budget calculations for campaign optimization
type BudgetCalculator struct {
	TotalBudget          float64
	TestBudgetPercentage float64
	MaxCPM               float64
}

// NewBudgetCalculator creates a new budget calculator
func NewBudgetCalculator(totalBudget, testBudgetPercentage, maxCPM float64) (*BudgetCalculator, error) {
	if totalBudget <= 0 {
		return nil, fmt.Errorf("total budget must be greater than 0")
	}
	
	if testBudgetPercentage <= 0 || testBudgetPercentage > 100 {
		return nil, fmt.Errorf("test budget percentage must be between 0 and 100")
	}
	
	if maxCPM <= 0 {
		return nil, fmt.Errorf("max CPM must be greater than 0")
	}
	
	return &BudgetCalculator{
		TotalBudget:          totalBudget,
		TestBudgetPercentage: testBudgetPercentage,
		MaxCPM:               maxCPM,
	}, nil
}

// GetTestBudget returns the total budget allocated for testing
func (bc *BudgetCalculator) GetTestBudget() float64 {
	return bc.TotalBudget * bc.TestBudgetPercentage / 100
}

// GetMainBudget returns the main budget (total - test)
func (bc *BudgetCalculator) GetMainBudget() float64 {
	return bc.TotalBudget - bc.GetTestBudget()
}

// GetBudgetPerCampaign calculates the budget for each test campaign
func (bc *BudgetCalculator) GetBudgetPerCampaign(numCampaigns int) (float64, error) {
	if numCampaigns <= 0 {
		return 0, fmt.Errorf("number of campaigns must be greater than 0")
	}
	
	testBudget := bc.GetTestBudget()
	budgetPerCampaign := testBudget / float64(numCampaigns)
	
	// Round to 2 decimal places for currency
	budgetPerCampaign = math.Round(budgetPerCampaign*100) / 100
	
	return budgetPerCampaign, nil
}

// CalculateImpressions estimates the number of impressions a campaign will get
func (bc *BudgetCalculator) CalculateImpressions(budget, cpm float64) (int, error) {
	if budget <= 0 {
		return 0, fmt.Errorf("budget must be greater than 0")
	}
	
	if cpm <= 0 {
		return 0, fmt.Errorf("CPM must be greater than 0")
	}
	
	// Budget / CPM * 1000 = estimated impressions
	impressions := budget / cpm * 1000
	
	return int(math.Floor(impressions)), nil
}

// CalculateOptimalCPM calculates the optimal CPM based on a set of active campaigns
func CalculateOptimalCPM(cpms []float64, maxCPM float64) (float64, error) {
	if len(cpms) == 0 {
		return 0, fmt.Errorf("no CPM data provided")
	}
	
	// Calculate mean
	sum := 0.0
	for _, cpm := range cpms {
		sum += cpm
	}
	mean := sum / float64(len(cpms))
	
	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, cpm := range cpms {
		diff := cpm - mean
		sumSquaredDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(len(cpms)))
	
	// Optimal CPM is mean + 1 standard deviation, but not higher than maxCPM
	optimalCPM := mean + stdDev
	if optimalCPM > maxCPM {
		optimalCPM = maxCPM
	}
	
	// Round to 2 decimal places
	optimalCPM = math.Round(optimalCPM*100) / 100
	
	return optimalCPM, nil
}

// ShouldTerminateCampaign determines if a campaign should be terminated based on performance
func ShouldTerminateCampaign(campaignImpressions int, worstActiveImpressions int) bool {
	return campaignImpressions < worstActiveImpressions
}