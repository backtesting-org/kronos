package ui

import (
	"github.com/backtesting-org/kronos-cli/internal/ui/router"
	"go.uber.org/fx"
)

// Module provides UI-related services including routing
var Module = fx.Module("ui",
	router.Module,
)
