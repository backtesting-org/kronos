package runtime

import (
	"go.uber.org/fx"
)

// Module provides our startup wrapper
// The actual SDK startup (sdkRuntime.Runtime) is provided by pkg.Module
var Module = fx.Module("startup",
	fx.Provide(
		NewRuntime,
	),
)
