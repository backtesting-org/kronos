package settings

type Configuration interface {
	LoadSettings() (*Settings, error)
	GetConnectors() ([]Connector, error)
	GetEnabledConnectors() ([]Connector, error)
}

// Settings represents the main settings structure
type Settings struct {
	Version    string         `mapstructure:"version"`
	Backtest   BacktestConfig `mapstructure:"backtest"`
	Live       LiveConfig     `mapstructure:"live"`
	Connectors []Connector    `mapstructure:"connectors"`
}

type Connector struct {
	Name        string            `yaml:"name"`
	Enabled     bool              `yaml:"enabled"`
	Network     string            `yaml:"network,omitempty"`
	Assets      []string          `yaml:"assets"`
	Credentials map[string]string `yaml:"credentials,omitempty"`
}

// BacktestConfig holds backtest settings
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

// LiveConfig holds live trading settings
type LiveConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Exchange  string `mapstructure:"exchange"`
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
}

// Validate validates the settings
//func (c *Config) Validate() error {
//	if c.Backtest.Strategy == "" {
//		return fmt.Errorf("strategy is required")
//	}
//
//	if c.Backtest.Exchange == "" {
//		return fmt.Errorf("exchange is required")
//	}
//
//	if c.Backtest.Pair == "" {
//		return fmt.Errorf("pair is required")
//	}
//
//	// Validate timeframe dates
//	if c.Backtest.Timeframe.Start != "" {
//		if _, err := time.Parse("2006-01-02", c.Backtest.Timeframe.Start); err != nil {
//			return fmt.Errorf("invalid start date format (use YYYY-MM-DD): %w", err)
//		}
//	}
//
//	if c.Backtest.Timeframe.End != "" {
//		if _, err := time.Parse("2006-01-02", c.Backtest.Timeframe.End); err != nil {
//			return fmt.Errorf("invalid end date format (use YYYY-MM-DD): %w", err)
//		}
//	}
//
//	return nil
//}
