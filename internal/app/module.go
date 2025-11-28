package app

import (
	"github.com/backtesting-org/kronos-cli/internal/config"
	"github.com/backtesting-org/kronos-cli/internal/handlers"
	"github.com/backtesting-org/kronos-cli/internal/setup"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-cli/internal/strategies/backtest"
	"github.com/backtesting-org/kronos-cli/internal/strategies/live"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"go.uber.org/fx"
)

// Module provides all application dependencies by composing domain modules
var Module = fx.Options(
	backtest.Module,
	setup.Module,
	shared.Module,
	handlers.Module,
	config.Module,
	ui.Module,
	live.Module,
)
