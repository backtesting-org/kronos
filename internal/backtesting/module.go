package backtesting

import (
	"github.com/backtesting-org/kronos-cli/internal/backtesting/handlers"
	"github.com/backtesting-org/kronos-cli/internal/backtesting/services"
	"go.uber.org/fx"
)

// Module provides all backtesting dependencies
var Module = fx.Module("backtesting",
	// Services
	fx.Provide(services.NewBacktestService),
	fx.Provide(services.NewAnalyzeService),

	// Handlers
	fx.Provide(handlers.NewBacktestHandler),
	fx.Provide(handlers.NewAnalyzeHandler),
)
