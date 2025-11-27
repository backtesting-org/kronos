package strategy

type StrategyConfig interface {
	Load(path string) (*Strategy, error)
	FindStrategies() ([]Strategy, error)
	Save(path string, config *Strategy) error
}

// StrategyExecutionConfig represents execution mode settings
type StrategyExecutionConfig struct {
	DryRun bool   `yaml:"dry_run"`
	Mode   string `yaml:"mode"`
}

// Strategy represents the parsed config.yml for a strategy
type Strategy struct {
	Name        string                  `yaml:"name"`
	Path        string                  `yaml:"-"`
	Description string                  `yaml:"description"`
	Status      StrategyStatus          `yaml:"-"`
	Error       string                  `yaml:"-"`
	Exchanges   []string                `yaml:"exchanges"`
	Assets      map[string][]string     `yaml:"assets"`
	Parameters  map[string]interface{}  `yaml:"parameters"`
	Risk        RiskConfig              `yaml:"risk"`
	Execution   StrategyExecutionConfig `yaml:"execution"`
}

// RiskConfig represents risk management parameters
type RiskConfig struct {
	MaxPositionSize float64 `yaml:"max_position_size"`
	MaxDailyLoss    float64 `yaml:"max_daily_loss"`
}

type StrategyStatus string

const (
	StatusReady   StrategyStatus = "ready"
	StatusRunning StrategyStatus = "running"
	StatusStopped StrategyStatus = "stopped"
	StatusError   StrategyStatus = "error"
)

type Exchange struct {
	Name    string
	Enabled bool
	Assets  []string
}
