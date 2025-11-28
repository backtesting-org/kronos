package app

import (
	"os"

	"github.com/backtesting-org/kronos-cli/internal/backtesting"
	"github.com/backtesting-org/kronos-cli/internal/config"
	"github.com/backtesting-org/kronos-cli/internal/handlers"
	"github.com/backtesting-org/kronos-cli/internal/live"
	"github.com/backtesting-org/kronos-cli/internal/setup"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/live-trading/pkg"
	"go.uber.org/fx"
)

// Module provides all application dependencies by composing domain modules
var Module = func() fx.Option {
	modules := []fx.Option{
		backtesting.Module,
		live.Module,
		setup.Module,
		shared.Module,
		handlers.Module,
		config.Module,
		ui.Module,
	}

	// Only load runtime module when running run-strategy command
	if len(os.Args) > 1 && os.Args[1] == "run-strategy" {
		modules = append(modules, pkg.Module)
	}

	return fx.Options(modules...)
}()
