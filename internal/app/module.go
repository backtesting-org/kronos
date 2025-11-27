package app

import (
	"github.com/backtesting-org/kronos-cli/internal/backtesting"
	"github.com/backtesting-org/kronos-cli/internal/live"
	"github.com/backtesting-org/kronos-cli/internal/setup"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"go.uber.org/fx"
)

// Module provides all application dependencies by composing domain modules
var Module = fx.Options(
	backtesting.Module,
	live.Module,
	setup.Module,
	shared.Module,
)
