package router

import (
	"go.uber.org/fx"
)

// Module provides routing services
var Module = fx.Module("router",
	fx.Provide(
		NewRouter,
	),
)
