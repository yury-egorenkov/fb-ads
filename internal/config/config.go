package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	APIVersion     string `json:"api_version"`
	AccessToken    string `json:"access_token"`
	AppID          string `json:"app_id"`
	AppSecret      string `json:"app_secret"`
	AccountID      string `json:"account_id"`
	ConfigDir      string `json:"config_dir"`
	OutputFormat   string `json:"output_format"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	
	return &Config{
		APIVersion:   "v18.0",
		ConfigDir:    filepath.Join(homeDir, ".fbads"),
		OutputFormat: "json",
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	
	err = json.Unmarshal(data, cfg)
	return cfg, err
}

// SaveConfig saves configuration to a file
func (c *Config) SaveConfig(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	
	return os.WriteFile(path, data, 0644)
}