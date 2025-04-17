package optimization

import (
	"strings"
	"testing"
)

func TestParseYAMLReader(t *testing.T) {
	validYAML := `
campaign:
  name: "Test Campaign Series Q1"
  total_budget: 1000.00
  test_budget_percentage: 20
  max_cpm: 15.00

creatives:
  - id: "creative1"
    title: "Summer Sale"
    description: "Get 50% off"
    image_url: "https://example.com/image1.jpg"
  - id: "creative2"
    title: "New Arrivals"
    description: "Check out our latest products"
    image_url: "https://example.com/image2.jpg"

targeting_options:
  audiences:
    - id: "audience1"
      name: "18-24 Male"
      parameters:
        age_min: 18
        age_max: 24
        genders: [1]
    - id: "audience2"
      name: "25-34 Female"
      parameters:
        age_min: 25
        age_max: 34
        genders: [2]
  placements:
    - id: "placement1"
      name: "Facebook Feed"
      position: "feed"
    - id: "placement2"
      name: "Instagram Stories"
      position: "story"
`

	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid YAML",
			yaml:    validYAML,
			wantErr: false,
		},
		{
			name: "Missing campaign name",
			yaml: strings.Replace(validYAML, `name: "Test Campaign Series Q1"`, ``, 1),
			wantErr: true,
			errMsg:  "campaign name is required",
		},
		{
			name: "Invalid total budget",
			yaml: strings.Replace(validYAML, `total_budget: 1000.00`, `total_budget: 0`, 1),
			wantErr: true,
			errMsg:  "total budget must be greater than 0",
		},
		{
			name: "Invalid test budget percentage",
			yaml: strings.Replace(validYAML, `test_budget_percentage: 20`, `test_budget_percentage: 0`, 1),
			wantErr: true,
			errMsg:  "test budget percentage must be between 0 and 100",
		},
		{
			name: "Invalid max CPM",
			yaml: strings.Replace(validYAML, `max_cpm: 15.00`, `max_cpm: -5`, 1),
			wantErr: true,
			errMsg:  "max CPM must be greater than 0",
		},
		{
			name: "No creatives",
			yaml: strings.Replace(validYAML, `creatives:
  - id: "creative1"
    title: "Summer Sale"
    description: "Get 50% off"
    image_url: "https://example.com/image1.jpg"
  - id: "creative2"
    title: "New Arrivals"
    description: "Check out our latest products"
    image_url: "https://example.com/image2.jpg"`, `creatives: []`, 1),
			wantErr: true,
			errMsg:  "at least one creative is required",
		},
		{
			name: "No audiences",
			yaml: strings.Replace(validYAML, `audiences:
    - id: "audience1"
      name: "18-24 Male"
      parameters:
        age_min: 18
        age_max: 24
        genders: [1]
    - id: "audience2"
      name: "25-34 Female"
      parameters:
        age_min: 25
        age_max: 34
        genders: [2]`, `audiences: []`, 1),
			wantErr: true,
			errMsg:  "at least one audience is required",
		},
		{
			name: "Missing creative ID",
			yaml: strings.Replace(validYAML, `id: "creative1"`, ``, 1),
			wantErr: true,
			errMsg:  "missing ID",
		},
		{
			name: "Missing audience ID",
			yaml: strings.Replace(validYAML, `id: "audience1"`, ``, 1),
			wantErr: true,
			errMsg:  "missing ID",
		},
		{
			name: "Missing placement position",
			yaml: strings.Replace(validYAML, `position: "feed"`, ``, 1),
			wantErr: true,
			errMsg:  "missing position",
		},
		{
			name: "Duplicate creative ID",
			yaml: strings.Replace(validYAML, `id: "creative2"`, `id: "creative1"`, 1),
			wantErr: true,
			errMsg:  "duplicate creative ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.yaml)
			config, err := ParseYAMLReader(reader)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseYAMLReader() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseYAMLReader() error = %v, should contain %v", err, tt.errMsg)
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseYAMLReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// Basic validation for successful parsing
			if config.Campaign.Name != "Test Campaign Series Q1" {
				t.Errorf("Expected campaign name 'Test Campaign Series Q1', got %s", config.Campaign.Name)
			}
			
			if len(config.Creatives) != 2 {
				t.Errorf("Expected 2 creatives, got %d", len(config.Creatives))
			}
			
			if len(config.TargetingOptions.Audiences) != 2 {
				t.Errorf("Expected 2 audiences, got %d", len(config.TargetingOptions.Audiences))
			}
			
			if len(config.TargetingOptions.Placements) != 2 {
				t.Errorf("Expected 2 placements, got %d", len(config.TargetingOptions.Placements))
			}
		})
	}
}