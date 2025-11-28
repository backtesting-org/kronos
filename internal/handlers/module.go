package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/strategies/browse/handlers"
	"go.uber.org/fx"
)

// Module provides the root handler
var Module = fx.Module("handlers",
	fx.Provide(handlers.NewStrategyBrowser),
	fx.Provide(NewRootHandler),
)
