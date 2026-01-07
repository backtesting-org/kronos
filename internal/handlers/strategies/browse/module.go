package browse

import (
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/compile"
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/live"
	strategyTypes "github.com/backtesting-org/kronos-cli/pkg/strategy"
	"github.com/backtesting-org/kronos-sdk/pkg/types/config"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/fx"
)

// Factory types - each defines the contract for creating a specific view
// Factories capture singleton dependencies (via DI) and take transient state as parameters
type StrategyListViewFactory func() tea.Model
type StrategyDetailViewFactory func(*config.Strategy) tea.Model

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
	strategyService config.StrategyConfig,
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
	compileFactory compile.CompileViewFactory,
	liveFactory live.LiveViewFactory,
) StrategyDetailViewFactory {
	return func(s *config.Strategy) tea.Model {
		return newStrategyDetailView(
			compileFactory,
			liveFactory,
			s,
		)
	}
}
