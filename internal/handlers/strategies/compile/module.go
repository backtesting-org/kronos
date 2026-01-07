package compile

import (
	strategyTypes "github.com/backtesting-org/kronos-cli/pkg/strategy"
	"github.com/backtesting-org/kronos-sdk/pkg/types/config"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/fx"
)

// CompileViewFactory creates compile views with transient strategy data
type CompileViewFactory func(*config.Strategy) tea.Model

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
	return func(s *config.Strategy) tea.Model {
		model := NewCompileModel(compileService)
		model.SetStrategy(s)
		return model
	}
}
