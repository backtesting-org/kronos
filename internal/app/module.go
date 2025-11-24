package app

import (
	"github.com/backtesting-org/kronos-cli/internal/handlers"
	"github.com/backtesting-org/kronos-cli/internal/services"
	"go.uber.org/fx"
)

// Module provides all application dependencies
var Module = fx.Options(
	// Services
	fx.Provide(services.NewScaffoldService),
	fx.Provide(services.NewLiveService),
	fx.Provide(services.NewBacktestService),
	fx.Provide(services.NewAnalyzeService),

	// Handlers
	fx.Provide(handlers.NewInitHandler),
	fx.Provide(handlers.NewLiveHandler),
	fx.Provide(handlers.NewBacktestHandler),
	fx.Provide(handlers.NewAnalyzeHandler),
	fx.Provide(handlers.NewRootHandler),
)
