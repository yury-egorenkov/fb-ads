package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/user/fb-ads/internal/api"
	"github.com/user/fb-ads/internal/config"
	"github.com/user/fb-ads/pkg/auth"
	"github.com/user/fb-ads/pkg/models"
)

func main() {
	// Define command line flags
	campaignID := flag.String("id", "", "Campaign ID to update")
	statusFlag := flag.String("status", "", "New campaign status (ACTIVE, PAUSED, ARCHIVED)")
	nameFlag := flag.String("name", "", "New campaign name")
	dailyBudgetFlag := flag.Float64("daily_budget", 0, "New daily budget in your currency (e.g. 50.00)")
	lifetimeBudgetFlag := flag.Float64("lifetime_budget", 0, "New lifetime budget in your currency (e.g. 1000.00)")
	bidStrategyFlag := flag.String("bid_strategy", "", "New bid strategy (e.g. LOWEST_COST_WITHOUT_CAP)")
	jsonFileFlag := flag.String("file", "", "JSON file containing campaign update details")
	
	flag.Parse()

	// Check if at least campaign ID and one update field is provided
	if *campaignID == "" {
		fmt.Println("Error: Campaign ID is required")
		fmt.Println("Usage: update_campaign -id=CAMPAIGN_ID [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Check if at least one update parameter is provided
	if *statusFlag == "" && *nameFlag == "" && *dailyBudgetFlag == 0 && *lifetimeBudgetFlag == 0 && 
	   *bidStrategyFlag == "" && *jsonFileFlag == "" {
		fmt.Println("Error: At least one update parameter must be provided")
		fmt.Println("Usage: update_campaign -id=CAMPAIGN_ID [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Create the Facebook auth object
	fbAuth := auth.NewFacebookAuth(cfg.AppID, cfg.AppSecret, cfg.AccessToken, cfg.APIVersion)

	// Create a new client
	client := api.NewClient(fbAuth, cfg.AccountID)

	// Build the update parameters
	params := url.Values{}

	// If a JSON file is provided, load update parameters from it
	if *jsonFileFlag != "" {
		fileParams, err := loadParamsFromFile(*jsonFileFlag)
		if err != nil {
			log.Fatalf("Error loading parameters from file: %v", err)
		}
		
		// Merge file parameters with params
		for key, values := range fileParams {
			for _, value := range values {
				params.Add(key, value)
			}
		}
	}

	// Add command-line parameters (these override file parameters)
	if *statusFlag != "" {
		validStatuses := map[string]bool{"ACTIVE": true, "PAUSED": true, "ARCHIVED": true}
		if !validStatuses[strings.ToUpper(*statusFlag)] {
			log.Fatalf("Invalid status: %s. Must be one of: ACTIVE, PAUSED, ARCHIVED", *statusFlag)
		}
		params.Set("status", strings.ToUpper(*statusFlag))
	}
	
	if *nameFlag != "" {
		params.Set("name", *nameFlag)
	}
	
	if *dailyBudgetFlag > 0 {
		// Convert to cents as required by the API
		params.Set("daily_budget", fmt.Sprintf("%d", int(*dailyBudgetFlag*100)))
	}
	
	if *lifetimeBudgetFlag > 0 {
		// Convert to cents as required by the API
		params.Set("lifetime_budget", fmt.Sprintf("%d", int(*lifetimeBudgetFlag*100)))
	}
	
	if *bidStrategyFlag != "" {
		params.Set("bid_strategy", *bidStrategyFlag)
	}

	// Make the API call to update the campaign
	fmt.Printf("Updating campaign %s with parameters: %v\n", *campaignID, params)
	err = client.UpdateCampaign(*campaignID, params)
	if err != nil {
		log.Fatalf("Error updating campaign: %v", err)
	}
	
	fmt.Printf("Campaign %s updated successfully\n", *campaignID)
}

// loadParamsFromFile loads campaign update parameters from a JSON file
func loadParamsFromFile(filePath string) (url.Values, error) {
	params := url.Values{}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return params, fmt.Errorf("error reading file: %w", err)
	}

	// Parse JSON
	var updateConfig struct {
		Status         string  `json:"status,omitempty"`
		Name           string  `json:"name,omitempty"`
		DailyBudget    float64 `json:"daily_budget,omitempty"`
		LifetimeBudget float64 `json:"lifetime_budget,omitempty"`
		BidStrategy    string  `json:"bid_strategy,omitempty"`
	}

	if err := json.Unmarshal(data, &updateConfig); err != nil {
		return params, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Add parameters
	if updateConfig.Status != "" {
		validStatuses := map[string]bool{"ACTIVE": true, "PAUSED": true, "ARCHIVED": true}
		status := strings.ToUpper(updateConfig.Status)
		if !validStatuses[status] {
			return params, fmt.Errorf("invalid status: %s. Must be one of: ACTIVE, PAUSED, ARCHIVED", status)
		}
		params.Set("status", status)
	}
	
	if updateConfig.Name != "" {
		params.Set("name", updateConfig.Name)
	}
	
	if updateConfig.DailyBudget > 0 {
		// Convert to cents as required by the API
		params.Set("daily_budget", fmt.Sprintf("%d", int(updateConfig.DailyBudget*100)))
	}
	
	if updateConfig.LifetimeBudget > 0 {
		// Convert to cents as required by the API
		params.Set("lifetime_budget", fmt.Sprintf("%d", int(updateConfig.LifetimeBudget*100)))
	}
	
	if updateConfig.BidStrategy != "" {
		params.Set("bid_strategy", updateConfig.BidStrategy)
	}

	return params, nil
}