package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config represents the kronos.yml configuration
type Config struct {
	Version  string         `mapstructure:"version"`
	Backtest BacktestConfig `mapstructure:"backtest"`
	Live     LiveConfig     `mapstructure:"live"`
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

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
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

	return &cfg, nil
}

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
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
