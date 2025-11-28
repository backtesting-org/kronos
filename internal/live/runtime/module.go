package runtime

import (
	"go.uber.org/fx"
)

// Module provides our runtime wrapper
// The actual SDK runtime (sdkRuntime.Runtime) is provided by pkg.Module
var Module = fx.Module("runtime",
	fx.Provide(
		NewRuntime,
	),
)
