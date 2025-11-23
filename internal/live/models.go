package live

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
	Version   string           `yaml:"version"`
	Strategy  StrategyInfo     `yaml:"strategy"`
	Exchanges []ExchangeConfig `yaml:"exchanges"`
	Risk      RiskConfig       `yaml:"risk"`
	Execution ExecutionConfig  `yaml:"execution"`
}

// StrategyInfo contains basic strategy metadata
type StrategyInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// ExchangeConfig represents exchange configuration in YAML
type ExchangeConfig struct {
	Name    string   `yaml:"name"`
	Enabled bool     `yaml:"enabled"`
	Assets  []string `yaml:"assets"`
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

// GetMockStrategies returns mock strategies for demo purposes
func GetMockStrategies() []Strategy {
	return []Strategy{
		{
			Name:        "momentum",
			Path:        "./strategies/momentum",
			Description: "Trend-following momentum strategy",
			Status:      StatusReady,
			Exchanges: []Exchange{
				{Name: "paradex", Enabled: true, Assets: []string{"BTC-USD", "ETH-USD"}},
			},
			Config: &StrategyConfig{
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
				{Name: "paradex", Enabled: true, Assets: []string{"SOL-USD"}},
			},
			Config: &StrategyConfig{
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
