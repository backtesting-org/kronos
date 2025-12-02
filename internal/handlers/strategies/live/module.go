package live

import (
	"github.com/backtesting-org/kronos-cli/internal/services/live"
	"github.com/backtesting-org/kronos-cli/internal/services/live/manager"
	"github.com/backtesting-org/kronos-cli/internal/services/live/runtime"
	"github.com/backtesting-org/live-trading/pkg"
	"go.uber.org/fx"
)

// Module provides all live trading dependencies including connectors registry and runtime
var Module = fx.Module("live",
	// Core SDK dependencies - provides connector registry
	pkg.Module,

	// Instance manager for multi-instance tracking and spawning
	manager.Module,

	// Runtime for strategy execution
	runtime.Module,

	// Services
	fx.Provide(live.NewLiveService),

	fx.Provide(
		NewLiveViewFactory,
	),
)
