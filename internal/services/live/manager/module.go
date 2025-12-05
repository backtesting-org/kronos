package manager

import (
	"context"

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
	fx.Invoke(initializeInstanceManager),
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

// initializeInstanceManager loads running instances from state file on startup
func initializeInstanceManager(lc fx.Lifecycle, manager live.InstanceManager) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Load instances from state file if they exist
			if err := manager.LoadRunning(ctx); err != nil {
				// Don't fail startup if we can't load instances - just log it
				// The state file might not exist on first run
				return nil
			}
			return nil
		},
	})
}
