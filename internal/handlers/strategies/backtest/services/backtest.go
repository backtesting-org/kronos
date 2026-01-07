package services

import (
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/backtest/types"
	"github.com/backtesting-org/kronos-sdk/pkg/types/config"
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

func (s *backtestService) ExecuteBacktest(cfg *config.Settings) error {
	// TODO: Implement actual backtest execution
	return nil
}
