package live

import (
	"github.com/backtesting-org/kronos-cli/internal/live/handlers"
	"github.com/backtesting-org/kronos-cli/internal/live/handlers/cli"
	"github.com/backtesting-org/kronos-cli/internal/live/handlers/interactive"
	"github.com/backtesting-org/kronos-cli/internal/live/services"
	"go.uber.org/fx"
)

// Module provides all live trading dependencies
var Module = fx.Module("live",
	// Services
	fx.Provide(services.NewLiveService),

	// Sub-handlers (CLI and TUI)
	fx.Provide(cli.NewCLILiveHandler),
	fx.Provide(interactive.NewTUIHandler),

	// Unified handler that routes between CLI and TUI
	fx.Provide(handlers.NewLiveHandler),
)
