package compile

import "go.uber.org/fx"

// Module provides all backtesting dependencies
var Module = fx.Module("compile",
	// Services
	fx.Provide(
		NewCompileService,
	),
)
