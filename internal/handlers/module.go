package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/strategies"
	"go.uber.org/fx"
)

// Module provides the root handler
var Module = fx.Module("handlers",
	fx.Provide(strategies.NewStrategyBrowser),
	fx.Provide(NewRootHandler),
)
