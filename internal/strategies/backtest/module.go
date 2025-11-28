package backtest

import (
	handlers2 "github.com/backtesting-org/kronos-cli/internal/strategies/backtest/handlers"
	services2 "github.com/backtesting-org/kronos-cli/internal/strategies/backtest/services"
	"go.uber.org/fx"
)

// Module provides all backtesting dependencies
var Module = fx.Module("backtesting",
	// Services
	fx.Provide(services2.NewBacktestService),
	fx.Provide(services2.NewAnalyzeService),

	// Handlers
	fx.Provide(handlers2.NewBacktestHandler),
	fx.Provide(handlers2.NewAnalyzeHandler),
)
