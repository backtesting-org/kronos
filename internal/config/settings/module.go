package settings

import "go.uber.org/fx"

var Module = fx.Module("settings",
	fx.Provide(
		NewConfiguration,
	),
)
