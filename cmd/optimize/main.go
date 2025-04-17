package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/fb-ads/internal/config"
	"github.com/user/fb-ads/internal/optimization"
)

func main() {
	// Parse command line arguments
	configPath := flag.String("config", "", "Path to the configuration file")
	yamlPath := flag.String("yaml", "", "Path to the YAML campaign configuration file")
	validate := flag.Bool("validate", false, "Only validate the YAML file without further action")
	flag.Parse()

	// Check if config file was provided
	if *configPath == "" {
		homeDir, _ := os.UserHomeDir()
		defaultPath := filepath.Join(homeDir, ".fbads", "config.json")
		if _, err := os.Stat(defaultPath); err == nil {
			*configPath = defaultPath
		} else {
			fmt.Println("Error: No configuration file provided")
			fmt.Println("Use -config flag to specify a configuration file or create ~/.fbads/config.json")
			os.Exit(1)
		}
	}

	// Load application configuration
	_, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Check if YAML file was provided
	if *yamlPath == "" {
		fmt.Println("Error: No YAML campaign configuration file provided")
		fmt.Println("Use -yaml flag to specify a YAML file")
		os.Exit(1)
	}

	// Parse YAML configuration
	campaignCfg, err := optimization.ParseYAMLConfig(*yamlPath)
	if err != nil {
		fmt.Printf("Error parsing YAML configuration: %v\n", err)
		os.Exit(1)
	}

	// If only validating, we're done
	if *validate {
		fmt.Println("YAML configuration is valid")
		fmt.Println("Campaign Name:", campaignCfg.Campaign.Name)
		fmt.Printf("Total Budget: $%.2f\n", campaignCfg.Campaign.TotalBudget)
		fmt.Printf("Test Budget: $%.2f (%.1f%%)\n", 
			campaignCfg.Campaign.TotalBudget * campaignCfg.Campaign.TestBudgetPercentage / 100, 
			campaignCfg.Campaign.TestBudgetPercentage)
		fmt.Printf("Max CPM: $%.2f\n", campaignCfg.Campaign.MaxCPM)
		fmt.Printf("Creatives: %d\n", len(campaignCfg.Creatives))
		fmt.Printf("Audiences: %d\n", len(campaignCfg.TargetingOptions.Audiences))
		fmt.Printf("Placements: %d\n", len(campaignCfg.TargetingOptions.Placements))
		
		// Create budget calculator
		budgetCalc, err := optimization.NewBudgetCalculator(
			campaignCfg.Campaign.TotalBudget,
			campaignCfg.Campaign.TestBudgetPercentage,
			campaignCfg.Campaign.MaxCPM,
		)
		if err != nil {
			fmt.Printf("Error creating budget calculator: %v\n", err)
			os.Exit(1)
		}
		
		// Calculate total number of test campaigns
		totalCombinations := len(campaignCfg.Creatives) * 
			(len(campaignCfg.TargetingOptions.Audiences) + len(campaignCfg.TargetingOptions.Placements))
		fmt.Printf("Total possible test combinations: %d\n", totalCombinations)
		
		// Calculate budget per campaign
		budgetPerCampaign, err := budgetCalc.GetBudgetPerCampaign(totalCombinations)
		if err != nil {
			fmt.Printf("Error calculating budget per campaign: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Budget per test campaign: $%.2f\n", budgetPerCampaign)
		
		// Estimate impressions with automatic CPM (using max CPM for estimate)
		impressions, err := budgetCalc.CalculateImpressions(budgetPerCampaign, budgetCalc.MaxCPM)
		if err != nil {
			fmt.Printf("Error calculating impressions: %v\n", err)
		} else {
			fmt.Printf("Estimated min impressions per campaign: %d\n", impressions)
			
			if impressions < 1000 {
				fmt.Printf("WARNING: Estimated impressions below recommended minimum (1000)\n")
				fmt.Printf("Consider reducing number of test combinations or increasing test budget\n")
			}
		}
		
		os.Exit(0)
	}

	// TODO: Implement the rest of the optimization workflow
	fmt.Println("Campaign optimization will be implemented in the next iteration")
}