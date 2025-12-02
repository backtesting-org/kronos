package strategy

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// strategyConfig implements the StrategyConfig interface
type strategyConfig struct {
}

func NewStrategyConfigService() StrategyConfig {
	return &strategyConfig{}
}

// Load loads and parses a strategy config.yml file
func (c *strategyConfig) Load(path string) (*Strategy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Strategy
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize assets map if nil
	if cfg.Assets == nil {
		cfg.Assets = map[string][]Asset{}
	}

	// Initialize parameters map if nil
	if cfg.Parameters == nil {
		cfg.Parameters = make(map[string]interface{})
	}

	return &cfg, nil
}

// Save saves a strategy config to config.yml
func (c *strategyConfig) Save(path string, config *Strategy) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// FindStrategies scans the ./strategies directory for available strategies
func (c *strategyConfig) FindStrategies() ([]Strategy, error) {
	strategiesDir := "./strategies"

	// Check if strategies directory exists
	if _, err := os.Stat(strategiesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("strategies directory not found: %s", strategiesDir)
	}

	// Read all subdirectories in strategies/
	entries, err := os.ReadDir(strategiesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read strategies directory: %w", err)
	}

	var strategies []Strategy

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		strategyName := entry.Name()
		strategyPath := filepath.Join(strategiesDir, strategyName)
		configPath := filepath.Join(strategyPath, "config.yml")

		// Check if config.yml exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Skip directories without config.yml (not a valid strategy)
			continue
		}

		// Load and parse the config
		cfg, err := c.Load(configPath)
		if err != nil {
			// Config is invalid, skip this strategy
			continue
		}

		// Check if .so file exists
		soPath := filepath.Join(strategyPath, strategyName+".so")
		if _, err := os.Stat(soPath); os.IsNotExist(err) {
			// Strategy not built yet, mark status as error
			cfg.Status = StatusError
		}

		cfg.Path = strategyPath

		strategies = append(strategies, *cfg)
	}

	if len(strategies) == 0 {
		return nil, fmt.Errorf("no strategies found in %s (make sure each strategy has a config.yml file)", strategiesDir)
	}

	return strategies, nil
}
