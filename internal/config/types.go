package config

import (
	"fmt"
	"time"
)

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Backtest.Strategy == "" {
		return fmt.Errorf("strategy is required")
	}

	if c.Backtest.Exchange == "" {
		return fmt.Errorf("exchange is required")
	}

	if c.Backtest.Pair == "" {
		return fmt.Errorf("pair is required")
	}

	// Validate timeframe dates
	if c.Backtest.Timeframe.Start != "" {
		if _, err := time.Parse("2006-01-02", c.Backtest.Timeframe.Start); err != nil {
			return fmt.Errorf("invalid start date format (use YYYY-MM-DD): %w", err)
		}
	}

	if c.Backtest.Timeframe.End != "" {
		if _, err := time.Parse("2006-01-02", c.Backtest.Timeframe.End); err != nil {
			return fmt.Errorf("invalid end date format (use YYYY-MM-DD): %w", err)
		}
	}

	return nil
}

// Config represents the kronos.yml configuration
type Config struct {
	Version   string           `mapstructure:"version"`
	Backtest  BacktestConfig   `mapstructure:"backtest"`
	Live      LiveConfig       `mapstructure:"live"`
	Exchanges []ExchangeConfig `mapstructure:"exchanges"`
}

// ExchangeCredentials wraps the exchange configurations
type ExchangeCredentials struct {
	Exchanges []ExchangeConfig
}

// ExchangeConfig represents exchange configuration in YAML
type ExchangeConfig struct {
	Name        string            `mapstructure:"name" yaml:"name"`
	Enabled     bool              `mapstructure:"enabled" yaml:"enabled"`
	Network     string            `mapstructure:"network,omitempty" yaml:"network,omitempty"`
	Assets      []string          `mapstructure:"assets" yaml:"assets"`
	Credentials map[string]string `mapstructure:"credentials,omitempty" yaml:"credentials,omitempty"`
}

// BacktestConfig holds backtest configuration
type BacktestConfig struct {
	Strategy   string                 `mapstructure:"strategy"`
	Exchange   string                 `mapstructure:"exchange"`
	Pair       string                 `mapstructure:"pair"`
	Timeframe  TimeframeConfig        `mapstructure:"timeframe"`
	Parameters map[string]interface{} `mapstructure:"parameters"`
	Execution  ExecutionConfig        `mapstructure:"execution"`
	Output     OutputConfig           `mapstructure:"output"`
}

// TimeframeConfig defines the backtest time period
type TimeframeConfig struct {
	Start string `mapstructure:"start"`
	End   string `mapstructure:"end"`
}

// ExecutionConfig defines execution parameters
type ExecutionConfig struct {
}

// StrategyConfig represents the parsed config.yml for a strategy
type StrategyConfig struct {
	Name        string                 `yaml:"name" mapstructure:"name"`
	Description string                 `yaml:"description" mapstructure:"description"`
	Exchanges   []string               `yaml:"exchanges" mapstructure:"exchanges"`
	Assets      map[string][]string    `yaml:"assets" mapstructure:"assets"`
	Parameters  map[string]interface{} `yaml:"parameters" mapstructure:"parameters"`
}

// OutputConfig defines output settings
type OutputConfig struct {
	Format      string `mapstructure:"format"`
	SaveResults bool   `mapstructure:"save_results"`
	ResultsDir  string `mapstructure:"results_dir"`
}

// LiveConfig holds live trading configuration
type LiveConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Exchange  string `mapstructure:"exchange"`
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
}
