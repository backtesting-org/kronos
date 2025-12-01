package services

import (
	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/strategies/backtest/types"
)

// backtestService handles backtest operations
type backtestService struct{}

func NewBacktestService() types.BacktestService {
	return &backtestService{}
}

func (s *backtestService) RunInteractive() error {
	//cfg, err := interactive.InteractiveMode()
	//if err != nil {
	//	return err
	//}
	//return s.ExecuteBacktest(cfg)

	return nil
}

func (s *backtestService) ExecuteBacktest(cfg *settings.Settings) error {
	// TODO: Implement actual backtest execution
	return nil
}
