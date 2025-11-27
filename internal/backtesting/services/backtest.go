package services

import (
	"github.com/backtesting-org/kronos-cli/internal/backtesting/interactive"
	"github.com/backtesting-org/kronos-cli/internal/backtesting/types"
	"github.com/backtesting-org/kronos-cli/internal/config"
)

// backtestService handles backtest operations
type backtestService struct{}

func NewBacktestService() types.BacktestService {
	return &backtestService{}
}

func (s *backtestService) RunInteractive() error {
	cfg, err := interactive.InteractiveMode()
	if err != nil {
		return err
	}
	return s.ExecuteBacktest(cfg)
}

func (s *backtestService) ExecuteBacktest(cfg *config.Config) error {
	// TODO: Implement actual backtest execution
	return nil
}

func (s *backtestService) LoadConfig(path string) (*config.Config, error) {
	return config.LoadConfig(path)
}
