package optimization

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// CampaignOptimizationConfig represents the top-level YAML configuration for campaign optimization
type CampaignOptimizationConfig struct {
	Campaign        CampaignConfig       `yaml:"campaign"`
	Creatives       []CreativeConfig     `yaml:"creatives"`
	TargetingOptions TargetingOptions     `yaml:"targeting_options"`
}

// CampaignConfig represents the campaign configuration section
type CampaignConfig struct {
	Name                 string  `yaml:"name"`
	TotalBudget          float64 `yaml:"total_budget"`
	TestBudgetPercentage float64 `yaml:"test_budget_percentage"`
	MaxCPM               float64 `yaml:"max_cpm"`
}

// CreativeConfig represents an ad creative configuration
type CreativeConfig struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	ImageURL    string `yaml:"image_url"`
	LinkURL     string `yaml:"link_url,omitempty"`
	CallToAction string `yaml:"call_to_action,omitempty"`
	PageID      string `yaml:"page_id,omitempty"`
}

// TargetingOptions represents all available targeting options for testing
type TargetingOptions struct {
	Audiences  []AudienceConfig  `yaml:"audiences"`
	Placements []PlacementConfig `yaml:"placements"`
}

// AudienceConfig represents an audience targeting configuration
type AudienceConfig struct {
	ID         string                 `yaml:"id"`
	Name       string                 `yaml:"name"`
	Parameters map[string]interface{} `yaml:"parameters"`
}

// PlacementConfig represents an ad placement configuration
type PlacementConfig struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Position string `yaml:"position"`
}

// ParseYAMLConfig parses a YAML file into a CampaignOptimizationConfig
func ParseYAMLConfig(filePath string) (*CampaignOptimizationConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening YAML config file: %w", err)
	}
	defer file.Close()

	return ParseYAMLReader(file)
}

// ParseYAMLReader parses YAML from an io.Reader into a CampaignOptimizationConfig
func ParseYAMLReader(reader io.Reader) (*CampaignOptimizationConfig, error) {
	config := &CampaignOptimizationConfig{}
	
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("error decoding YAML: %w", err)
	}
	
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	
	return config, nil
}

// validateConfig checks if the configuration is valid
func validateConfig(config *CampaignOptimizationConfig) error {
	// Validate campaign section
	if config.Campaign.Name == "" {
		return fmt.Errorf("campaign name is required")
	}
	
	if config.Campaign.TotalBudget <= 0 {
		return fmt.Errorf("total budget must be greater than 0")
	}
	
	if config.Campaign.TestBudgetPercentage <= 0 || config.Campaign.TestBudgetPercentage > 100 {
		return fmt.Errorf("test budget percentage must be between 0 and 100")
	}
	
	if config.Campaign.MaxCPM <= 0 {
		return fmt.Errorf("max CPM must be greater than 0")
	}
	
	// Validate creatives
	if len(config.Creatives) == 0 {
		return fmt.Errorf("at least one creative is required")
	}
	
	creativeIDs := make(map[string]bool)
	for i, creative := range config.Creatives {
		if creative.ID == "" {
			return fmt.Errorf("creative #%d missing ID", i+1)
		}
		
		if creative.Title == "" {
			return fmt.Errorf("creative #%d (%s) missing title", i+1, creative.ID)
		}
		
		if creative.ImageURL == "" {
			return fmt.Errorf("creative #%d (%s) missing image URL", i+1, creative.ID)
		}
		
		if _, exists := creativeIDs[creative.ID]; exists {
			return fmt.Errorf("duplicate creative ID: %s", creative.ID)
		}
		creativeIDs[creative.ID] = true
	}
	
	// Validate targeting options
	if len(config.TargetingOptions.Audiences) == 0 {
		return fmt.Errorf("at least one audience is required")
	}
	
	audienceIDs := make(map[string]bool)
	for i, audience := range config.TargetingOptions.Audiences {
		if audience.ID == "" {
			return fmt.Errorf("audience #%d missing ID", i+1)
		}
		
		if audience.Name == "" {
			return fmt.Errorf("audience #%d (%s) missing name", i+1, audience.ID)
		}
		
		if len(audience.Parameters) == 0 {
			return fmt.Errorf("audience #%d (%s) has no targeting parameters", i+1, audience.ID)
		}
		
		if _, exists := audienceIDs[audience.ID]; exists {
			return fmt.Errorf("duplicate audience ID: %s", audience.ID)
		}
		audienceIDs[audience.ID] = true
	}
	
	// At least one placement is required if placements section exists
	if len(config.TargetingOptions.Placements) > 0 {
		placementIDs := make(map[string]bool)
		for i, placement := range config.TargetingOptions.Placements {
			if placement.ID == "" {
				return fmt.Errorf("placement #%d missing ID", i+1)
			}
			
			if placement.Name == "" {
				return fmt.Errorf("placement #%d (%s) missing name", i+1, placement.ID)
			}
			
			if placement.Position == "" {
				return fmt.Errorf("placement #%d (%s) missing position", i+1, placement.ID)
			}
			
			if _, exists := placementIDs[placement.ID]; exists {
				return fmt.Errorf("duplicate placement ID: %s", placement.ID)
			}
			placementIDs[placement.ID] = true
		}
	}
	
	return nil
}