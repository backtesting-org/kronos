package connectors

import "go.uber.org/fx"

var Module = fx.Module("connectors",
	fx.Provide(
		NewConnectorService,
		NewValidationService,
	),
)
