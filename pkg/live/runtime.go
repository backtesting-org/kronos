package live

// Runtime is the interface for the live trading startup
type Runtime interface {
	Run(strategyDir string) error
}
