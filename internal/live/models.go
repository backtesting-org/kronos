package live

import (
	"fmt"
	"os"
	"os/exec"
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
}

// StrategyStatus represents the current status of a strategy
type StrategyStatus string

const (
	StatusReady   StrategyStatus = "ready"
	StatusRunning StrategyStatus = "running"
	StatusStopped StrategyStatus = "stopped"
	StatusError   StrategyStatus = "error"
)

// Exchange represents an exchange configuration
type Exchange struct {
	Name    string
	Enabled bool
	Assets  []string
}

// StrategyConfig represents the parsed live.yml config
type StrategyConfig struct {
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Status      string                  `yaml:"status"`
	Exchanges   []ExchangeConfig        `yaml:"exchanges"`
	Risk        RiskConfig              `yaml:"risk"`
	Execution   ExecutionConfig         `yaml:"execution"`
}

// ParadexCredentials represents Paradex-specific credentials
type ParadexCredentials struct {
	AccountAddress string `yaml:"account_address"`
	EthPrivateKey  string `yaml:"eth_private_key"`
	L2PrivateKey   string `yaml:"l2_private_key,omitempty"`
}

// ExchangeConfig represents exchange configuration in YAML
type ExchangeConfig struct {
	Name        string            `yaml:"name"`
	Enabled     bool              `yaml:"enabled"`
	Network     string            `yaml:"network,omitempty"`      // For Paradex: mainnet/testnet
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

		strategyPath := filepath.Join(strategiesDir, entry.Name())
		configPath := filepath.Join(strategyPath, "live.yml")

		// Check if live.yml exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Skip directories without live.yml
			continue
		}

		// Load and parse the config
		config, err := LoadStrategyConfig(configPath)
		if err != nil {
			// Skip strategies with invalid config
			continue
		}

		// Check if .so file exists
		soPath := filepath.Join(strategyPath, entry.Name()+".so")
		if _, err := os.Stat(soPath); os.IsNotExist(err) {
			// Strategy not built yet
			config.Status = "error"
		}

		// Convert ExchangeConfig to Exchange
		exchanges := make([]Exchange, 0, len(config.Exchanges))
		for _, exConfig := range config.Exchanges {
			if exConfig.Enabled {
				exchanges = append(exchanges, Exchange{
					Name:    exConfig.Name,
					Enabled: exConfig.Enabled,
					Assets:  exConfig.Assets,
				})
			}
		}

		// Convert config to Strategy
		strategy := Strategy{
			Name:        config.Name,
			Path:        strategyPath,
			Description: config.Description,
			Status:      parseStatus(config.Status),
			Exchanges:   exchanges,
			Config:      config,
		}

		strategies = append(strategies, strategy)
	}

	if len(strategies) == 0 {
		return nil, fmt.Errorf("no strategies found in %s (make sure each strategy has a live.yml file)", strategiesDir)
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

	// Initialize credentials map for each exchange
	for i := range config.Exchanges {
		if config.Exchanges[i].Credentials == nil {
			config.Exchanges[i].Credentials = make(map[string]string)
		}
		// Set default network for Paradex
		if config.Exchanges[i].Name == "paradex" && config.Exchanges[i].Network == "" {
			config.Exchanges[i].Network = "mainnet"
		}
	}

	return &config, nil
}

// SaveStrategyConfig saves a strategy config to live.yml
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

// GetMockStrategies returns mock strategies for demo purposes
func GetMockStrategies() []Strategy {
	return []Strategy{
		{
			Name:        "momentum",
			Path:        "./strategies/momentum",
			Description: "Trend-following momentum strategy",
			Status:      StatusReady,
			Exchanges: []Exchange{
				{Name: "paradex", Enabled: true, Assets: []string{"BTC-USD-PERP", "ETH-USD-PERP"}},
			},
			Config: &StrategyConfig{
				Name:        "momentum",
				Description: "Trend-following momentum strategy",
				Status:      "ready",
				Exchanges: []ExchangeConfig{
					{
						Name:        "paradex",
						Enabled:     true,
						Network:     "mainnet",
						Assets:      []string{"BTC-USD-PERP", "ETH-USD-PERP"},
						Credentials: make(map[string]string),
					},
				},
				Risk: RiskConfig{
					MaxPositionSize: 10000,
					MaxDailyLoss:    500,
				},
				Execution: ExecutionConfig{
					DryRun: true,
					Mode:   "live",
				},
			},
		},
		{
			Name:        "arbitrage",
			Path:        "./strategies/arbitrage",
			Description: "Cross-exchange arbitrage",
			Status:      StatusReady,
			Exchanges: []Exchange{
				{Name: "binance", Enabled: true, Assets: []string{"BTCUSDT", "ETHUSDT"}},
				{Name: "kraken", Enabled: true, Assets: []string{"BTCUSD", "ETHUSD"}},
			},
			Config: &StrategyConfig{
				Name:        "arbitrage",
				Description: "Cross-exchange arbitrage",
				Status:      "ready",
				Exchanges: []ExchangeConfig{
					{
						Name:        "binance",
						Enabled:     true,
						Assets:      []string{"BTCUSDT", "ETHUSDT"},
						Credentials: make(map[string]string),
					},
					{
						Name:        "kraken",
						Enabled:     true,
						Assets:      []string{"BTCUSD", "ETHUSD"},
						Credentials: make(map[string]string),
					},
				},
				Risk: RiskConfig{
					MaxPositionSize: 5000,
					MaxDailyLoss:    250,
				},
				Execution: ExecutionConfig{
					DryRun: true,
					Mode:   "live",
				},
			},
		},
		{
			Name:        "market_making",
			Path:        "./strategies/market_making",
			Description: "Automated market making strategy",
			Status:      StatusReady,
			Exchanges: []Exchange{
				{Name: "paradex", Enabled: true, Assets: []string{"SOL-USD-PERP"}},
			},
			Config: &StrategyConfig{
				Name:        "market_making",
				Description: "Automated market making strategy",
				Status:      "ready",
				Exchanges: []ExchangeConfig{
					{
						Name:        "paradex",
						Enabled:     true,
						Network:     "mainnet",
						Assets:      []string{"SOL-USD-PERP"},
						Credentials: make(map[string]string),
					},
				},
				Risk: RiskConfig{
					MaxPositionSize: 20000,
					MaxDailyLoss:    1000,
				},
				Execution: ExecutionConfig{
					DryRun: false,
					Mode:   "live",
				},
			},
		},
	}
}

// ExecuteLiveTrading builds and executes the live-trading CLI command
func ExecuteLiveTrading(strategy *Strategy, exchange *ExchangeConfig) error {
	if strategy == nil || exchange == nil {
		return fmt.Errorf("strategy and exchange must be provided")
	}

	// Get the .so file path
	strategyName := filepath.Base(strategy.Path)
	soPath, err := filepath.Abs(filepath.Join(strategy.Path, strategyName+".so"))
	if err != nil {
		return fmt.Errorf("failed to get absolute path for strategy: %w", err)
	}

	// Check if .so file exists
	if _, err := os.Stat(soPath); os.IsNotExist(err) {
		return fmt.Errorf("strategy plugin not found: %s (did you build the strategy?)", soPath)
	}

	// Path to kronos-live binary
	// TODO: Make this configurable or discover it automatically
	kronosLivePath := "/Users/williamr/Documents/holdex/repos/live-trading/bin/kronos-live"

	// Check if kronos-live exists
	if _, err := os.Stat(kronosLivePath); os.IsNotExist(err) {
		return fmt.Errorf("kronos-live binary not found: %s", kronosLivePath)
	}

	// Build command arguments
	args := []string{"run", "--exchange", exchange.Name, "--strategy", soPath}

	// Add exchange-specific flags
	switch exchange.Name {
	case "paradex":
		if accountAddr, ok := exchange.Credentials["account_address"]; ok && accountAddr != "" {
			args = append(args, "--paradex-account-address", accountAddr)
		} else {
			return fmt.Errorf("paradex account address is required")
		}

		if ethKey, ok := exchange.Credentials["eth_private_key"]; ok && ethKey != "" {
			args = append(args, "--paradex-eth-private-key", ethKey)
		} else {
			return fmt.Errorf("paradex eth private key is required")
		}

		if l2Key, ok := exchange.Credentials["l2_private_key"]; ok && l2Key != "" {
			args = append(args, "--paradex-l2-private-key", l2Key)
		}

		if exchange.Network != "" {
			args = append(args, "--paradex-network", exchange.Network)
		}

	case "bybit", "binance", "kraken":
		if apiKey, ok := exchange.Credentials["api_key"]; ok && apiKey != "" {
			args = append(args, "--api-key", apiKey)
		} else {
			return fmt.Errorf("%s api key is required", exchange.Name)
		}

		if apiSecret, ok := exchange.Credentials["api_secret"]; ok && apiSecret != "" {
			args = append(args, "--api-secret", apiSecret)
		} else {
			return fmt.Errorf("%s api secret is required", exchange.Name)
		}

	default:
		return fmt.Errorf("unsupported exchange: %s", exchange.Name)
	}

	// Create and execute the command
	cmd := exec.Command(kronosLivePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run the command (this will block until the user stops it with Ctrl+C)
	return cmd.Run()
}
