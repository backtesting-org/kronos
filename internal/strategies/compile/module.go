package compile

import (
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	strategyTypes "github.com/backtesting-org/kronos-cli/pkg/strategy"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/fx"
)

// CompileViewFactory creates compile views with transient strategy data
type CompileViewFactory func(*strategy.Strategy) tea.Model

// Module provides compile view constructor in DI
var Module = fx.Module("compile",
	fx.Provide(
		NewCompileViewFactory,
	),
)

// NewCompileViewFactory creates the factory function for compile views
func NewCompileViewFactory(
	compileService strategyTypes.CompileService,
) CompileViewFactory {
	return func(s *strategy.Strategy) tea.Model {
		model := NewCompileModel(compileService)
		model.SetStrategy(s)
		return model
	}
}
