package manager

import (
	"github.com/backtesting-org/kronos-cli/pkg/live"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"go.uber.org/fx"
)

// Module provides the manager components via Fx
var Module = fx.Module(
	"live/manager",
	fx.Provide(
		NewFileStateStore,
		NewProcessSpawner,
		provideInstanceManager,
	),
)

type instanceManagerParams struct {
	fx.In
	StateStore live.StateStore
	Spawner    live.ProcessSpawner
	Logger     logging.ApplicationLogger
}

func provideInstanceManager(params instanceManagerParams) live.InstanceManager {
	return NewInstanceManager(params.StateStore, params.Spawner, params.Logger)
}
