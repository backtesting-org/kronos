package setup

import (
	"github.com/backtesting-org/kronos-cli/internal/setup/handlers"
	"github.com/backtesting-org/kronos-cli/internal/setup/services"
	"go.uber.org/fx"
)

// Module provides all setup/scaffolding dependencies
var Module = fx.Module("setup",
	// Services
	fx.Provide(services.NewScaffoldService),

	// Handlers
	fx.Provide(handlers.NewInitHandler),
)
