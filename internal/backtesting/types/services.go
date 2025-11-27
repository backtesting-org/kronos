package types

import (
	"github.com/backtesting-org/kronos-cli/internal/config"
)

type AnalyzeService interface {
	AnalyzeResults(path string) error
}

type BacktestService interface {
	RunInteractive() error
	ExecuteBacktest(cfg *config.Config) error
	LoadConfig(path string) (*config.Config, error)
}
