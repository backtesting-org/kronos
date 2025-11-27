package types

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Strategy represents a strategy available for live trading
type Strategy struct {
	Name        string
	Path        string
	Description string
	Exchanges   []Exchange
	Status      StrategyStatus
	Config      *StrategyConfig
	Error       string // Error message if status is StatusError
}

// StrategyStatus represents the current status of a strategy
type StrategyStatus string

const (
	StatusReady   StrategyStatus = "ready"
	StatusRunning StrategyStatus = "running"
	StatusStopped StrategyStatus = "stopped"
	StatusError   StrategyStatus = "error"
)

// Connectors represents an exchange configuration
type Exchange struct {
	Name    string
	Enabled bool
	Assets  []string
}

// StrategyConfig represents the parsed config.yml for a strategy
type StrategyConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Status      StrategyStatus         `yaml:"status"`
	Exchanges   []string               `yaml:"exchanges"`  // Exchanges names (references exchanges.yml)
	Assets      map[string][]string    `yaml:"assets"`     // Assets per exchange
	Parameters  map[string]interface{} `yaml:"parameters"` // Strategy-specific parameters
	Risk        RiskConfig             `yaml:"risk"`
	Execution   ExecutionConfig        `yaml:"execution"`
}

// Connectors represents the exchanges.yml file at project root
type Connectors struct {
	Exchanges []ExchangeConfig `yaml:"exchanges"`
}

// ExchangeConfig represents exchange configuration in YAML
type ExchangeConfig struct {
	Name        string            `yaml:"name"`
	Enabled     bool              `yaml:"enabled"`
	Network     string            `yaml:"network,omitempty"` // For Paradex: mainnet/testnet
	Assets      []string          `yaml:"assets"`
	Credentials map[string]string `yaml:"credentials,omitempty"`
}

// RiskConfig represents risk management parameters
type RiskConfig struct {
	MaxPositionSize float64 `yaml:"max_position_size"`
	MaxDailyLoss    float64 `yaml:"max_daily_loss"`
}

// ExecutionConfig represents execution mode settings
type ExecutionConfig struct {
	DryRun bool   `yaml:"dry_run"`
	Mode   string `yaml:"mode"`
}

// DiscoverStrategies scans the ./strategies directory for available strategies
func DiscoverStrategies() ([]Strategy, error) {
	strategiesDir := "./strategies"
	exchangesConfigPath := "./exchanges.yml"

	// Check if strategies directory exists
	if _, err := os.Stat(strategiesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("strategies directory not found: %s", strategiesDir)
	}

	// Load global exchanges config
	globalExchanges, err := LoadGlobalExchangesConfig(exchangesConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load global exchanges config: %w", err)
	}

	// Create a map of exchange configs by name for quick lookup
	exchangeMap := make(map[string]*ExchangeConfig)
	for i := range globalExchanges.Exchanges {
		exchangeMap[globalExchanges.Exchanges[i].Name] = &globalExchanges.Exchanges[i]
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

		// Initialize a strategy with error state by default
		var strategy Strategy
		var config *StrategyConfig

		// Check if config.yml exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Skip directories without config.yml (not a valid strategy)
			continue
		}

		// Load and parse the config
		config, err = LoadStrategyConfig(configPath)
		if err != nil {
			// Config is invalid, but we still want to show the strategy with an error
			strategy = Strategy{
				Name:   strategyName,
				Path:   strategyPath,
				Status: StatusError,
				Error:  fmt.Sprintf("Invalid config: %s", err.Error()),
				Config: &StrategyConfig{
					Name:   strategyName,
					Status: StatusError,
				},
			}
			strategies = append(strategies, strategy)
			continue
		}

		// Check if .so file exists
		soPath := filepath.Join(strategyPath, strategyName+".so")
		errorMsg := ""
		if _, err := os.Stat(soPath); os.IsNotExist(err) {
			// Strategy not built yet, mark as error
			config.Status = StatusError
			errorMsg = fmt.Sprintf("Strategy not compiled - missing %s.so file", strategyName)
		} else {
			// .so exists, mark as ready
			if config.Status == StatusError {
				config.Status = StatusReady
			}
		}

		// Merge strategy config with global exchange configs
		exchanges := make([]Exchange, 0, len(config.Exchanges))
		for _, exchangeName := range config.Exchanges {
			if globalEx, ok := exchangeMap[exchangeName]; ok && globalEx.Enabled {
				// Get assets for this exchange from strategy config
				assets := config.Assets[exchangeName]
				if assets == nil {
					assets = []string{}
				}

				exchanges = append(exchanges, Exchange{
					Name:    globalEx.Name,
					Enabled: globalEx.Enabled,
					Assets:  assets,
				})
			}
		}

		// Convert config to Strategy
		strategy = Strategy{
			Name:        config.Name,
			Path:        strategyPath,
			Description: config.Description,
			Status:      config.Status,
			Exchanges:   exchanges,
			Config:      config,
			Error:       errorMsg,
		}

		strategies = append(strategies, strategy)
	}

	if len(strategies) == 0 {
		return nil, fmt.Errorf("no strategies found in %s (make sure each strategy has a config.yml file)", strategiesDir)
	}

	return strategies, nil
}

// LoadStrategyConfig loads and parses a live.yml config file
func LoadStrategyConfig(path string) (*StrategyConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config StrategyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Status == "" {
		config.Status = "ready"
	}

	// Initialize assets map if nil
	if config.Assets == nil {
		config.Assets = make(map[string][]string)
	}

	// Initialize parameters map if nil
	if config.Parameters == nil {
		config.Parameters = make(map[string]interface{})
	}

	return &config, nil
}

// SaveStrategyConfig saves a strategy config to config.yml
func SaveStrategyConfig(path string, config *StrategyConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadGlobalExchangesConfig loads the global exchanges.yml from project root
func LoadGlobalExchangesConfig(path string) (*Connectors, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Connectors{Exchanges: []ExchangeConfig{}}, nil
		}
		return nil, fmt.Errorf("failed to read exchanges config: %w", err)
	}

	var config Connectors
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse exchanges config: %w", err)
	}

	// Initialize credentials maps
	for i := range config.Exchanges {
		if config.Exchanges[i].Credentials == nil {
			config.Exchanges[i].Credentials = make(map[string]string)
		}
	}

	return &config, nil
}

// SaveGlobalExchangesConfig saves the global exchanges config to exchanges.yml
func SaveGlobalExchangesConfig(path string, config *Exchange) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal exchanges config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write exchanges config: %w", err)
	}

	return nil
}

// parseStatus converts string status to StrategyStatus
func parseStatus(status string) StrategyStatus {
	switch status {
	case "ready":
		return StatusReady
	case "running":
		return StatusRunning
	case "stopped":
		return StatusStopped
	case "error":
		return StatusError
	default:
		return StatusReady
	}
}
