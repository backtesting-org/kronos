package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// StrategyMetadata represents metadata from a strategy config file
type StrategyMetadata struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"display_name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Icon        string `yaml:"icon,omitempty"`
	SDKExample  string `yaml:"-"` // Directory name in examples/
}

// ToTemplate converts StrategyMetadata to a generic template format
func (s *StrategyMetadata) ToTemplate() map[string]interface{} {
	return map[string]interface{}{
		"Name":        s.Name,
		"DisplayName": s.DisplayName,
		"Description": s.Description,
		"Type":        s.Type,
		"Icon":        s.Icon,
		"SDKExample":  s.SDKExample,
	}
}

// FetchAvailableStrategies fetches strategy metadata from SDK examples
func FetchAvailableStrategies() ([]StrategyMetadata, error) {
	// Try local SDK directory first (for development/testing)
	localSDKPath := "/Users/williamr/Documents/holdex/repos/kronos-sdk/examples"
	strategies, err := fetchFromLocal(localSDKPath)
	if err == nil && len(strategies) > 0 {
		return strategies, nil
	}

	// Fallback to GitHub if local not available
	return fetchFromGitHub()
}

// fetchFromLocal reads strategy metadata from local SDK directory
func fetchFromLocal(basePath string) ([]StrategyMetadata, error) {
	// Check if local SDK directory exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("local SDK path not found: %s", basePath)
	}

	// List of example directories to check
	exampleDirs := []string{
		"mean_reversion",
		"arbitrage",
		"grid_trading",
		"momentum",
	}

	var strategies []StrategyMetadata

	for _, dir := range exampleDirs {
		configPath := fmt.Sprintf("%s/%s/config.yml", basePath, dir)

		// Read config.yml from local filesystem
		body, err := os.ReadFile(configPath)
		if err != nil {
			// Skip this example - no config.yml
			continue
		}

		var metadata StrategyMetadata
		if err := yaml.Unmarshal(body, &metadata); err != nil {
			// Skip this example - invalid YAML
			continue
		}

		// Validate required fields
		if metadata.Name == "" || metadata.DisplayName == "" || metadata.Description == "" || metadata.Type == "" {
			// Skip this example - missing required metadata
			continue
		}

		metadata.SDKExample = dir
		strategies = append(strategies, metadata)
	}

	if len(strategies) == 0 {
		return nil, fmt.Errorf("no valid strategies found in local SDK")
	}

	return strategies, nil
}

// fetchFromGitHub fetches strategy metadata from GitHub repository
func fetchFromGitHub() ([]StrategyMetadata, error) {
	// GitHub raw content base URL
	baseURL := "https://raw.githubusercontent.com/backtesting-org/kronos-sdk/main/examples"

	// List of example directories to check
	exampleDirs := []string{
		"mean_reversion",
		"arbitrage",
		"grid_trading",
		"momentum",
	}

	var strategies []StrategyMetadata

	for _, dir := range exampleDirs {
		// Only check config.yml (skip if not present)
		configURL := fmt.Sprintf("%s/%s/config.yml", baseURL, dir)

		resp, err := http.Get(configURL)
		if err != nil {
			// Skip this example - no config.yml
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// Skip this example - config.yml not found (404, etc)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			// Skip this example - couldn't read config.yml
			continue
		}

		var metadata StrategyMetadata
		if err := yaml.Unmarshal(body, &metadata); err != nil {
			// Skip this example - invalid YAML
			continue
		}

		// Validate required fields
		if metadata.Name == "" || metadata.DisplayName == "" || metadata.Description == "" || metadata.Type == "" {
			// Skip this example - missing required metadata
			continue
		}

		metadata.SDKExample = dir
		strategies = append(strategies, metadata)
	}

	if len(strategies) == 0 {
		return nil, fmt.Errorf("no valid strategies found in SDK")
	}

	return strategies, nil
}

// GetDefaultIcon returns a default icon for a strategy type
func GetDefaultIcon(strategyType string) string {
	iconMap := map[string]string{
		"momentum":       "📈",
		"mean_reversion": "📉",
		"arbitrage":      "💱",
		"grid":           "⚡",
		"grid_trading":   "⚡",
		"technical":      "📊",
		"volume":         "📦",
	}

	if icon, ok := iconMap[strings.ToLower(strategyType)]; ok {
		return icon
	}
	return "🎯"
}
