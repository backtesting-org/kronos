package backtesting

import (
	"github.com/backtesting-org/kronos-cli/internal/backtesting/handlers"
	"github.com/backtesting-org/kronos-cli/internal/backtesting/services"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"go.uber.org/fx"
)

// Module provides all backtesting dependencies
var Module = fx.Module("backtesting",
	// Services
	fx.Provide(services.NewBacktestService),
	fx.Provide(services.NewAnalyzeService),
	fx.Provide(shared.NewCompileService),

	// Handlers
	fx.Provide(handlers.NewBacktestHandler),
	fx.Provide(handlers.NewAnalyzeHandler),
)
