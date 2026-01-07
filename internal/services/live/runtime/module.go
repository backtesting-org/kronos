package runtime

import (
	"go.uber.org/fx"
)

// Module provides our startup wrapper
var Module = fx.Module("startup",
	fx.Provide(
		NewRuntime,
	),
)
