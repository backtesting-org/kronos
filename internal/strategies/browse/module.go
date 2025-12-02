package browse

import (
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/strategies/compile"
	"github.com/backtesting-org/kronos-cli/internal/strategies/live"
	strategyTypes "github.com/backtesting-org/kronos-cli/pkg/strategy"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/fx"
)

// Factory types - each defines the contract for creating a specific view
// Factories capture singleton dependencies (via DI) and take transient state as parameters
type StrategyListViewFactory func() tea.Model
type StrategyDetailViewFactory func(*strategy.Strategy) tea.Model

// Module provides browse view factories in DI
var Module = fx.Module("browse",
	fx.Provide(
		NewStrategyListViewFactory,
		NewStrategyDetailViewFactory,
	),
)

// NewStrategyListViewFactory creates the factory function for list views
// All singleton dependencies are captured by the closure
func NewStrategyListViewFactory(
	compileService strategyTypes.CompileService,
	strategyService strategy.StrategyConfig,
	detailFactory StrategyDetailViewFactory,
) StrategyListViewFactory {
	return func() tea.Model {
		return newStrategyListView(
			compileService,
			strategyService,
			detailFactory,
		)
	}
}

// NewStrategyDetailViewFactory creates the factory function for detail views
// All singleton dependencies are captured by the closure
func NewStrategyDetailViewFactory(
	compileService strategyTypes.CompileService,
	compileFactory compile.CompileViewFactory,
	liveFactory live.LiveViewFactory,
) StrategyDetailViewFactory {
	return func(s *strategy.Strategy) tea.Model {
		return newStrategyDetailView(
			compileFactory,
			liveFactory,
			s,
		)
	}
}
