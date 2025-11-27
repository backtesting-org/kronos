package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type Configuration interface {
	LoadKronosConfig() (*Config, error)
	LoadStrategyConfig(path string) (*StrategyConfig, error)
	GetExchangeCredentials() (*ExchangeCredentials, error)
}

type configuration struct {
	kronosConfigPath string
	config           *Config
	credentials      *ExchangeCredentials
	mu               sync.RWMutex
}

func NewConfiguration() Configuration {
	return &configuration{
		kronosConfigPath: "kronos.yml",
	}
}

// LoadKronosConfig loads the Kronos configuration from the specified file path
func (c *configuration) LoadKronosConfig() (*Config, error) {
	if c.config != nil {
		return c.config, nil
	}
	if !c.fileExists(c.kronosConfigPath) {
		return nil, fmt.Errorf("kronos instance not found, please run 'kronos init' to create one")
	}

	v := viper.New()
	v.SetConfigFile(c.kronosConfigPath)
	v.SetConfigType("yaml")

	// Enable environment variable substitution
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Cache the config
	c.config = &cfg

	// Cache exchange credentials
	c.credentials = &ExchangeCredentials{
		Exchanges: cfg.Exchanges,
	}

	return c.config, nil
}

func (c *configuration) LoadStrategyConfig(path string) (*StrategyConfig, error) {
	if !c.fileExists(path) {
		return nil, fmt.Errorf("strategy config file does not exist: %s", path)
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Enable environment variable substitution
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read strategy config file: %w", err)
	}

	var cfg StrategyConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal strategy config: %w", err)
	}

	return &cfg, nil
}

// GetExchangeCredentials returns the cached exchange credentials from kronos.yml
// If not loaded yet, it will load the kronos config first
func (c *configuration) GetExchangeCredentials() (*ExchangeCredentials, error) {
	if c.credentials != nil {
		return c.credentials, nil
	}

	// Load the full config which will also cache credentials
	if _, err := c.LoadKronosConfig(); err != nil {
		return nil, err
	}

	return c.credentials, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Backtest: BacktestConfig{
			Strategy: "market_making",
			Exchange: "binance",
			Pair:     "BTC/USDT",
			Timeframe: TimeframeConfig{
				Start: "2024-01-01",
				End:   "2024-06-30",
			},
			Parameters: map[string]interface{}{
				"bid_spread":      0.1,
				"ask_spread":      0.1,
				"order_size":      1.0,
				"inventory_limit": 5.0,
			},
			Execution: ExecutionConfig{},
			Output: OutputConfig{
				Format:      "text",
				SaveResults: true,
				ResultsDir:  "./results",
			},
		},
		Live: LiveConfig{
			Enabled: false,
		},
	}
}

// FileExists checks if the config file exists
func (c *configuration) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
