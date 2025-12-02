package strategy

type StrategyConfig interface {
	Load(path string) (*Strategy, error)
	FindStrategies() ([]Strategy, error)
	Save(path string, config *Strategy) error
}

// Asset represents a trading asset with its required instruments
type Asset struct {
	Symbol      string   `yaml:"symbol"`
	Instruments []string `yaml:"instruments"`
}

// Strategy represents the parsed config.yml for a strategy
type Strategy struct {
	Name        string                 `yaml:"name"`
	Path        string                 `yaml:"-"`
	Description string                 `yaml:"description"`
	Status      StrategyStatus         `yaml:"-"`
	Error       string                 `yaml:"-"`
	Exchanges   []string               `yaml:"exchanges"`
	Assets      map[string][]Asset     `yaml:"assets"`
	Parameters  map[string]interface{} `yaml:"parameters"`
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
