package strategies

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-cli/internal/strategies/browse"
	"github.com/backtesting-org/kronos-cli/internal/ui/router"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// StrategyBrowser handles browsing strategies and selecting actions
type StrategyBrowser interface {
	Handle(cmd *cobra.Command, args []string) error
}

type strategyBrowser struct {
	strategyService strategy.StrategyConfig
	compileService  shared.CompileService
	detailView      browse.StrategyDetailView
	browse          browse.StrategyListView
	router          router.Router
}

func NewStrategyBrowser(
	strategyService strategy.StrategyConfig,
	compileService shared.CompileService,
	detailView browse.StrategyDetailView,
	browse browse.StrategyListView,
	r router.Router,
) StrategyBrowser {
	return &strategyBrowser{
		strategyService: strategyService,
		compileService:  compileService,
		detailView:      detailView,
		browse:          browse,
		router:          r,
	}
}

func (h *strategyBrowser) Handle(_ *cobra.Command, _ []string) error {
	// Load all strategies
	strategies, err := h.strategyService.FindStrategies()
	if err != nil {
		return fmt.Errorf("failed to load strategies: %w", err)
	}

	if len(strategies) == 0 {
		return fmt.Errorf("no strategies found")
	}

	// Set the initial view on the router
	h.router.SetInitialView(h.browse)

	// Router IS the Tea model - pass it to the program
	p := tea.NewProgram(h.router, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
