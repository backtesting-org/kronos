package services

import (
	"github.com/backtesting-org/kronos-cli/internal/config"
	"github.com/backtesting-org/kronos-cli/internal/interactive"
)

// BacktestService handles backtest operations
type BacktestService struct{}

func NewBacktestService() *BacktestService {
	return &BacktestService{}
}

func (s *BacktestService) RunInteractive() error {
	cfg, err := interactive.InteractiveMode()
	if err != nil {
		return err
	}
	return s.ExecuteBacktest(cfg)
}

func (s *BacktestService) ExecuteBacktest(cfg *config.Config) error {
	// TODO: Implement actual backtest execution
	return nil
}

func (s *BacktestService) LoadConfig(path string) (*config.Config, error) {
	return config.LoadConfig(path)
}
