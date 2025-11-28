package handlers

import (
	"go.uber.org/fx"
)

// Module provides the root handler
var Module = fx.Module("handlers",
	fx.Provide(NewRootHandler),
)
