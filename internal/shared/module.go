package shared

import (
	"go.uber.org/fx"
)

// Module provides all backtesting dependencies
var Module = fx.Module("shared",
	// Services
	fx.Provide(
		NewCompileService,
		NewStrategyDiscovery,
		NewApplicationLogger,
	),
)
