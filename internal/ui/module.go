package ui

import (
	"github.com/backtesting-org/kronos-cli/internal/ui/factory"
	"github.com/backtesting-org/kronos-cli/internal/ui/router"
)

// Module provides UI-related services including routing
var Module = fx.Module("ui",
	router.Module,
	factory.Module,
)
