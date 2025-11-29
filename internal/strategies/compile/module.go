package compile

import (
	"go.uber.org/fx"
)

// Module provides compile view constructor in DI
var Module = fx.Module("compile",
	fx.Provide(
		NewCompileModel,
	),
)
