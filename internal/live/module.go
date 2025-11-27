package live

import (
	"github.com/backtesting-org/kronos-cli/internal/live/handlers/cli"
	"github.com/backtesting-org/kronos-cli/internal/live/handlers/interactive"
	"github.com/backtesting-org/kronos-cli/internal/live/services"
	"go.uber.org/fx"
)

// Module provides all live trading dependencies
var Module = fx.Module("live",
	// Services
	fx.Provide(services.NewLiveService),
	fx.Provide(services.NewConfigService),

	// Handlers
	fx.Provide(cli.NewLiveHandler),
	fx.Provide(interactive.NewTUIHandler),
)
