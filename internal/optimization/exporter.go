package optimization

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/user/fb-ads/pkg/models"
	"gopkg.in/yaml.v3"
)

// ExporterConfig contains configuration for the campaign exporter
type ExporterConfig struct {
	// Initial budget for the test campaign
	TotalBudget float64

	// Percentage of the budget to allocate for testing (0-100)
	TestBudgetPercentage float64

	// Maximum CPM bid allowed
	MaxCPM float64

	// Output file path. If empty, uses stdout
	OutputPath string
}

// DefaultExporterConfig returns a default configuration for campaign export
func DefaultExporterConfig() *ExporterConfig {
	return &ExporterConfig{
		TotalBudget:          1000.0,
		TestBudgetPercentage: 20.0,
		MaxCPM:               15.0,
	}
}

// CampaignExporter handles exporting Facebook campaigns to YAML format
type CampaignExporter struct {
	config *ExporterConfig
}

// NewExporter creates a new campaign exporter
func NewExporter(config *ExporterConfig) *CampaignExporter {
	if config == nil {
		config = DefaultExporterConfig()
	}
	return &CampaignExporter{
		config: config,
	}
}

// ExportCampaign exports a campaign configuration to YAML format
func (e *CampaignExporter) ExportCampaign(campaign *models.CampaignDetails) error {
	// Convert to optimization format
	optConfig := e.convertToOptimizationConfig(campaign)

	// Create YAML
	yamlData, err := yaml.Marshal(optConfig)
	if err != nil {
		return fmt.Errorf("error marshaling to YAML: %w", err)
	}

	// Write to output
	if e.config.OutputPath == "" {
		// Write to stdout
		_, err = fmt.Println(string(yamlData))
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(e.config.OutputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to file
	return os.WriteFile(e.config.OutputPath, yamlData, 0644)
}

// ExportCampaignToWriter exports a campaign configuration to a writer
func (e *CampaignExporter) ExportCampaignToWriter(campaign *models.CampaignDetails, writer io.Writer) error {
	// Convert to optimization format
	optConfig := e.convertToOptimizationConfig(campaign)

	// Create YAML
	yamlData, err := yaml.Marshal(optConfig)
	if err != nil {
		return fmt.Errorf("error marshaling to YAML: %w", err)
	}

	// Write to output
	_, err = writer.Write(yamlData)
	return err
}

// convertToOptimizationConfig converts a CampaignDetails to the optimization format
func (e *CampaignExporter) convertToOptimizationConfig(campaign *models.CampaignDetails) *CampaignOptimizationConfig {
	// Create the base config
	config := &CampaignOptimizationConfig{
		Campaign: CampaignConfig{
			Name:                 campaign.Name,
			TotalBudget:          e.config.TotalBudget,
			TestBudgetPercentage: e.config.TestBudgetPercentage,
			MaxCPM:               e.config.MaxCPM,
		},
		Creatives:      []CreativeConfig{},
		TargetingOptions: TargetingOptions{
			Audiences:  []AudienceConfig{},
			Placements: []PlacementConfig{},
		},
	}

	// Extract creatives from ads
	creativeMap := make(map[string]bool) // Track unique creatives
	for i, ad := range campaign.Ads {
		// Create a unique ID if empty
		id := fmt.Sprintf("creative%d", i+1)

		// Skip if we already processed this creative (based on title/body combination)
		key := ad.Creative.Title + "|" + ad.Creative.Body
		if _, exists := creativeMap[key]; exists {
			continue
		}
		creativeMap[key] = true

		// Add creative
		creative := CreativeConfig{
			ID:           id,
			Title:        ad.Creative.Title,
			Description:  ad.Creative.Body,
			ImageURL:     ad.Creative.ImageURL,
			LinkURL:      ad.Creative.LinkURL,
			CallToAction: ad.Creative.CallToActionType,
			PageID:       ad.Creative.PageID,
		}
		config.Creatives = append(config.Creatives, creative)
	}

	// Extract audiences and placements from ad sets
	audienceMap := make(map[string]bool) // Track unique audiences
	placementMap := make(map[string]bool) // Track unique placements

	for i, adSet := range campaign.AdSets {
		// Process targeting information
		if adSet.Targeting != nil {
			// Extract audience information
			audienceID := fmt.Sprintf("audience%d", i+1)
			audienceName := fmt.Sprintf("Audience %d", i+1)

			// Try to extract a more meaningful name from targeting
			if adSet.Name != "" {
				audienceName = adSet.Name
			}

			// Check if we've already added this audience
			audienceKey := fmt.Sprintf("%v", adSet.Targeting)
			if _, exists := audienceMap[audienceKey]; !exists {
				audienceMap[audienceKey] = true

				audience := AudienceConfig{
					ID:         audienceID,
					Name:       audienceName,
					Parameters: adSet.Targeting,
				}
				config.TargetingOptions.Audiences = append(config.TargetingOptions.Audiences, audience)
			}

			// Extract placement information
			if platforms, ok := adSet.Targeting["publisher_platforms"].([]interface{}); ok {
				for _, platform := range platforms {
					platformStr, ok := platform.(string)
					if !ok {
						continue
					}

					placementID := fmt.Sprintf("placement%d", len(config.TargetingOptions.Placements)+1)
					placementName := fmt.Sprintf("%s", strings.Title(platformStr))
					position := "feed" // Default position

					// Check for specific positions
					if fbPositions, ok := adSet.Targeting["facebook_positions"].([]interface{}); ok && len(fbPositions) > 0 {
						if pos, ok := fbPositions[0].(string); ok {
							position = pos
							placementName = fmt.Sprintf("Facebook %s", strings.Title(pos))
						}
					} else if igPositions, ok := adSet.Targeting["instagram_positions"].([]interface{}); ok && len(igPositions) > 0 {
						if pos, ok := igPositions[0].(string); ok {
							position = pos
							placementName = fmt.Sprintf("Instagram %s", strings.Title(pos))
						}
					}

					// Check if we've already added this placement
					placementKey := platformStr + "|" + position
					if _, exists := placementMap[placementKey]; !exists {
						placementMap[placementKey] = true

						placement := PlacementConfig{
							ID:       placementID,
							Name:     placementName,
							Position: position,
						}
						config.TargetingOptions.Placements = append(config.TargetingOptions.Placements, placement)
					}
				}
			}
		}
	}

	// If no placements were found, add a default one
	if len(config.TargetingOptions.Placements) == 0 {
		config.TargetingOptions.Placements = append(config.TargetingOptions.Placements, PlacementConfig{
			ID:       "placement1",
			Name:     "Facebook Feed",
			Position: "feed",
		})
	}

	// If no audiences were found, add a default one
	if len(config.TargetingOptions.Audiences) == 0 {
		defaultParams := make(map[string]interface{})
		defaultParams["age_min"] = 18
		defaultParams["age_max"] = 65
		defaultParams["genders"] = []int{1, 2}

		config.TargetingOptions.Audiences = append(config.TargetingOptions.Audiences, AudienceConfig{
			ID:         "audience1",
			Name:       "Default Audience",
			Parameters: defaultParams,
		})
	}

	return config
}

// ExportCampaignFromID fetches a campaign by ID and exports it to YAML
func (e *CampaignExporter) ExportCampaignFromID(client interface{}, campaignID string) error {
	// Check if client implements the right interface
	apiClient, ok := client.(CampaignGetter)
	if !ok {
		return fmt.Errorf("client does not implement the CampaignGetter interface")
	}

	// Fetch campaign details
	campaign, err := apiClient.GetCampaignDetails(campaignID)
	if err != nil {
		return fmt.Errorf("failed to get campaign details: %w", err)
	}

	// Export campaign
	return e.ExportCampaign(campaign)
}

// CampaignGetter interface for getting campaign details
type CampaignGetter interface {
	GetCampaignDetails(id string) (*models.CampaignDetails, error)
}

// Helper function to handle budget conversion
func convertBudget(value float64) float64 {
	// Facebook stores budgets in cents, so convert to dollars
	if value > 0 {
		return value / 100.0
	}
	return 0
}

// Helper function to format map keys
func formatKey(key string) string {
	// Capitalize first letter and remove underscores
	parts := strings.Split(key, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[0:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

// Helper function to parse numeric values from interface{}
func parseNumericValue(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}
