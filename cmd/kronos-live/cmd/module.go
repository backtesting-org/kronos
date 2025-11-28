package cmd

import (
	"go.uber.org/fx"
)

var Module = fx.Module("kronos-live-cmd",
	fx.Provide(
		NewLiveCommand,
		NewRootCommand,
	),
)
