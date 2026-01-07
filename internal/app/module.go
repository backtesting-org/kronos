package app

import (
	"github.com/backtesting-org/kronos-cli/internal/handlers"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/backtest"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/live"
	"github.com/backtesting-org/kronos-cli/internal/router"
	"github.com/backtesting-org/kronos-cli/internal/services/compile"
	"github.com/backtesting-org/kronos-cli/internal/setup"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"go.uber.org/fx"
)

// Module provides all application dependencies by composing domain modules
var Module = fx.Options(
	backtest.Module,
	setup.Module,
	shared.Module,
	handlers.Module,
	router.Module,
	live.Module,
	strategies.Module,
	compile.Module,
)
