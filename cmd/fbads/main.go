package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/fb-ads/internal/api"
	"github.com/user/fb-ads/internal/audience"
	internal_campaign "github.com/user/fb-ads/internal/campaign"
	"github.com/user/fb-ads/internal/config"
	"github.com/user/fb-ads/internal/optimization"
	"github.com/user/fb-ads/pkg/auth"
	"github.com/user/fb-ads/pkg/models"
	"github.com/user/fb-ads/pkg/utils"
)

func main() {
	fmt.Println("Facebook Ads Manager CLI")
	fmt.Println("------------------------")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	// Set default config path
	configPath := filepath.Join(homeDir, ".fbads", "config.json")

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error loading configuration: %v\n", err)
		fmt.Println("Using default configuration...")
		cfg = config.DefaultConfig()
	}

	// Process commands
	cmd := os.Args[1]

	switch cmd {
	case "list":
		listCampaigns(cfg)
	case "create":
		createCampaign(cfg)
	case "update":
		updateCampaign(cfg)
	case "delete":
		if len(os.Args) < 3 {
			fmt.Println("Missing campaign ID. Use: fbads delete <campaign_id>")
			os.Exit(1)
		}
		deleteCampaign(cfg, os.Args[2])
	case "duplicate":
		if len(os.Args) < 3 {
			fmt.Println("Missing campaign ID. Use: fbads duplicate <campaign_id> [options]")
			os.Exit(1)
		}
		duplicateCampaign(cfg, os.Args[2], os.Args[3:])
	case "export":
		if len(os.Args) < 3 {
			fmt.Println("Missing campaign ID. Use: fbads export <campaign_id> [output_file]")
			os.Exit(1)
		}
		exportCampaign(cfg, os.Args[2], os.Args[3:])
	case "exportyaml":
		if len(os.Args) < 3 {
			fmt.Println("Missing campaign ID. Use: fbads exportyaml <campaign_id> [output_file] [options]")
			os.Exit(1)
		}
		exportCampaignYAML(cfg, os.Args[2], os.Args[3:])
	case "pages":
		listPages(cfg)
	case "audience":
		analyzeAudience(cfg)
	case "stats":
		if len(os.Args) < 3 {
			fmt.Println("Missing stats subcommand. Use: fbads stats [collect|analyze|export]")
			os.Exit(1)
		}
		handleStatistics(cfg, os.Args[2], os.Args[3:])
	case "report":
		if len(os.Args) < 3 {
			fmt.Println("Missing report type. Use: fbads report [daily|weekly|monthly|custom]")
			os.Exit(1)
		}
		generateReport(cfg, os.Args[2], os.Args[3:])
	case "optimize":
		optimizeCampaigns(cfg)
	case "dashboard":
		startDashboard(cfg)
	case "config":
		configureApp(cfg, configPath)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func listCampaigns(cfg *config.Config) {
	// Parse flags
	var (
		limit  int
		status string
		format string
	)

	// Check for flags
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit", "-l":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &limit)
				i++
			}
		case "--status", "-s":
			if i+1 < len(args) {
				status = args[i+1]
				i++
			}
		case "--format", "-f":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		}
	}

	// Set defaults
	if limit <= 0 {
		limit = 10
	}
	if format == "" {
		format = "table" // Default to table format
	}

	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create API client
	client := api.NewClient(authClient, cfg.AccountID)

	fmt.Println("Fetching campaigns...")

	// Get campaigns
	campaigns, err := client.GetAllCampaigns()
	if err != nil {
		fmt.Printf("Error fetching campaigns: %v\n", err)
		os.Exit(1)
	}

	// Filter by status if specified
	if status != "" {
		filteredCampaigns := make([]models.Campaign, 0)
		status = strings.ToUpper(status)
		for _, campaign := range campaigns {
			if campaign.Status == status {
				filteredCampaigns = append(filteredCampaigns, campaign)
			}
		}
		campaigns = filteredCampaigns
	}

	// Limit results
	if limit > 0 && limit < len(campaigns) {
		campaigns = campaigns[:limit]
	}

	// Display results based on format
	switch format {
	case "json":
		displayCampaignsJSON(campaigns)
	case "csv":
		displayCampaignsCSV(campaigns)
	case "table":
		displayCampaignsTable(campaigns)
	default:
		fmt.Printf("Unknown format: %s. Supported formats: table, json, csv\n", format)
		os.Exit(1)
	}

	fmt.Printf("\nTotal: %d campaigns\n", len(campaigns))
}

// displayCampaignsTable displays campaigns in a formatted table
func displayCampaignsTable(campaigns []models.Campaign) {
	if len(campaigns) == 0 {
		fmt.Println("No campaigns found.")
		return
	}

	// Calculate column widths
	idWidth := 10
	nameWidth := 30
	statusWidth := 10
	budgetWidth := 15
	objectiveWidth := 20

	for _, campaign := range campaigns {
		if len(campaign.ID) > idWidth {
			idWidth = len(campaign.ID)
		}
		if len(campaign.Name) > nameWidth {
			nameWidth = len(campaign.Name)
		}
		if len(campaign.Status) > statusWidth {
			statusWidth = len(campaign.Status)
		}
		if len(campaign.ObjectiveType) > objectiveWidth {
			objectiveWidth = len(campaign.ObjectiveType)
		}
	}

	// Print header
	fmt.Printf("%-*s | %-*s | %-*s | %-*s | %-*s\n",
		idWidth, "ID",
		nameWidth, "NAME",
		statusWidth, "STATUS",
		budgetWidth, "BUDGET",
		objectiveWidth, "OBJECTIVE")

	// Print separator
	fmt.Printf("%s-+-%s-+-%s-+-%s-+-%s\n",
		strings.Repeat("-", idWidth),
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", statusWidth),
		strings.Repeat("-", budgetWidth),
		strings.Repeat("-", objectiveWidth))

	// Print rows
	for _, campaign := range campaigns {
		// Determine budget to display (daily or lifetime)
		var budget string
		if campaign.DailyBudget > 0 {
			budget = fmt.Sprintf("$%.2f/day", campaign.DailyBudget/100)
		} else if campaign.LifetimeBudget > 0 {
			budget = fmt.Sprintf("$%.2f total", campaign.LifetimeBudget/100)
		} else {
			budget = "N/A"
		}

		fmt.Printf("%-*s | %-*s | %-*s | %-*s | %-*s\n",
			idWidth, campaign.ID,
			nameWidth, truncateString(campaign.Name, nameWidth),
			statusWidth, campaign.Status,
			budgetWidth, budget,
			objectiveWidth, campaign.ObjectiveType)
	}
}

// displayCampaignsJSON displays campaigns in JSON format
func displayCampaignsJSON(campaigns []models.Campaign) {
	// Create a response structure to wrap the campaigns
	response := struct {
		Campaigns []models.Campaign `json:"campaigns"`
		Count     int               `json:"count"`
	}{
		Campaigns: campaigns,
		Count:     len(campaigns),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

// displayCampaignsCSV displays campaigns in CSV format
func displayCampaignsCSV(campaigns []models.Campaign) {
	// Print header
	fmt.Println("id,name,status,objective,budget_type,budget,bid_strategy,buying_type,created,updated")

	// Print rows
	for _, campaign := range campaigns {
		// Determine budget type and value
		budgetType := "none"
		var budget float64
		if campaign.DailyBudget > 0 {
			budgetType = "daily"
			budget = campaign.DailyBudget
		} else if campaign.LifetimeBudget > 0 {
			budgetType = "lifetime"
			budget = campaign.LifetimeBudget
		}

		// Format created and updated dates
		created := campaign.Created.Format("2006-01-02T15:04:05")
		updated := campaign.Updated.Format("2006-01-02T15:04:05")

		// Print the campaign as a CSV row
		fmt.Printf("%s,%s,%s,%s,%s,%.2f,%s,%s,%s,%s\n",
			campaign.ID,
			escapeCSV(campaign.Name),
			campaign.Status,
			campaign.ObjectiveType,
			budgetType,
			budget,
			campaign.BidStrategy,
			campaign.BuyingType,
			created,
			updated)
	}
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func escapeCSV(s string) string {
	if strings.Contains(s, ",") || strings.Contains(s, "\"") || strings.Contains(s, "\n") {
		s = strings.Replace(s, "\"", "\"\"", -1)
		s = "\"" + s + "\""
	}
	return s
}

func createCampaign(cfg *config.Config) {
	if len(os.Args) < 3 {
		fmt.Println("Missing campaign configuration file. Use: fbads create <config_file.json>")
		os.Exit(1)
	}

	configFile := os.Args[2]

	// Check for dry run flag
	dryRun := false
	for _, arg := range os.Args {
		if arg == "--dry-run" || arg == "-d" {
			dryRun = true
			break
		}
	}

	fmt.Printf("Reading campaign configuration from: %s\n", configFile)

	// Read the configuration file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading configuration file: %v\n", err)
		os.Exit(1)
	}

	// Parse the configuration
	var campaignConfig models.CampaignConfig
	if err := json.Unmarshal(configData, &campaignConfig); err != nil {
		fmt.Printf("Error parsing configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate the configuration
	if err := validateCampaignConfig(&campaignConfig); err != nil {
		fmt.Printf("Invalid campaign configuration: %v\n", err)
		os.Exit(1)
	}

	// Print configuration summary
	printCampaignConfigSummary(&campaignConfig)

	// If dry run, just print configuration summary and exit
	if dryRun {
		fmt.Println("\nDry run: No campaigns will be created.")
		return
	}

	// Ask for confirmation
	fmt.Print("\nDo you want to create this campaign? (y/n): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" && confirm != "yes" && confirm != "Yes" {
		fmt.Println("Campaign creation cancelled.")
		return
	}

	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create campaign creator from the internal/campaign package
	creator := internal_campaign.NewCampaignCreator(authClient, cfg.AccountID)

	fmt.Println("Creating campaign...")

	// Create the campaign
	err = creator.CreateFromConfig(&campaignConfig)
	if err != nil {
		fmt.Printf("Error creating campaign: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Campaign created successfully!")
}

// validateCampaignConfig validates the campaign configuration
func validateCampaignConfig(config *models.CampaignConfig) error {
	if config.Name == "" {
		return fmt.Errorf("campaign name is required")
	}

	if config.Objective == "" {
		return fmt.Errorf("campaign objective is required")
	}

	if config.BuyingType == "" {
		return fmt.Errorf("campaign buying type is required")
	}

	if config.DailyBudget == 0 && config.LifetimeBudget == 0 {
		return fmt.Errorf("either daily budget or lifetime budget is required")
	}

	if len(config.AdSets) == 0 {
		return fmt.Errorf("at least one ad set is required")
	}

	for i, adSet := range config.AdSets {
		if adSet.Name == "" {
			return fmt.Errorf("ad set #%d: name is required", i+1)
		}

		if adSet.OptimizationGoal == "" {
			return fmt.Errorf("ad set #%d: optimization goal is required", i+1)
		}

		if adSet.BillingEvent == "" {
			return fmt.Errorf("ad set #%d: billing event is required", i+1)
		}

		if len(adSet.Targeting) == 0 {
			return fmt.Errorf("ad set #%d: targeting is required", i+1)
		}
	}

	if len(config.Ads) == 0 {
		return fmt.Errorf("at least one ad is required")
	}

	for i, ad := range config.Ads {
		if ad.Name == "" {
			return fmt.Errorf("ad #%d: name is required", i+1)
		}

		// Check for title or name in the creative
		// Different templates might use Name instead of Title field
		if ad.Creative.Title == "" && ad.Creative.Name == "" {
			return fmt.Errorf("ad #%d: creative title/name is required", i+1)
		}

		if ad.Creative.LinkURL == "" {
			return fmt.Errorf("ad #%d: creative link URL is required", i+1)
		}

		// Now validate the Page ID as well, which is required
		if ad.Creative.PageID == "" {
			return fmt.Errorf("ad #%d: creative page_id is required", i+1)
		}
	}

	return nil
}

// printCampaignConfigSummary prints a summary of the campaign configuration
func printCampaignConfigSummary(config *models.CampaignConfig) {
	fmt.Println("\nCampaign Configuration Summary:")
	fmt.Printf("Name: %s\n", config.Name)
	fmt.Printf("Status: %s\n", config.Status)
	fmt.Printf("Objective: %s\n", config.Objective)
	fmt.Printf("Buying Type: %s\n", config.BuyingType)

	if config.DailyBudget > 0 {
		fmt.Printf("Daily Budget: $%.2f\n", config.DailyBudget)
	}

	if config.LifetimeBudget > 0 {
		fmt.Printf("Lifetime Budget: $%.2f\n", config.LifetimeBudget)
	}

	if config.StartTime != "" {
		fmt.Printf("Start Time: %s\n", config.StartTime)
	}

	if config.EndTime != "" {
		fmt.Printf("End Time: %s\n", config.EndTime)
	}

	fmt.Printf("\nAd Sets: %d\n", len(config.AdSets))
	for i, adSet := range config.AdSets {
		fmt.Printf("  %d. %s (Status: %s)\n", i+1, adSet.Name, adSet.Status)
		fmt.Printf("     Optimization Goal: %s\n", adSet.OptimizationGoal)
		fmt.Printf("     Billing Event: %s\n", adSet.BillingEvent)

		// Print targeting summary (simplified)
		if targeting, ok := adSet.Targeting["geo_locations"].(map[string]interface{}); ok {
			if countries, ok := targeting["countries"].([]interface{}); ok {
				fmt.Printf("     Countries: %v\n", countries)
			}
		}

		if ageMin, ok := adSet.Targeting["age_min"].(float64); ok {
			if ageMax, ok := adSet.Targeting["age_max"].(float64); ok {
				fmt.Printf("     Age Range: %d-%d\n", int(ageMin), int(ageMax))
			}
		}
	}

	fmt.Printf("\nAds: %d\n", len(config.Ads))
	for i, ad := range config.Ads {
		fmt.Printf("  %d. %s (Status: %s)\n", i+1, ad.Name, ad.Status)
		// Display either Title or Name
		titleValue := ad.Creative.Title
		if titleValue == "" {
			titleValue = ad.Creative.Name
		}
		fmt.Printf("     Title: %s\n", titleValue)
		if len(ad.Creative.Body) > 50 {
			fmt.Printf("     Body: %s...\n", ad.Creative.Body[:50])
		} else {
			fmt.Printf("     Body: %s\n", ad.Creative.Body)
		}
		fmt.Printf("     Link URL: %s\n", ad.Creative.LinkURL)
		if ad.Creative.CallToAction != "" {
			fmt.Printf("     Call to Action: %s\n", ad.Creative.CallToAction)
		}
		fmt.Printf("     Page ID: %s\n", ad.Creative.PageID)
	}
}

func analyzeAudience(cfg *config.Config) {
	// Parse flags and subcommands
	if len(os.Args) < 3 {
		fmt.Println("Missing audience subcommand. Available commands: search, filter, stats")
		os.Exit(1)
	}

	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create audience analyzer
	analyzer := audience.NewAudienceAnalyzer(authClient, cfg.AccountID)

	// Process subcommand
	subCmd := os.Args[2]

	switch subCmd {
	case "search":
		searchAudience(analyzer, os.Args[3:])
	case "filter":
		filterAudience(analyzer, os.Args[3:])
	case "stats":
		audienceStats(analyzer, os.Args[3:])
	default:
		fmt.Printf("Unknown audience subcommand: %s\n", subCmd)
		fmt.Println("Available subcommands: search, filter, stats")
		os.Exit(1)
	}
}

// searchAudience handles searching for audience segments
func searchAudience(analyzer *audience.AudienceAnalyzer, args []string) {
	if len(args) < 1 {
		fmt.Println("Missing search query. Use: fbads audience search <query> [--type TYPE] [--output FILE] [--class CLASS]")
		fmt.Println(`Available type options:
	adTargetingCategory: Search for interests, behaviors, demographics to use in ad targeting:
		--class [interests|behaviors|demographics]
	adinterest: Search for interests to use in ad targeting.
	adgeolocation: Search for geographic locations for targeting, such as countries, regions, cities, or zip codes.
	adlocale: Search for locales (languages) for targeting.
	adcountry: Search for countries for targeting.
	adregion: Search for regions within countries for targeting.
	adcity: Search for cities for targeting.
	adzip: Search for postal codes for targeting.
	adcustomaudience: Search for custom audiences.
	adworkemployer: Search for employers for targeting.
	adworkposition: Search for job positions for targeting.
	adeducationschool: Search for educational institutions for targeting.
	adeducationmajor: Search for education majors for targeting.
	adinterestvalid: Validate if an interest is valid for targeting.
		`)
		os.Exit(1)
	}

	index := 1
	query := args[0]
	if args[0] == "--class" || args[0] == "-c" {
		index = 0
		query = ""
	}

	searchType := "adinterest" // Default to interests

	var outputFile string

	var class string

	// Parse flags
	for i := index; i < len(args); i++ {
		switch args[i] {
		case "--type", "-t":
			if i+1 < len(args) {
				searchType = args[i+1]
				i++
			}
		case "--class", "-c":
			searchType = "adTargetingCategory"
			if i+1 < len(args) {
				class = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		}
	}

	var segments []audience.AudienceSegment
	var err error

	// Perform search based on type
	segments, err = analyzer.Search(searchType, class, query)

	if err != nil {
		fmt.Printf("Error searching for audience segments: %v\n", err)
		os.Exit(1)
	}

	// Display results
	if len(segments) == 0 {
		fmt.Printf("No %ss found matching your query.\n", searchType)
		return
	}

	fmt.Printf("Found %d %ss matching '%s':\n\n", len(segments), searchType, query)
	for i, segment := range segments {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, segment.Name, segment.ID)
		if segment.Description != "" {
			fmt.Printf("   Description: %s\n", segment.Description)
		}
		if segment.Type != "" {
			fmt.Printf("   Type: %s\n", segment.Type)
		}
		if segment.Path != "" {
			fmt.Printf("   Category: %s\n", segment.Path)
		}
		if segment.LowerBound > 0 || segment.UpperBound > 0 {
			fmt.Printf("   Audience size: %s\n", audience.FormatAudienceRange(segment.LowerBound, segment.UpperBound))
		}
		fmt.Println()
	}

	// Export to file if requested
	if outputFile != "" {
		err = analyzer.ExportAudienceData(outputFile, segments)
		if err != nil {
			fmt.Printf("Error exporting to file: %v\n", err)
			return
		}
		fmt.Printf("Exported %d segments to %s\n", len(segments), outputFile)
	}
}

// filterAudience handles filtering audience segments
func filterAudience(analyzer *audience.AudienceAnalyzer, args []string) {
	var query string
	var minSize, maxSize int64
	var types, keywords string
	var outputFile string

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--query", "-q":
			if i+1 < len(args) {
				query = args[i+1]
				i++
			}
		case "--min-size":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &minSize)
				i++
			}
		case "--max-size":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxSize)
				i++
			}
		case "--types", "-t":
			if i+1 < len(args) {
				types = args[i+1]
				i++
			}
		case "--keywords", "-k":
			if i+1 < len(args) {
				keywords = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		}
	}

	// First, we need to load some audience segments to filter
	// For simplicity, we'll search for a default term if no query is provided
	if query == "" {
		query = "shopping"
	}

	// For now, we'll just log what we would do in a full implementation
	fmt.Printf("Loading audience segments for '%s'...\n", query)

	// In a real implementation, we would search for both interests and behaviors
	// For example:
	// interests, err := analyzer.Search(query)
	// if err != nil {
	//     fmt.Printf("Error searching for interests: %v\n", err)
	//     os.Exit(1)
	// }
	//
	// behaviors, err := analyzer.GetBehaviors(query)
	// if err != nil {
	//     fmt.Printf("Error searching for behaviors: %v\n", err)
	// }

	// Create filter options
	options := make(map[string]interface{})

	if minSize > 0 {
		options["min_size"] = minSize
	}

	if maxSize > 0 {
		options["max_size"] = maxSize
	}

	if types != "" {
		typesArray := strings.Split(types, ",")
		options["types"] = typesArray
	}

	if keywords != "" {
		keywordsArray := strings.Split(keywords, ",")
		options["keywords"] = keywordsArray
	}

	fmt.Println("Filtering audience segments...")
	filtered, err := analyzer.FilterAudiences(options)
	if err != nil {
		fmt.Printf("Error filtering audiences: %v\n", err)
		os.Exit(1)
	}

	// Display results
	if len(filtered) == 0 {
		fmt.Println("No audience segments match your filter criteria.")
		return
	}

	fmt.Printf("Found %d audience segments matching your criteria:\n\n", len(filtered))
	for i, segment := range filtered {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, segment.Name, segment.ID)
		fmt.Printf("   Type: %s\n", segment.Type)
		if segment.Description != "" {
			fmt.Printf("   Description: %s\n", segment.Description)
		}
		if segment.LowerBound > 0 || segment.UpperBound > 0 {
			fmt.Printf("   Audience size: %s\n", audience.FormatAudienceRange(segment.LowerBound, segment.UpperBound))
		}
		fmt.Println()
	}

	// Export to file if requested
	if outputFile != "" {
		err = analyzer.ExportAudienceData(outputFile, filtered)
		if err != nil {
			fmt.Printf("Error exporting to file: %v\n", err)
			return
		}
		fmt.Printf("Exported filtered audience segments to %s\n", outputFile)
	}
}

// audienceStats handles collecting audience statistics
func audienceStats(analyzer *audience.AudienceAnalyzer, args []string) {
	var campaignID string
	days := 30 // Default to 30 days

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--campaign", "-c":
			if i+1 < len(args) {
				campaignID = args[i+1]
				i++
			}
		case "--days", "-d":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &days)
				i++
			}
		}
	}

	// Check if campaign ID is provided
	if campaignID == "" {
		fmt.Println("Missing campaign ID. Use: fbads audience stats --campaign CAMPAIGN_ID [--days DAYS]")
		os.Exit(1)
	}

	fmt.Printf("Collecting audience statistics for campaign %s over the last %d days...\n", campaignID, days)
	err := analyzer.CollectSegmentStatistics(campaignID, days)
	if err != nil {
		fmt.Printf("Error collecting audience statistics: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully collected audience statistics.")
}

func generateReport(cfg *config.Config, reportType string, args []string) {
	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create metrics collector
	metricsCollector := api.NewMetricsCollector(authClient, cfg.AccountID)

	// Create audience analyzer
	audienceAnalyzer := audience.NewAudienceAnalyzer(authClient, cfg.AccountID)

	// Create performance analyzer
	analyzer := api.NewPerformanceAnalyzer(metricsCollector, audienceAnalyzer)

	// Set default reports directory
	reportsDir := filepath.Join(cfg.ConfigDir, "reports")

	// Create report generator
	reportGenerator := api.NewReportGenerator(analyzer, metricsCollector, reportsDir)

	var err error

	switch reportType {
	case "daily":
		fmt.Println("Generating daily report...")
		err = reportGenerator.GenerateDailyReport()
	case "weekly":
		fmt.Println("Generating weekly report...")
		err = reportGenerator.GenerateWeeklyReport()
	case "custom":
		if len(args) < 2 {
			fmt.Println("Missing date range. Use: fbads report custom <start_date> <end_date>")
			fmt.Println("Date format: YYYY-MM-DD")
			os.Exit(1)
		}

		startDate, err := time.Parse("2006-01-02", args[0])
		if err != nil {
			fmt.Printf("Invalid start date format: %v\n", err)
			os.Exit(1)
		}

		endDate, err := time.Parse("2006-01-02", args[1])
		if err != nil {
			fmt.Printf("Invalid end date format: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Generating custom report for period: %s to %s\n", args[0], args[1])
		err = reportGenerator.GenerateCustomReport(startDate, endDate)
		if err != nil {
			fmt.Printf("Invalid end date format: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown report type: %s\n", reportType)
		fmt.Println("Available report types: daily, weekly, custom")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Report generated successfully in: %s\n", reportsDir)
}

func optimizeCampaigns(cfg *config.Config) {
	// Parse optimize subcommands
	if len(os.Args) < 3 {
		fmt.Println("Missing optimize subcommand. Available commands: validate, create, update")
		fmt.Println("\nUsage: fbads optimize <subcommand> [options]")
		fmt.Println("\nSubcommands:")
		fmt.Println("  validate <yaml_file>     Validate a YAML campaign configuration file")
		fmt.Println("  create <yaml_file>       Create test campaigns from a YAML configuration")
		fmt.Println("  update <campaign_ids>    Update campaign CPM based on performance data")
		os.Exit(1)
	}

	subCmd := os.Args[2]

	switch subCmd {
	case "validate":
		validateYAMLConfig(cfg, os.Args[3:])
	case "create":
		createTestCampaigns(cfg, os.Args[3:])
	case "update":
		updateCampaignCPM(cfg, os.Args[3:])
	default:
		fmt.Printf("Unknown optimize subcommand: %s\n", subCmd)
		fmt.Println("Available subcommands: validate, create, update")
		os.Exit(1)
	}
}

// validateYAMLConfig validates a YAML campaign configuration file
func validateYAMLConfig(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Println("Missing YAML file path. Use: fbads optimize validate <yaml_file>")
		os.Exit(1)
	}

	yamlPath := args[0]

	// Parse YAML configuration
	campaignCfg, err := optimization.ParseYAMLConfig(yamlPath)
	if err != nil {
		fmt.Printf("Error parsing YAML configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("YAML configuration is valid")
	fmt.Println("Campaign Name:", campaignCfg.Campaign.Name)
	fmt.Printf("Total Budget: $%.2f\n", campaignCfg.Campaign.TotalBudget)
	fmt.Printf("Test Budget: $%.2f (%.1f%%)\n",
		campaignCfg.Campaign.TotalBudget*campaignCfg.Campaign.TestBudgetPercentage/100,
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
}

// createTestCampaigns creates test campaigns from a YAML configuration
func createTestCampaigns(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Println("Missing YAML file path. Use: fbads optimize create <yaml_file> [--template=campaign.json] [--limit=N] [--batch-size=N] [--dry-run]")
		os.Exit(1)
	}

	yamlPath := args[0]
	templatePath := ""
	limit := 0
	batchSize := 3
	dryRun := false
	priority := "audience"

	// Parse optional flags
	for i := 1; i < len(args); i++ {
		switch {
		case strings.HasPrefix(args[i], "--template="):
			templatePath = strings.TrimPrefix(args[i], "--template=")
		case args[i] == "--template" && i+1 < len(args):
			templatePath = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--limit="):
			fmt.Sscanf(strings.TrimPrefix(args[i], "--limit="), "%d", &limit)
		case args[i] == "--limit" && i+1 < len(args):
			fmt.Sscanf(args[i+1], "%d", &limit)
			i++
		case strings.HasPrefix(args[i], "--batch-size="):
			fmt.Sscanf(strings.TrimPrefix(args[i], "--batch-size="), "%d", &batchSize)
		case args[i] == "--batch-size" && i+1 < len(args):
			fmt.Sscanf(args[i+1], "%d", &batchSize)
			i++
		case args[i] == "--dry-run" || args[i] == "-d":
			dryRun = true
		case strings.HasPrefix(args[i], "--priority="):
			priority = strings.TrimPrefix(args[i], "--priority=")
		case args[i] == "--priority" && i+1 < len(args):
			priority = args[i+1]
			i++
		}
	}

	// Parse YAML configuration
	campaignCfg, err := optimization.ParseYAMLConfig(yamlPath)
	if err != nil {
		fmt.Printf("Error parsing YAML configuration: %v\n", err)
		os.Exit(1)
	}
	
	// Load template if provided
	var templateCampaign *models.CampaignConfig
	if templatePath != "" {
		fmt.Printf("Using campaign template from: %s\n", templatePath)
		// Read the template file
		templateData, err := os.ReadFile(templatePath)
		if err != nil {
			fmt.Printf("Error reading template file: %v\n", err)
			os.Exit(1)
		}

		// Parse the template
		if err := json.Unmarshal(templateData, &templateCampaign); err != nil {
			fmt.Printf("Error parsing template: %v\n", err)
			os.Exit(1)
		}
	}

	// Validate and print configuration details
	fmt.Println("Creating test campaigns from configuration:")
	fmt.Println("Campaign Name:", campaignCfg.Campaign.Name)
	fmt.Printf("Total Budget: $%.2f\n", campaignCfg.Campaign.TotalBudget)
	fmt.Printf("Test Budget Percentage: %.1f%%\n", campaignCfg.Campaign.TestBudgetPercentage)

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

	// Create campaign generator
	generator := optimization.NewCampaignGenerator(campaignCfg, budgetCalc)
	generator.SetLimit(limit)
	generator.SetMaxBatchSize(batchSize)
	generator.SetPriority(priority)
	if templateCampaign != nil {
		generator.SetTemplate(templateCampaign)
	}

	// Generate all combinations
	if err := generator.GenerateAllCombinations(); err != nil {
		fmt.Printf("Error generating campaign combinations: %v\n", err)
		os.Exit(1)
	}

	// Display generation summary
	totalCombinations := generator.TotalCombinations()
	totalBatches := generator.TotalBatches()

	if limit > 0 && limit < totalCombinations {
		fmt.Printf("Generated %d combinations (limited from %d possible)\n",
			totalCombinations, len(campaignCfg.Creatives)*
				(len(campaignCfg.TargetingOptions.Audiences)+len(campaignCfg.TargetingOptions.Placements)))
	} else {
		fmt.Printf("Generated %d combinations\n", totalCombinations)
	}
	fmt.Printf("Batch size: %d, Total batches: %d\n", batchSize, totalBatches)

	// Get budget per campaign
	budgetPerCampaign, err := budgetCalc.GetBudgetPerCampaign(totalCombinations)
	if err != nil {
		fmt.Printf("Error calculating budget per campaign: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Budget per test campaign: $%.2f\n", budgetPerCampaign)

	// Create rate limiter for Facebook API calls
	rateLimiter := optimization.NewRateLimiter()
	rateLimiter.SetRequestInterval(500 * time.Millisecond) // Facebook's rate limit is relatively low

	// Process all batches
	if dryRun {
		fmt.Println("\nDry run mode - showing first batch without creating campaigns:")

		// Just get the first batch for preview
		batch := generator.GetNextBatch()
		for i, combination := range batch {
			facebookCampaign := generator.ConvertToFacebookCampaign(combination)
			fmt.Printf("\nCampaign %d: %s\n", i+1, facebookCampaign.Name)
			fmt.Printf("  Creative: %s\n", combination.Creative.Title)
			if combination.TargetingType == "audience" {
				fmt.Printf("  Audience: %s\n", combination.AudienceName)
			} else {
				fmt.Printf("  Placement: %s (%s)\n", combination.PlacementName, combination.PlacementParams)
			}
			fmt.Printf("  Budget: $%.2f\n", combination.Budget)
			fmt.Printf("  CPM Bid: $%.2f\n", combination.BidAmount)
		}

		fmt.Printf("\nRemaining batches: %d\n", totalBatches-1)
		fmt.Println("\nNo campaigns were created (dry run mode)")
	} else {
		// Create auth client
		authClient := auth.NewFacebookAuth(
			cfg.AppID,
			cfg.AppSecret,
			cfg.AccessToken,
			cfg.APIVersion,
		)

		// Create campaign creator
		campaignCreator := internal_campaign.NewCampaignCreator(authClient, cfg.AccountID)

		// Ask for confirmation before proceeding
		fmt.Printf("\nThis will create %d test campaigns. Proceed? (y/n): ", totalCombinations)
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" && confirm != "yes" && confirm != "Yes" {
			fmt.Println("Campaign creation cancelled.")
			return
		}

		// Create a context with timeout for the entire operation
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		createdCount := 0
		failedCount := 0

		// Process all batches
		for {
			batch := generator.GetNextBatch()
			if len(batch) == 0 {
				break // No more combinations
			}

			fmt.Printf("\nProcessing batch %d/%d (%d campaigns)...\n",
				generator.CurrentBatch, totalBatches, len(batch))

			for i, combination := range batch {
				// Convert to Facebook campaign configuration
				facebookCampaign := generator.ConvertToFacebookCampaign(combination)

				fmt.Printf("[%d/%d] Creating campaign: %s... ",
					createdCount+failedCount+1, totalCombinations, facebookCampaign.Name)
				// Use i to avoid "not used" warning
				_ = i

				// Execute with rate limiting and retries
				err := rateLimiter.Execute(ctx, func() error {
					return campaignCreator.CreateFromConfig(facebookCampaign)
				})

				if err != nil {
					fmt.Printf("FAILED: %v\n", err)
					failedCount++
				} else {
					fmt.Println("SUCCESS")
					createdCount++
				}

				// Check if context was cancelled (timeout or user interrupt)
				select {
				case <-ctx.Done():
					fmt.Printf("\nOperation cancelled: %v\n", ctx.Err())
					return
				default:
					// Continue with next campaign
				}
			}
		}

		// Print final summary
		fmt.Printf("\nCampaign creation completed:\n")
		fmt.Printf("  Successfully created: %d\n", createdCount)
		fmt.Printf("  Failed: %d\n", failedCount)
		fmt.Printf("  Total: %d\n", totalCombinations)

		// For now, provide a placeholder message since we haven't fully implemented the API integration
		if createdCount == 0 && failedCount == 0 {
			fmt.Println("\nNote: Campaign creation functionality will be implemented in the next version.")
			fmt.Println("This command currently simulates the creation process without making API calls.")
		}
	}
}

// updateCampaignCPM updates campaign CPM based on performance data
func updateCampaignCPM(cfg *config.Config, args []string) {
	if len(args) < 1 {
		fmt.Println("Missing campaign IDs. Use: fbads optimize update <campaign_id1,campaign_id2,...> [--max-cpm=N]")
		os.Exit(1)
	}

	campaignIDs := strings.Split(args[0], ",")
	maxCPM := 15.0 // Default max CPM

	// Parse optional flags
	for i := 1; i < len(args); i++ {
		switch {
		case strings.HasPrefix(args[i], "--max-cpm="):
			fmt.Sscanf(strings.TrimPrefix(args[i], "--max-cpm="), "%f", &maxCPM)
		case args[i] == "--max-cpm" && i+1 < len(args):
			fmt.Sscanf(args[i+1], "%f", &maxCPM)
			i++
		}
	}

	fmt.Printf("Processing CPM optimization for %d campaigns\n", len(campaignIDs))
	fmt.Printf("Maximum CPM: $%.2f\n", maxCPM)

	// This is placeholder code for the future implementation
	// Will be implemented in the next version

	// For now, just show placeholders to indicate future functionality

	// TODO: Implement CPM optimization logic with the API client

	for _, campaignID := range campaignIDs {
		fmt.Printf("Campaign %s: CPM optimization will be implemented in the next version\n", campaignID)

		// In a real implementation, we would:
		// 1. Get campaign performance data
		// 2. Calculate optimal CPM
		// 3. Update the campaign's CPM if needed
	}
}

func configureApp(cfg *config.Config, configPath string) {
	fmt.Println("Configuring application...")

	// Simple configuration prompt (to be expanded)
	fmt.Print("Enter Facebook App ID: ")
	fmt.Scanln(&cfg.AppID)

	fmt.Print("Enter Facebook App Secret: ")
	fmt.Scanln(&cfg.AppSecret)

	fmt.Print("Enter Facebook Access Token: ")
	fmt.Scanln(&cfg.AccessToken)

	fmt.Print("Enter Facebook Ad Account ID (without act_ prefix): ")
	fmt.Scanln(&cfg.AccountID)

	// Save configuration
	if err := cfg.SaveConfig(configPath); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration saved successfully!")
}

func startDashboard(cfg *config.Config) {
	// Parse optional port flag
	port := 8080
	if len(os.Args) >= 3 {
		fmt.Sscanf(os.Args[2], "%d", &port)
	}

	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create metrics collector
	metricsCollector := api.NewMetricsCollector(authClient, cfg.AccountID)

	// Create audience analyzer
	audienceAnalyzer := audience.NewAudienceAnalyzer(authClient, cfg.AccountID)

	// Create performance analyzer
	analyzer := api.NewPerformanceAnalyzer(metricsCollector, audienceAnalyzer)

	// Set dashboard directories
	dashboardDir := filepath.Join(cfg.ConfigDir, "dashboard")
	templateDir := filepath.Join(dashboardDir, "templates")
	dataDir := filepath.Join(dashboardDir, "data")

	// Create dashboard
	dashboard := api.NewDashboard(metricsCollector, analyzer, port, templateDir, dataDir)

	// Create dashboard files
	if err := dashboard.CreateDashboardFiles(); err != nil {
		fmt.Printf("Error creating dashboard files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting dashboard on http://localhost:%d\n", port)

	// Start dashboard
	if err := dashboard.Start(); err != nil {
		fmt.Printf("Error starting dashboard: %v\n", err)
		os.Exit(1)
	}
}

// exportCampaign exports a campaign by ID to a configuration file
func exportCampaign(cfg *config.Config, campaignID string, args []string) {
	// Determine output file name
	outputFile := campaignID + ".json"
	if len(args) > 0 {
		outputFile = args[0]
	}

	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create API client
	client := api.NewClient(authClient, cfg.AccountID)

	fmt.Printf("Fetching campaign details for ID: %s\n", campaignID)

	// Get campaign details
	details, err := client.GetCampaignDetails(campaignID)
	if err != nil {
		fmt.Printf("Error fetching campaign details: %v\n", err)
		os.Exit(1)
	}

	// Convert to a campaign configuration
	config := convertToConfig(details)

	// Write to file
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error serializing configuration: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		fmt.Printf("Error writing configuration to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Campaign exported successfully to: %s\n", outputFile)
}

// exportCampaignYAML exports a campaign by ID to a YAML file for optimization
func exportCampaignYAML(cfg *config.Config, campaignID string, args []string) {
	// Set up default export config
	exporterConfig := optimization.DefaultExporterConfig()

	// Determine output file name
	outputFile := campaignID + ".yaml"

	// Parse arguments
	for i := 0; i < len(args); i++ {
		if args[i] == "--budget" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%f", &exporterConfig.TotalBudget)
			i++
		} else if args[i] == "--test-percent" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%f", &exporterConfig.TestBudgetPercentage)
			i++
		} else if args[i] == "--max-cpm" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%f", &exporterConfig.MaxCPM)
			i++
		} else if !strings.HasPrefix(args[i], "--") && i == 0 {
			// First non-flag argument is the output file
			outputFile = args[i]
		}
	}

	// Set output path
	exporterConfig.OutputPath = outputFile

	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create API client
	client := api.NewClient(authClient, cfg.AccountID)

	fmt.Printf("Fetching campaign details for ID: %s\n", campaignID)

	// Get campaign details
	details, err := client.GetCampaignDetails(campaignID)
	if err != nil {
		fmt.Printf("Error fetching campaign details: %v\n", err)
		os.Exit(1)
	}

	// Create exporter
	exporter := optimization.NewExporter(exporterConfig)

	// Export campaign to YAML
	if err := exporter.ExportCampaign(details); err != nil {
		fmt.Printf("Error exporting campaign to YAML: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Campaign exported to YAML for optimization: %s\n", outputFile)
	fmt.Printf("Configuration: Total Budget: $%.2f, Test Budget: %.1f%%, Max CPM: $%.2f\n",
		exporterConfig.TotalBudget,
		exporterConfig.TestBudgetPercentage,
		exporterConfig.MaxCPM)
}

// listPages lists all Facebook Pages accessible with the current access token
func listPages(cfg *config.Config) {
	// Parse flags
	var format string

	// Check for flags
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format", "-f":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		}
	}

	// Set default format
	if format == "" {
		format = "table" // Default to table format
	}

	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create API client
	client := api.NewClient(authClient, cfg.AccountID)

	fmt.Println("Fetching available Facebook Pages...")

	// Get pages
	pages, err := client.GetPages()
	if err != nil {
		fmt.Printf("Error fetching pages: %v\n", err)
		os.Exit(1)
	}

	if len(pages) == 0 {
		fmt.Println("No Facebook Pages found for this access token.")
		fmt.Println("Make sure your access token has the 'pages_show_list' and 'pages_read_engagement' permissions.")
		return
	}

	// Display results based on format
	switch format {
	case "json":
		displayPagesJSON(pages)
	case "csv":
		displayPagesCSV(pages)
	case "table":
		displayPagesTable(pages)
	default:
		fmt.Printf("Unknown format: %s. Supported formats: table, json, csv\n", format)
		os.Exit(1)
	}

	fmt.Printf("\nTotal: %d Facebook Pages\n", len(pages))
	fmt.Println("\nNote: Use the page ID in your campaign configuration's 'page_id' field.")
}

// displayPagesTable displays pages in a formatted table
func displayPagesTable(pages []models.Page) {
	if len(pages) == 0 {
		fmt.Println("No pages found.")
		return
	}

	// Calculate column widths
	idWidth := 20
	nameWidth := 40
	categoryWidth := 25

	// Print header
	fmt.Printf("%-*s | %-*s | %-*s\n",
		idWidth, "PAGE ID",
		nameWidth, "NAME",
		categoryWidth, "CATEGORY")

	// Print separator
	fmt.Printf("%s-+-%s-+-%s\n",
		strings.Repeat("-", idWidth),
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", categoryWidth))

	// Print rows
	for _, page := range pages {
		fmt.Printf("%-*s | %-*s | %-*s\n",
			idWidth, page.ID,
			nameWidth, truncateString(page.Name, nameWidth),
			categoryWidth, truncateString(page.Category, categoryWidth))
	}
}

// displayPagesJSON displays pages in JSON format
func displayPagesJSON(pages []models.Page) {
	// Create a response structure to wrap the pages
	response := struct {
		Pages []models.Page `json:"pages"`
		Count int           `json:"count"`
	}{
		Pages: pages,
		Count: len(pages),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

// displayPagesCSV displays pages in CSV format
func displayPagesCSV(pages []models.Page) {
	// Print header
	fmt.Println("id,name,category")

	// Print rows
	for _, page := range pages {
		fmt.Printf("%s,%s,%s\n",
			page.ID,
			escapeCSV(page.Name),
			escapeCSV(page.Category))
	}
}

// convertToConfig converts campaign details to a configuration
func convertToConfig(details *models.CampaignDetails) *models.CampaignConfig {
	config := &models.CampaignConfig{
		Name:                details.Name,
		Status:              details.Status,
		Objective:           details.ObjectiveType,
		BuyingType:          details.BuyingType,
		SpecialAdCategories: details.SpecialAdCategories,
		BidStrategy:         details.BidStrategy,
		DailyBudget:         details.DailyBudget,
		LifetimeBudget:      details.LifetimeBudget,
		AdSets:              []models.AdSetConfig{},
		Ads:                 []models.AdConfig{},
	}

	// Add start/end times if available
	if !details.StartTime.IsZero() {
		config.StartTime = details.StartTime.Format(time.RFC3339)
	}

	if !details.StopTime.IsZero() {
		config.EndTime = details.StopTime.Format(time.RFC3339)
	}

	// Process AdSets
	for _, adset := range details.AdSets {
		adsetConfig := models.AdSetConfig{
			Name:             adset.Name,
			Status:           adset.Status,
			Targeting:        adset.Targeting,
			OptimizationGoal: adset.OptimizationGoal,
			BillingEvent:     adset.BillingEvent,
			BidAmount:        adset.BidAmount,
		}

		// Add start/end times if available
		if !adset.StartTime.IsZero() {
			adsetConfig.StartTime = adset.StartTime.Format(time.RFC3339)
		}

		if !adset.EndTime.IsZero() {
			adsetConfig.EndTime = adset.EndTime.Format(time.RFC3339)
		}

		config.AdSets = append(config.AdSets, adsetConfig)
	}

	// Process Ads
	for _, ad := range details.Ads {
		adConfig := models.AdConfig{
			Name:   ad.Name,
			Status: ad.Status,
			Creative: models.CreativeConfig{
				Name:         ad.Creative.Title, // Use name field for title value per API requirements
				Body:         ad.Creative.Body,
				ImageURL:     ad.Creative.ImageURL,
				LinkURL:      ad.Creative.LinkURL,
				CallToAction: ad.Creative.CallToActionType,
				PageID:       ad.Creative.PageID,
			},
		}

		config.Ads = append(config.Ads, adConfig)
	}

	return config
}

// updateCampaign handles updating an existing campaign
func updateCampaign(cfg *config.Config) {
	// Parse flags
	var (
		campaignID     string
		status         string
		name           string
		dailyBudget    float64
		lifetimeBudget float64
		bidStrategy    string
		jsonFile       string
	)

	// Skip the first two args (fbads update)
	args := os.Args[2:]

	// Handle flags
	for i := 0; i < len(args); i++ {
		switch {
		case strings.HasPrefix(args[i], "--id="):
			campaignID = strings.TrimPrefix(args[i], "--id=")
		case args[i] == "--id" && i+1 < len(args):
			campaignID = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--status="):
			status = strings.TrimPrefix(args[i], "--status=")
		case args[i] == "--status" && i+1 < len(args):
			status = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--name="):
			name = strings.TrimPrefix(args[i], "--name=")
		case args[i] == "--name" && i+1 < len(args):
			name = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--daily-budget="):
			fmt.Sscanf(strings.TrimPrefix(args[i], "--daily-budget="), "%f", &dailyBudget)
		case args[i] == "--daily-budget" && i+1 < len(args):
			fmt.Sscanf(args[i+1], "%f", &dailyBudget)
			i++
		case strings.HasPrefix(args[i], "--lifetime-budget="):
			fmt.Sscanf(strings.TrimPrefix(args[i], "--lifetime-budget="), "%f", &lifetimeBudget)
		case args[i] == "--lifetime-budget" && i+1 < len(args):
			fmt.Sscanf(args[i+1], "%f", &lifetimeBudget)
			i++
		case strings.HasPrefix(args[i], "--bid-strategy="):
			bidStrategy = strings.TrimPrefix(args[i], "--bid-strategy=")
		case args[i] == "--bid-strategy" && i+1 < len(args):
			bidStrategy = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--file="):
			jsonFile = strings.TrimPrefix(args[i], "--file=")
		case args[i] == "--file" && i+1 < len(args):
			jsonFile = args[i+1]
			i++
		}
	}

	// Check if at least campaign ID is provided
	if campaignID == "" {
		fmt.Println("Error: Campaign ID is required")
		fmt.Println("Usage: fbads update --id=CAMPAIGN_ID [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  --id=ID                   Campaign ID to update (required)")
		fmt.Println("  --status=STATUS           New status (ACTIVE, PAUSED, ARCHIVED)")
		fmt.Println("  --name=NAME               New campaign name")
		fmt.Println("  --daily-budget=BUDGET     New daily budget (e.g., 50.00)")
		fmt.Println("  --lifetime-budget=BUDGET  New lifetime budget (e.g., 1000.00)")
		fmt.Println("  --bid-strategy=STRATEGY   New bid strategy (e.g., LOWEST_COST_WITHOUT_CAP)")
		fmt.Println("  --file=FILE               JSON file with update parameters")
		os.Exit(1)
	}

	// Check if at least one update parameter is provided
	if status == "" && name == "" && dailyBudget == 0 && lifetimeBudget == 0 &&
		bidStrategy == "" && jsonFile == "" {
		fmt.Println("Error: At least one update parameter must be provided")
		fmt.Println("Usage: fbads update --id=CAMPAIGN_ID [options]")
		os.Exit(1)
	}

	// Create the Facebook auth object
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create API client
	client := api.NewClient(authClient, cfg.AccountID)

	// Build the update parameters
	params := url.Values{}

	// If a JSON file is provided, load update parameters from it
	if jsonFile != "" {
		fileParams, err := loadParamsFromFile(jsonFile)
		if err != nil {
			fmt.Printf("Error loading parameters from file: %v\n", err)
			os.Exit(1)
		}

		// Merge file parameters with params
		for key, values := range fileParams {
			for _, value := range values {
				params.Add(key, value)
			}
		}
	}

	// Add command-line parameters (these override file parameters)
	if status != "" {
		validStatuses := map[string]bool{"ACTIVE": true, "PAUSED": true, "ARCHIVED": true}
		if !validStatuses[strings.ToUpper(status)] {
			fmt.Printf("Invalid status: %s. Must be one of: ACTIVE, PAUSED, ARCHIVED\n", status)
			os.Exit(1)
		}
		params.Set("status", strings.ToUpper(status))
	}

	if name != "" {
		params.Set("name", name)
	}

	if dailyBudget > 0 {
		// Convert to cents as required by the API
		params.Set("daily_budget", fmt.Sprintf("%d", int(dailyBudget*100)))
	}

	if lifetimeBudget > 0 {
		// Convert to cents as required by the API
		params.Set("lifetime_budget", fmt.Sprintf("%d", int(lifetimeBudget*100)))
	}

	if bidStrategy != "" {
		params.Set("bid_strategy", bidStrategy)
	}

	// Verify the campaign exists before updating
	fmt.Printf("Verifying campaign %s exists...\n", campaignID)
	_, verifyErr := client.GetCampaignDetails(campaignID)
	if verifyErr != nil {
		fmt.Printf("Error: Campaign not found or cannot be accessed: %v\n", verifyErr)
		fmt.Println("Please check that the campaign ID is correct and you have permission to access it.")
		os.Exit(1)
	}

	// Make the API call to update the campaign
	fmt.Printf("Updating campaign %s with parameters: %v\n", campaignID, params)
	updateErr := client.UpdateCampaign(campaignID, params)
	if updateErr != nil {
		fmt.Printf("Error updating campaign: %v\n", updateErr)
		os.Exit(1)
	}

	fmt.Printf("Campaign %s updated successfully\n", campaignID)
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

// duplicateCampaign handles duplicating a campaign with all its internals
func duplicateCampaign(cfg *config.Config, campaignID string, args []string) {
	// Parse flags
	var (
		campaignName string
		status       string = "PAUSED" // Default to PAUSED for safety
		startDateStr string
		endDateStr   string
		budgetFactor float64 = 1.0 // Default to same budget
		dryRun       bool
	)

	// Handle flags
	for i := 0; i < len(args); i++ {
		switch {
		case strings.HasPrefix(args[i], "--name="):
			campaignName = strings.TrimPrefix(args[i], "--name=")
		case args[i] == "--name" && i+1 < len(args):
			campaignName = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--status="):
			status = strings.TrimPrefix(args[i], "--status=")
		case args[i] == "--status" && i+1 < len(args):
			status = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--start="):
			startDateStr = strings.TrimPrefix(args[i], "--start=")
		case args[i] == "--start" && i+1 < len(args):
			startDateStr = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--end="):
			endDateStr = strings.TrimPrefix(args[i], "--end=")
		case args[i] == "--end" && i+1 < len(args):
			endDateStr = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--budget-factor="):
			fmt.Sscanf(strings.TrimPrefix(args[i], "--budget-factor="), "%f", &budgetFactor)
		case args[i] == "--budget-factor" && i+1 < len(args):
			fmt.Sscanf(args[i+1], "%f", &budgetFactor)
			i++
		case args[i] == "--dry-run" || args[i] == "-d":
			dryRun = true
		}
	}

	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create API client
	client := api.NewClient(authClient, cfg.AccountID)

	fmt.Printf("Fetching campaign details for ID: %s\n", campaignID)

	// Get campaign details
	details, err := client.GetCampaignDetails(campaignID)
	if err != nil {
		fmt.Printf("Error fetching campaign details: %v\n", err)
		os.Exit(1)
	}

	// If no custom name provided, create a default name
	if campaignName == "" {
		campaignName = "Copy of " + details.Name
	}

	// Convert to a campaign configuration
	campaignConfig := convertToConfig(details)

	// For duplication, we need to ensure we're not carrying over any IDs
	// The Create function will assign new IDs

	// Remove any unsupported fields from creatives based on recent API changes
	// The Facebook API error shows that image_url is no longer supported in link_data

	// Update the campaign config with the new parameters
	campaignConfig.Name = campaignName
	campaignConfig.Status = status

	// Parse and update dates if provided
	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			fmt.Printf("Invalid start date format: %v\n", err)
			os.Exit(1)
		}
		campaignConfig.StartTime = startDate.Format(time.RFC3339)
	}

	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			fmt.Printf("Invalid end date format: %v\n", err)
			os.Exit(1)
		}
		campaignConfig.EndTime = endDate.Format(time.RFC3339)
	}

	// Fix budget values: when retrieved from Facebook, budgets are in cents
	// but the CampaignConfig expects dollars for display
	if campaignConfig.DailyBudget > 0 {
		// Convert from cents to dollars (e.g., 2000 cents -> $20.00)
		campaignConfig.DailyBudget = campaignConfig.DailyBudget / 100
	}

	if campaignConfig.LifetimeBudget > 0 {
		// Convert from cents to dollars (e.g., 2000 cents -> $20.00)
		campaignConfig.LifetimeBudget = campaignConfig.LifetimeBudget / 100
	}

	// Apply budget factor after the conversion
	if budgetFactor != 1.0 {
		if campaignConfig.DailyBudget > 0 {
			campaignConfig.DailyBudget = campaignConfig.DailyBudget * budgetFactor
		}
		if campaignConfig.LifetimeBudget > 0 {
			campaignConfig.LifetimeBudget = campaignConfig.LifetimeBudget * budgetFactor
		}
	}

	// Clear any ID fields from the AdSets and Ads to ensure new ones are created
	for i := range campaignConfig.AdSets {
		// Update ad set names to indicate they're copies
		if !strings.HasPrefix(campaignConfig.AdSets[i].Name, "Copy of ") {
			campaignConfig.AdSets[i].Name = "Copy of " + campaignConfig.AdSets[i].Name
		}
		// Set the status to match the campaign
		campaignConfig.AdSets[i].Status = status
	}

	for i := range campaignConfig.Ads {
		// Update ad names to indicate they're copies
		if !strings.HasPrefix(campaignConfig.Ads[i].Name, "Copy of ") {
			campaignConfig.Ads[i].Name = "Copy of " + campaignConfig.Ads[i].Name
		}
		// Set the status to match the campaign
		campaignConfig.Ads[i].Status = status

		// Remove ImageURL field which is no longer supported by the Facebook API
		// This fixes the error "The field image_url is not supported in the field link_data of object_story_spec"
		campaignConfig.Ads[i].Creative.ImageURL = ""

		// Ensure the LinkURL is not empty
		if campaignConfig.Ads[i].Creative.LinkURL == "" {
			fmt.Println("Warning: Link URL is empty in ad creative. Setting a default link to prevent API error.")
			campaignConfig.Ads[i].Creative.LinkURL = "https://corespirit.com/funnels/pract"
		}
	}

	// Print configuration summary
	fmt.Println("\nDuplicated Campaign Configuration Summary:")
	printCampaignConfigSummary(campaignConfig)

	// If dry run, just print configuration summary and exit
	if dryRun {
		fmt.Println("\nDry run: No campaigns will be created.")
		return
	}

	// Ask for confirmation
	fmt.Print("\nDo you want to create this duplicated campaign? (y/n): ")
	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" && confirm != "yes" && confirm != "Yes" {
		fmt.Println("Campaign duplication cancelled.")
		return
	}

	// Create campaign creator
	creator := internal_campaign.NewCampaignCreator(authClient, cfg.AccountID)

	fmt.Println("Creating duplicated campaign...")

	// Create the campaign
	err = creator.CreateFromConfig(campaignConfig)
	if err != nil {
		fmt.Printf("Error creating duplicated campaign: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Campaign duplicated successfully!")
}

// handleStatistics processes statistics subcommands
func handleStatistics(cfg *config.Config, subCmd string, args []string) {
	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create metrics collector
	metricsCollector := api.NewMetricsCollector(authClient, cfg.AccountID)

	// Set default storage directory
	statsDir := filepath.Join(cfg.ConfigDir, "stats")

	// Create statistics manager
	statsManager := api.NewStatisticsManager(metricsCollector, api.StorageTypeFile, statsDir)

	// Parse common flags
	var (
		startDateStr string
		endDateStr   string
		campaignID   string
		outputFile   string
		days         int    = 30     // Default to 30 days
		format       string = "json" // Default format
	)

	// Process flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--start", "-s":
			if i+1 < len(args) {
				startDateStr = args[i+1]
				i++
			}
		case "--end", "-e":
			if i+1 < len(args) {
				endDateStr = args[i+1]
				i++
			}
		case "--days", "-d":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &days)
				i++
			}
		case "--campaign", "-c":
			if i+1 < len(args) {
				campaignID = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "--format", "-f":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		}
	}

	// Set default date range if not specified
	var startDate, endDate time.Time
	var err error

	if startDateStr == "" {
		// Default start date (30 days ago or as specified by --days)
		startDate = time.Now().AddDate(0, 0, -days)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			fmt.Printf("Invalid start date format: %v\n", err)
			fmt.Println("Date format should be YYYY-MM-DD")
			os.Exit(1)
		}
	}

	if endDateStr == "" {
		// Default end date (yesterday)
		endDate = time.Now().AddDate(0, 0, -1)
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			fmt.Printf("Invalid end date format: %v\n", err)
			fmt.Println("Date format should be YYYY-MM-DD")
			os.Exit(1)
		}
	}

	// Process subcommand
	switch subCmd {
	case "collect":
		collectStatistics(statsManager, startDate, endDate)
	case "analyze":
		analyzeStatistics(statsManager, startDate, endDate, campaignID, format)
	case "export":
		if outputFile == "" {
			// Default output file name
			outputFile = fmt.Sprintf("stats_export_%s_to_%s.csv",
				startDate.Format("2006-01-02"),
				endDate.Format("2006-01-02"))
		}
		exportStatistics(statsManager, startDate, endDate, outputFile)
	case "validate":
		validateCampaignData(statsManager, startDate, endDate, campaignID, format)
	default:
		fmt.Printf("Unknown stats subcommand: %s\n", subCmd)
		fmt.Println("Available subcommands: collect, analyze, export, validate")
		os.Exit(1)
	}
}

// collectStatistics collects metrics for the given date range
func collectStatistics(statsManager *api.StatisticsManager, startDate, endDate time.Time) {
	fmt.Printf("Collecting campaign statistics from %s to %s...\n",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	// Process one day at a time to get daily statistics
	current := startDate
	var collectErrors []string

	for !current.After(endDate) {
		// Create time range for the day
		timeRange := api.TimeRange{
			Since: current.Format("2006-01-02"),
			Until: current.Format("2006-01-02"),
		}

		fmt.Printf("Collecting data for %s...\n", current.Format("2006-01-02"))
		err := statsManager.CollectAndStoreStatistics(timeRange)
		if err != nil {
			fmt.Printf("Error collecting data for %s: %v\n", current.Format("2006-01-02"), err)
			collectErrors = append(collectErrors, fmt.Sprintf("%s: %v", current.Format("2006-01-02"), err))
		}

		// Move to next day
		current = current.AddDate(0, 0, 1)
	}

	if len(collectErrors) > 0 {
		fmt.Println("\nWarning: Some data collection operations failed:")
		for _, errMsg := range collectErrors {
			fmt.Printf("  - %s\n", errMsg)
		}
		fmt.Println("\nPartial data may still be available for analysis.")
	} else {
		fmt.Println("\nStatistics collection completed successfully!")
	}
}

// analyzeStatistics analyzes campaign performance for the given date range
func analyzeStatistics(statsManager *api.StatisticsManager, startDate, endDate time.Time, campaignID, format string) {
	if campaignID != "" {
		fmt.Printf("Analyzing statistics for campaign %s from %s to %s...\n",
			campaignID,
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"))

		// Get campaign-specific statistics
		stats, err := statsManager.GetCampaignStatistics(campaignID, startDate, endDate)
		if err != nil {
			fmt.Printf("Error analyzing campaign statistics: %v\n", err)
			os.Exit(1)
		}

		if len(stats) == 0 {
			fmt.Println("No statistics found for the specified campaign and date range.")
			return
		}

		// Display statistics based on format
		switch format {
		case "json":
			displayStatisticsJSON(stats)
		case "table":
			displayCampaignStatisticsTable(stats)
		default:
			fmt.Printf("Unsupported format: %s. Using table format.\n", format)
			displayCampaignStatisticsTable(stats)
		}

	} else {
		fmt.Printf("Analyzing statistics for all campaigns from %s to %s...\n",
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"))

		// Perform full statistical analysis
		analysis, err := statsManager.AnalyzeStatistics(startDate, endDate)
		if err != nil {
			fmt.Printf("Error analyzing statistics: %v\n", err)
			os.Exit(1)
		}

		if len(analysis.CampaignStats) == 0 {
			fmt.Println("No statistics found for the specified date range.")
			return
		}

		// Display statistics based on format
		switch format {
		case "json":
			displayAnalysisJSON(analysis)
		case "table":
			displayAnalysisTable(analysis)
		default:
			fmt.Printf("Unsupported format: %s. Using table format.\n", format)
			displayAnalysisTable(analysis)
		}
	}
}

// displayStatisticsJSON displays campaign performance data in JSON format
func displayStatisticsJSON(stats []utils.CampaignPerformance) {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding statistics to JSON: %v\n", err)
		return
	}

	fmt.Println(string(data))
}

// displayCampaignStatisticsTable displays campaign performance data in a table format
func displayCampaignStatisticsTable(stats []utils.CampaignPerformance) {
	if len(stats) == 0 {
		fmt.Println("No statistics available.")
		return
	}

	// Print header
	fmt.Printf("%-10s | %-10s | %-10s | %-8s | %-6s | %-8s | %-8s | %-8s | %-8s\n",
		"DATE", "IMPRESSIONS", "CLICKS", "CTR (%)", "SPEND", "CPM", "CPC", "CONV", "ROAS")

	// Print separator
	fmt.Printf("%s-+-%s-+-%s-+-%s-+-%s-+-%s-+-%s-+-%s-+-%s\n",
		strings.Repeat("-", 10),
		strings.Repeat("-", 10),
		strings.Repeat("-", 10),
		strings.Repeat("-", 8),
		strings.Repeat("-", 6),
		strings.Repeat("-", 8),
		strings.Repeat("-", 8),
		strings.Repeat("-", 8),
		strings.Repeat("-", 8))

	// Print data rows
	totalImpressions := 0
	totalClicks := 0
	totalSpend := 0.0
	totalConversions := 0

	// Sort by date
	sortPerformancesByDate(stats)

	for _, stat := range stats {
		fmt.Printf("%-10s | %-10d | %-10d | %-8.2f | %-6.2f | %-8.2f | %-8.2f | %-8d | %-8.2f\n",
			stat.LastUpdated.Format("2006-01-02"),
			stat.Impressions,
			stat.Clicks,
			stat.CTR,
			stat.Spend,
			stat.CPM,
			stat.CPC,
			stat.Conversions,
			stat.ROAS)

		totalImpressions += stat.Impressions
		totalClicks += stat.Clicks
		totalSpend += stat.Spend
		totalConversions += stat.Conversions
	}

	// Print totals
	fmt.Printf("%s-+-%s-+-%s-+-%s-+-%s-+-%s-+-%s-+-%s-+-%s\n",
		strings.Repeat("-", 10),
		strings.Repeat("-", 10),
		strings.Repeat("-", 10),
		strings.Repeat("-", 8),
		strings.Repeat("-", 6),
		strings.Repeat("-", 8),
		strings.Repeat("-", 8),
		strings.Repeat("-", 8),
		strings.Repeat("-", 8))

	// Calculate averages for totals
	var avgCTR, avgCPM, avgCPC, avgROAS float64

	if totalImpressions > 0 {
		avgCTR = float64(totalClicks) / float64(totalImpressions) * 100
		avgCPM = totalSpend / float64(totalImpressions) * 1000
	}

	if totalClicks > 0 {
		avgCPC = totalSpend / float64(totalClicks)
	}

	if totalSpend > 0 && totalConversions > 0 {
		// Simplified ROAS calculation
		avgOrderValue := 50.0 // Example value, same as in the analyzer
		avgROAS = float64(totalConversions) * avgOrderValue / totalSpend
	}

	fmt.Printf("%-10s | %-10d | %-10d | %-8.2f | %-6.2f | %-8.2f | %-8.2f | %-8d | %-8.2f\n",
		"TOTAL",
		totalImpressions,
		totalClicks,
		avgCTR,
		totalSpend,
		avgCPM,
		avgCPC,
		totalConversions,
		avgROAS)
}

// sortPerformancesByDate sorts campaign performances by date
func sortPerformancesByDate(performances []utils.CampaignPerformance) {
	for i := 0; i < len(performances); i++ {
		for j := i + 1; j < len(performances); j++ {
			if performances[j].LastUpdated.Before(performances[i].LastUpdated) {
				performances[i], performances[j] = performances[j], performances[i]
			}
		}
	}
}

// displayAnalysisJSON displays analysis results in JSON format
func displayAnalysisJSON(analysis *api.AggregateStatistics) {
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding analysis to JSON: %v\n", err)
		return
	}

	fmt.Println(string(data))
}

// displayAnalysisTable displays analysis results in table format
func displayAnalysisTable(analysis *api.AggregateStatistics) {
	// Print summary header
	fmt.Println("Campaign Performance Summary")
	fmt.Printf("Date Range: %s to %s\n\n",
		analysis.StartDate.Format("2006-01-02"),
		analysis.EndDate.Format("2006-01-02"))

	// Print overall statistics
	fmt.Printf("Total Impressions: %d\n", analysis.TotalImpressions)
	fmt.Printf("Total Clicks: %d\n", analysis.TotalClicks)
	fmt.Printf("Total Spend: $%.2f\n", analysis.TotalSpend)
	fmt.Printf("Total Conversions: %d\n", analysis.TotalConversions)
	fmt.Printf("Average CTR: %.2f%%\n", analysis.AvgCTR)
	fmt.Printf("Average CPM: $%.2f\n", analysis.AvgCPM)
	fmt.Printf("Average CPC: $%.2f\n", analysis.AvgCPC)
	if analysis.TotalConversions > 0 {
		fmt.Printf("Average CPA: $%.2f\n", analysis.AvgCPA)
	}

	// Print trend summary if available
	if analysis.TrendImpressions != nil && len(analysis.TrendImpressions.Values) > 1 {
		fmt.Printf("\nTrends (%% change over period):\n")

		if analysis.TrendImpressions != nil {
			fmt.Printf("  Impressions: %.1f%%\n", analysis.TrendImpressions.Change)
		}
		if analysis.TrendClicks != nil {
			fmt.Printf("  Clicks: %.1f%%\n", analysis.TrendClicks.Change)
		}
		if analysis.TrendCTR != nil {
			fmt.Printf("  CTR: %.1f%%\n", analysis.TrendCTR.Change)
		}
		if analysis.TrendCPM != nil {
			fmt.Printf("  CPM: %.1f%%\n", analysis.TrendCPM.Change)
		}
		if analysis.TrendConversions != nil {
			fmt.Printf("  Conversions: %.1f%%\n", analysis.TrendConversions.Change)
		}
	}

	// Print campaign-specific statistics
	fmt.Printf("\nCampaign Performance Breakdown:\n")
	fmt.Printf("%-20s | %-10s | %-10s | %-8s | %-8s | %-8s | %-8s | %-8s\n",
		"CAMPAIGN", "IMPRESSIONS", "CLICKS", "CTR (%)", "SPEND", "CPM", "CPC", "CONV")

	// Print separator
	fmt.Printf("%s-+-%s-+-%s-+-%s-+-%s-+-%s-+-%s-+-%s\n",
		strings.Repeat("-", 20),
		strings.Repeat("-", 10),
		strings.Repeat("-", 10),
		strings.Repeat("-", 8),
		strings.Repeat("-", 8),
		strings.Repeat("-", 8),
		strings.Repeat("-", 8),
		strings.Repeat("-", 8))

	// Print data rows
	for _, campaign := range analysis.CampaignStats {
		// Truncate campaign name if too long
		name := campaign.Name
		if len(name) > 17 {
			name = name[:14] + "..."
		}

		fmt.Printf("%-20s | %-10d | %-10d | %-8.2f | %-8.2f | %-8.2f | %-8.2f | %-8d\n",
			name,
			campaign.TotalImpressions,
			campaign.TotalClicks,
			campaign.AvgCTR,
			campaign.TotalSpend,
			campaign.AvgCPM,
			campaign.AvgCPC,
			campaign.TotalConversions)
	}
}

// exportStatistics exports campaign statistics to a CSV file
func exportStatistics(statsManager *api.StatisticsManager, startDate, endDate time.Time, outputFile string) {
	fmt.Printf("Exporting statistics from %s to %s...\n",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	// Analyze statistics
	analysis, err := statsManager.AnalyzeStatistics(startDate, endDate)
	if err != nil {
		fmt.Printf("Error analyzing statistics: %v\n", err)
		os.Exit(1)
	}

	if len(analysis.CampaignStats) == 0 {
		fmt.Println("No statistics found for the specified date range.")
		return
	}

	// Export to CSV
	if err := statsManager.ExportStatisticsCSV(analysis, outputFile); err != nil {
		fmt.Printf("Error exporting statistics to CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Statistics exported successfully to: %s\n", outputFile)
}

// validateCampaignData validates campaign performance data against thresholds
func validateCampaignData(statsManager *api.StatisticsManager, startDate, endDate time.Time, campaignID, format string) {
	// Create an optimization validator with default thresholds
	validator := optimization.NewPerformanceValidator()

	// Retrieve campaign data for validation
	var validationResults map[string]optimization.ValidationResult

	if campaignID != "" {
		// Validate a specific campaign
		fmt.Printf("Validating statistics for campaign %s from %s to %s...\n",
			campaignID,
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"))

		stats, err := statsManager.GetCampaignStatistics(campaignID, startDate, endDate)
		if err != nil {
			fmt.Printf("Error getting campaign statistics: %v\n", err)
			os.Exit(1)
		}

		if len(stats) == 0 {
			fmt.Println("No statistics found for the specified campaign and date range.")
			return
		}

		result := validator.ValidateCampaignData(campaignID, stats)
		validationResults = map[string]optimization.ValidationResult{
			campaignID: result,
		}
	} else {
		// Validate all campaigns
		fmt.Printf("Validating statistics for all campaigns from %s to %s...\n",
			startDate.Format("2006-01-02"),
			endDate.Format("2006-01-02"))

		allStats, err := statsManager.GetAllCampaignStatistics(startDate, endDate)
		if err != nil {
			fmt.Printf("Error getting campaign statistics: %v\n", err)
			os.Exit(1)
		}

		if len(allStats) == 0 {
			fmt.Println("No statistics found for the specified date range.")
			return
		}

		validationResults = validator.ValidateCampaignsData(allStats)
	}

	// Output results based on format
	switch format {
	case "json":
		displayValidationResultsJSON(validationResults)
	case "table":
		displayValidationResultsTable(validationResults)
	default:
		fmt.Printf("Unsupported format: %s. Using table format.\n", format)
		displayValidationResultsTable(validationResults)
	}
}

// displayValidationResultsJSON displays validation results in JSON format
func displayValidationResultsJSON(results map[string]optimization.ValidationResult) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding validation results to JSON: %v\n", err)
		return
	}

	fmt.Println(string(data))
}

// displayValidationResultsTable displays validation results in table format
func displayValidationResultsTable(results map[string]optimization.ValidationResult) {
	// Count valid and invalid campaigns
	var validCount, invalidCount int

	// Print header
	fmt.Println("Campaign Data Validation Results")
	fmt.Println("-------------------------------")
	fmt.Println("")

	// Group campaigns by validity
	validCampaigns := make([]optimization.ValidationResult, 0)
	invalidCampaigns := make([]optimization.ValidationResult, 0)

	for _, result := range results {
		if result.IsValid {
			validCampaigns = append(validCampaigns, result)
			validCount++
		} else {
			invalidCampaigns = append(invalidCampaigns, result)
			invalidCount++
		}
	}

	// Print summary
	fmt.Printf("Total campaigns: %d\n", len(results))
	fmt.Printf("Valid campaigns: %d\n", validCount)
	fmt.Printf("Invalid campaigns: %d\n", invalidCount)
	fmt.Println("")

	// Print valid campaigns
	if len(validCampaigns) > 0 {
		fmt.Println("Valid Campaigns:")
		fmt.Printf("%-20s | %-10s | %-10s | %-10s | %-10s | %-10s\n",
			"CAMPAIGN", "IMPRESSIONS", "CLICKS", "SPEND", "RUNTIME", "DATA POINTS")

		// Print separator
		fmt.Printf("%s-+-%s-+-%s-+-%s-+-%s-+-%s\n",
			strings.Repeat("-", 20),
			strings.Repeat("-", 10),
			strings.Repeat("-", 10),
			strings.Repeat("-", 10),
			strings.Repeat("-", 10),
			strings.Repeat("-", 10))

		for _, result := range validCampaigns {
			// If campaign name is available, use it; otherwise use ID
			campaignName := result.CampaignID
			if len(results) > 0 {
				campaignName = truncateString(campaignName, 17)
			}

			fmt.Printf("%-20s | %-10d | %-10d | $%-9.2f | %-10s | %-10d\n",
				campaignName,
				result.Metrics.TotalImpressions,
				result.Metrics.TotalClicks,
				result.Metrics.TotalSpend,
				formatDuration(result.RunningTime),
				result.DataPoints)
		}
		fmt.Println("")
	}

	// Print invalid campaigns
	if len(invalidCampaigns) > 0 {
		fmt.Println("Invalid Campaigns:")
		fmt.Printf("%-20s | %-10s | %-10s | %-10s | %-10s | %-35s\n",
			"CAMPAIGN", "IMPRESSIONS", "CLICKS", "SPEND", "RUNTIME", "REASONS")

		// Print separator
		fmt.Printf("%s-+-%s-+-%s-+-%s-+-%s-+-%s\n",
			strings.Repeat("-", 20),
			strings.Repeat("-", 10),
			strings.Repeat("-", 10),
			strings.Repeat("-", 10),
			strings.Repeat("-", 10),
			strings.Repeat("-", 35))

		for _, result := range invalidCampaigns {
			// If campaign name is available, use it; otherwise use ID
			campaignName := result.CampaignID
			if len(results) > 0 {
				campaignName = truncateString(campaignName, 17)
			}

			// Format reasons
			reasonsText := ""
			if len(result.Reasons) > 0 {
				reasonsText = truncateString(result.Reasons[0], 32)
				if len(result.Reasons) > 1 {
					reasonsText += fmt.Sprintf(" (+%d more)", len(result.Reasons)-1)
				}
			}

			fmt.Printf("%-20s | %-10d | %-10d | $%-9.2f | %-10s | %-35s\n",
				campaignName,
				result.Metrics.TotalImpressions,
				result.Metrics.TotalClicks,
				result.Metrics.TotalSpend,
				formatDuration(result.RunningTime),
				reasonsText)
		}
		fmt.Println("")

		// Print wait recommendations
		waitRecommendations := false
		for _, result := range invalidCampaigns {
			if result.RecommendWait {
				if !waitRecommendations {
					fmt.Println("Wait Recommendations:")
					waitRecommendations = true
				}
				fmt.Printf("- Campaign %s: Wait approximately %s more\n",
					result.CampaignID,
					formatDuration(result.WaitTimeNeeded))
			}
		}

		if waitRecommendations {
			fmt.Println("\nNote: Some campaigns need more time to gather sufficient data for optimization decisions.")
			fmt.Println("     Consider waiting before making optimization changes.")
		}
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}

	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}

	return fmt.Sprintf("%dm", minutes)
}

// deleteCampaign deletes a campaign by ID
func deleteCampaign(cfg *config.Config, campaignID string) {
	// Create auth client
	authClient := auth.NewFacebookAuth(
		cfg.AppID,
		cfg.AppSecret,
		cfg.AccessToken,
		cfg.APIVersion,
	)

	// Create API client
	client := api.NewClient(authClient, cfg.AccountID)

	// Verify the campaign exists before deleting
	fmt.Printf("Verifying campaign %s exists...\n", campaignID)
	campaign, verifyErr := client.GetCampaignDetails(campaignID)
	if verifyErr != nil {
		fmt.Printf("Error: Campaign not found or cannot be accessed: %v\n", verifyErr)
		fmt.Println("Please check that the campaign ID is correct and you have permission to access it.")
		os.Exit(1)
	}

	fmt.Printf("Found campaign: %s (Status: %s)\n", campaign.Name, campaign.Status)

	// Ask for confirmation before proceeding
	fmt.Printf("\nWARNING: This will permanently delete the campaign. This action cannot be undone.\n")
	fmt.Print("Are you sure you want to delete this campaign? (y/n): ")
	var confirm string
	fmt.Scanln(&confirm)
	
	if confirm != "y" && confirm != "Y" && confirm != "yes" && confirm != "Yes" {
		fmt.Println("Campaign deletion cancelled.")
		return
	}

	// Delete the campaign
	fmt.Printf("Deleting campaign %s...\n", campaignID)
	err := client.DeleteCampaign(campaignID)
	if err != nil {
		fmt.Printf("Error deleting campaign: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Campaign %s deleted successfully\n", campaignID)
}

func printUsage() {
	fmt.Println("Usage: fbads <command> [arguments]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("")
	fmt.Println("  list [options]           List all campaigns")
	fmt.Println("    --limit, -l <num>      Limit the number of results (default: 10)")
	fmt.Println("    --status, -s <status>  Filter by status (ACTIVE, PAUSED, etc.)")
	fmt.Println("    --format, -f <format>  Output format (table, json, csv)")
	fmt.Println("")
	fmt.Println("  create <config_file>     Create a new campaign from configuration")
	fmt.Println("    --dry-run, -d          Preview the campaign without creating it")
	fmt.Println("")
	fmt.Println("  update                   Update an existing campaign")
	fmt.Println("    --id=ID                Campaign ID to update (required)")
	fmt.Println("    --status=STATUS        New status (ACTIVE, PAUSED, ARCHIVED)")
	fmt.Println("    --name=NAME            New campaign name")
	fmt.Println("    --daily-budget=BUDGET  New daily budget (e.g., 50.00)")
	fmt.Println("    --lifetime-budget=BUDGET  New lifetime budget (e.g., 1000.00)")
	fmt.Println("    --bid-strategy=STRATEGY   New bid strategy (e.g., LOWEST_COST_WITHOUT_CAP)")
	fmt.Println("    --file=FILE            JSON file with update parameters")
	fmt.Println("")
	fmt.Println("  delete <campaign_id>     Delete a campaign by ID")
	fmt.Println("")
	fmt.Println("  duplicate <campaign_id>  Duplicate an existing campaign with all its internals")
	fmt.Println("    --name=NAME            Name for the duplicated campaign (defaults to 'Copy of [original]')")
	fmt.Println("    --status=STATUS        Status for the duplicated campaign (default: PAUSED)")
	fmt.Println("    --start=YYYY-MM-DD     New start date for the duplicated campaign")
	fmt.Println("    --end=YYYY-MM-DD       New end date for the duplicated campaign")
	fmt.Println("    --budget-factor=X      Multiply budget by factor X (e.g., 1.5)")
	fmt.Println("    --dry-run, -d          Preview without creating the duplicate")
	fmt.Println("")
	fmt.Println("  export <campaign_id> [output_file]")
	fmt.Println("                           Export campaign to JSON configuration file")
	fmt.Println("")
	fmt.Println("  exportyaml <campaign_id> [output_file]")
	fmt.Println("                           Export campaign to YAML for optimization testing")
	fmt.Println("    --budget <amount>      Set the total budget for testing (default: 1000.00)")
	fmt.Println("    --test-percent <pct>   Set the test budget percentage (default: 20)")
	fmt.Println("    --max-cpm <amount>     Set the maximum CPM for bidding (default: 15.00)")
	fmt.Println("")
	fmt.Println("  pages                    List Facebook Pages available for the API token")
	fmt.Println("")
	fmt.Println("  stats <subcommand> [args] Campaign statistics analysis")
	fmt.Println("    - collect              Collect performance statistics")
	fmt.Println("      --start, -s <date>    Start date (YYYY-MM-DD)")
	fmt.Println("      --end, -e <date>      End date (YYYY-MM-DD)")
	fmt.Println("      --days, -d <num>      Number of days back from today (default: 30)")
	fmt.Println("    - analyze              Analyze campaign statistics")
	fmt.Println("      --start, -s <date>    Start date (YYYY-MM-DD)")
	fmt.Println("      --end, -e <date>      End date (YYYY-MM-DD)")
	fmt.Println("      --days, -d <num>      Number of days back from today (default: 30)")
	fmt.Println("      --campaign, -c <id>   Specific campaign to analyze (optional)")
	fmt.Println("      --format, -f <fmt>    Output format: json or table (default: json)")
	fmt.Println("    - export               Export campaign statistics to CSV")
	fmt.Println("      --start, -s <date>    Start date (YYYY-MM-DD)")
	fmt.Println("      --end, -e <date>      End date (YYYY-MM-DD)")
	fmt.Println("      --days, -d <num>      Number of days back from today (default: 30)")
	fmt.Println("      --output, -o <file>   Output file path (defaults to stats_export_<date>.csv)")
	fmt.Println("    - validate             Validate campaign data for optimization")
	fmt.Println("      --start, -s <date>    Start date (YYYY-MM-DD)")
	fmt.Println("      --end, -e <date>      End date (YYYY-MM-DD)")
	fmt.Println("      --days, -d <num>      Number of days back from today (default: 30)")
	fmt.Println("      --campaign, -c <id>   Specific campaign to validate (optional)")
	fmt.Println("      --format, -f <fmt>    Output format: json or table (default: json)")
	fmt.Println("")
	fmt.Println("  audience <subcommand> [args]")
	fmt.Println("                           Audience targeting and analysis commands")
	fmt.Println("    - search <query>           Search for audience segments")
	fmt.Println("      --type, -t <type>        Segment type (default: adinterest)")
	fmt.Println("      --class, -c <class>      Category class when type is adTargetingCategory")
	fmt.Println("      --output, -o <file>      Export results to file")
	fmt.Println("    - filter                   Filter audience segments")
	fmt.Println("      --query, -q <query>      Initial search query")
	fmt.Println("      --min-size <size>        Minimum audience size")
	fmt.Println("      --max-size <size>        Maximum audience size")
	fmt.Println("      --types <types>          Comma-separated list of types")
	fmt.Println("      --keywords, -k <kw>      Comma-separated list of keywords")
	fmt.Println("      --output, -o <file>      Export results to file")
	fmt.Println("    - stats                    Collect segment statistics")
	fmt.Println("      --campaign, -c <id>      Campaign ID to analyze")
	fmt.Println("      --days, -d <days>        Number of days to analyze (default: 30)")
	fmt.Println("")
	fmt.Println("  report <type> [args]     Generate performance reports")
	fmt.Println("    - daily                Daily report for yesterday")
	fmt.Println("    - weekly               Weekly report for the last 7 days")
	fmt.Println("    - custom <start> <end> Custom date range report (YYYY-MM-DD format)")
	fmt.Println("")
	fmt.Println("  optimize <subcommand>    Campaign optimization commands")
	fmt.Println("    - validate <yaml_file>  Validate a YAML campaign configuration file")
	fmt.Println("    - create <yaml_file>    Create test campaigns from a YAML configuration")
	fmt.Println("      --limit <num>         Limit the number of test combinations to create")
	fmt.Println("      --batch-size <num>    Number of campaigns to create in each batch (default: 3)")
	fmt.Println("      --priority <type>     Priority for combinations: audience or placement (default: audience)")
	fmt.Println("      --dry-run, -d         Preview campaigns without creating them")
	fmt.Println("    - update <campaign_ids> Update campaign CPM based on performance data")
	fmt.Println("      --max-cpm <value>     Maximum CPM price allowed (default: 15.0)")
	fmt.Println("")
	fmt.Println("  dashboard [port]         Start web dashboard (default port: 8080)")
	fmt.Println("")
	fmt.Println("  config                   Configure the application")
	fmt.Println("")
	fmt.Println("  help                     Show help information")
}
