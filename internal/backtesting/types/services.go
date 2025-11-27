package types

import (
	"github.com/backtesting-org/kronos-cli/internal/config/settings"
)

type AnalyzeService interface {
	AnalyzeResults(path string) error
}

type BacktestService interface {
	RunInteractive() error
	ExecuteBacktest(cfg *settings.Settings) error
}
