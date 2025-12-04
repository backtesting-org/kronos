package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/monitor"
	"go.uber.org/fx"
)

// Module provides the root handler
var Module = fx.Module("handlers",
	// Monitor view factory
	monitor.Module,

	fx.Provide(strategies.NewStrategyBrowser),
	fx.Provide(NewRootHandler),
)
