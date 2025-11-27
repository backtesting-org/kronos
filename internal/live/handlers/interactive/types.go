package interactive

import (
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/ui/router"
)

// Screen represents which screen we're on
type Screen int

const (
	ScreenSelection Screen = iota
	ScreenEmptyState
)

const (
	// visibleStrategies is the maximum number of strategies shown at once
	visibleStrategies = 3
)

// SelectionModel is the Bubble Tea model for strategy selection
type SelectionModel struct {
	strategies    []strategy.Strategy
	cursor        int
	scrollOffset  int
	selected      *strategy.Strategy
	currentScreen Screen
	width         int
	height        int
	err           error
	router        router.Router
}
