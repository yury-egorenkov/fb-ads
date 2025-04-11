package utils

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/user/fb-ads/pkg/auth"
)

// CampaignPerformance contains performance metrics for a campaign
type CampaignPerformance struct {
	CampaignID    string  `json:"campaign_id"`
	Name          string  `json:"name"`
	Spend         float64 `json:"spend"`
	Impressions   int     `json:"impressions"`
	Clicks        int     `json:"clicks"`
	Conversions   int     `json:"conversions"`
	CPC           float64 `json:"cpc"`
	CPM           float64 `json:"cpm"`
	CTR           float64 `json:"ctr"`
	CPA           float64 `json:"cpa"`
	ROAS          float64 `json:"roas"`
	LastUpdated   time.Time `json:"last_updated"`
}

// BidAdjustment contains information about a bid adjustment
type BidAdjustment struct {
	CampaignID    string    `json:"campaign_id"`
	AdSetID       string    `json:"adset_id"`
	OldBid        float64   `json:"old_bid"`
	NewBid        float64   `json:"new_bid"`
	Reason        string    `json:"reason"`
	PercentChange float64   `json:"percent_change"`
	Timestamp     time.Time `json:"timestamp"`
}

// Optimizer handles campaign optimizations
type Optimizer struct {
	httpClient      *http.Client
	auth            *auth.FacebookAuth
	accountID       string
	targetCPA       float64
	minBid          float64
	maxBid          float64
	adjustThreshold float64
}

// NewOptimizer creates a new campaign optimizer
func NewOptimizer(auth *auth.FacebookAuth, accountID string, targetCPA float64) *Optimizer {
	return &Optimizer{
		httpClient:      &http.Client{},
		auth:            auth,
		accountID:       accountID,
		targetCPA:       targetCPA,
		minBid:          1.0,    // $1 minimum bid
		maxBid:          20.0,   // $20 maximum bid
		adjustThreshold: 0.20,   // 20% adjustment threshold
	}
}

// OptimizeCampaigns adjusts bids based on performance
func (o *Optimizer) OptimizeCampaigns() ([]BidAdjustment, error) {
	// Get campaign performance data
	performances, err := o.GetCampaignPerformances()
	if err != nil {
		return nil, fmt.Errorf("error getting campaign performances: %w", err)
	}
	
	var adjustments []BidAdjustment
	
	for _, perf := range performances {
		// Skip campaigns with no conversions
		if perf.Conversions == 0 {
			continue
		}
		
		// Calculate current CPA
		currentCPA := perf.Spend / float64(perf.Conversions)
		
		// Determine if adjustment is needed
		if currentCPA > o.targetCPA*(1+o.adjustThreshold) {
			// CPA is too high, decrease bid
			adjustment := o.calculateBidAdjustment(perf, currentCPA, false)
			if adjustment != nil {
				adjustments = append(adjustments, *adjustment)
			}
		} else if currentCPA < o.targetCPA*(1-o.adjustThreshold) {
			// CPA is too low, we can increase bid
			adjustment := o.calculateBidAdjustment(perf, currentCPA, true)
			if adjustment != nil {
				adjustments = append(adjustments, *adjustment)
			}
		}
	}
	
	return adjustments, nil
}

// GetCampaignPerformances retrieves performance data for all campaigns
func (o *Optimizer) GetCampaignPerformances() ([]CampaignPerformance, error) {
	// TODO: Implement actual API call to get performance data
	// This is a placeholder
	return []CampaignPerformance{}, nil
}

// AdjustBid changes the bid for an ad set
func (o *Optimizer) AdjustBid(adSetID string, newBid float64) error {
	// TODO: Implement actual bid adjustment via API
	log.Printf("Adjusting bid for ad set %s to $%.2f", adSetID, newBid)
	return nil
}

// calculateBidAdjustment determines the appropriate bid adjustment
func (o *Optimizer) calculateBidAdjustment(perf CampaignPerformance, currentCPA float64, increase bool) *BidAdjustment {
	// TODO: Get current bid amount for the ad set
	// For now, we'll use a placeholder value
	currentBid := 10.0 // $10 placeholder bid
	
	var adjustment float64
	var reason string
	
	if increase {
		// Increase bid to try to get more conversions
		adjustment = 1.15 // 15% increase
		reason = "CPA below target, increasing bid to maximize delivery"
	} else {
		// Decrease bid to lower CPA
		adjustment = 0.85 // 15% decrease
		reason = "CPA above target, decreasing bid to improve efficiency"
	}
	
	newBid := currentBid * adjustment
	
	// Enforce min/max bid limits
	if newBid < o.minBid {
		newBid = o.minBid
	} else if newBid > o.maxBid {
		newBid = o.maxBid
	}
	
	// If the adjustment is very small, don't bother
	if newBid == currentBid {
		return nil
	}
	
	return &BidAdjustment{
		CampaignID:    perf.CampaignID,
		AdSetID:       "placeholder-adset-id", // TODO: Get actual ad set ID
		OldBid:        currentBid,
		NewBid:        newBid,
		Reason:        reason,
		PercentChange: (newBid - currentBid) / currentBid * 100,
		Timestamp:     time.Now(),
	}
}