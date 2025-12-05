package monitor

import (
	"go.uber.org/fx"
)

// Module provides monitor view dependencies
var Module = fx.Module("monitor-handlers",
	fx.Provide(
		NewMonitorViewFactory,
	),
)
