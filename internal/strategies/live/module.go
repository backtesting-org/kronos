package live

import (
	"github.com/backtesting-org/kronos-cli/internal/strategies/live/handlers"
	"github.com/backtesting-org/kronos-cli/internal/strategies/live/handlers/cli"
	"github.com/backtesting-org/kronos-cli/internal/strategies/live/handlers/interactive"
	"github.com/backtesting-org/kronos-cli/internal/strategies/live/runtime"
	"github.com/backtesting-org/kronos-cli/internal/strategies/live/services"
	"github.com/backtesting-org/live-trading/pkg"
	"go.uber.org/fx"
)

// Module provides all live trading dependencies including connectors registry and runtime
var Module = fx.Module("live",
	// Core SDK dependencies - provides connector registry
	pkg.Module,

	// Runtime for strategy execution
	runtime.Module,

	// Services
	fx.Provide(services.NewLiveService),

	// Sub-handlers (CLI and TUI)
	fx.Provide(cli.NewCLILiveHandler),
	fx.Provide(interactive.NewTUIHandler),

	// Unified handler that routes between CLI and TUI
	fx.Provide(handlers.NewLiveHandler),
)
