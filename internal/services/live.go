package services

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/backtesting-org/kronos-cli/internal/live"
)

// LiveService handles live trading operations
type LiveService struct{
	compileSvc *CompileService
}

func NewLiveService(compileSvc *CompileService) *LiveService {
	return &LiveService{
		compileSvc: compileSvc,
	}
}

func (s *LiveService) RunSelectionTUI() error {
	// Pre-compile all strategies BEFORE showing the TUI
	fmt.Println("🔍 Checking strategies...")
	s.compileSvc.PreCompileStrategies("./strategies")
	fmt.Println()

	// Try to discover strategies from ./strategies directory
	strategies, err := live.DiscoverStrategies()
	if err != nil {
		// No strategies found, but continue to show empty state
		strategies = []live.Strategy{}
	}

	// Load global exchanges config (or create empty one)
	exchangesConfigPath := "./exchanges.yml"
	globalExchanges, err := live.LoadGlobalExchangesConfig(exchangesConfigPath)
	if err != nil {
		// Create empty config
		globalExchanges = &live.GlobalExchangesConfig{Exchanges: []live.ExchangeConfig{}}
	}

	// Create model with injected compile service
	m := live.NewSelectionModel(strategies, globalExchanges, s.compileSvc)

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	// Check for errors in final model
	if model, ok := finalModel.(live.SelectionModel); ok {
		if model.Err() != nil {
			// Check if user requested project initialization
			if model.Err().Error() == "INIT_PROJECT_REQUESTED" {
				return fmt.Errorf("INIT_PROJECT_REQUESTED")
			}
			return model.Err()
		}

		// Save credentials back to global exchanges.yml if a strategy was successfully configured
		if model.Selected() != nil && model.SelectedExchange() != nil && model.CurrentScreen() == live.ScreenSuccess {
			if err := live.SaveGlobalExchangesConfig(exchangesConfigPath, model.GlobalExchanges()); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			// Execute live trading
			fmt.Println("\n🚀 Starting live trading...")
			if err := live.ExecuteLiveTrading(model.Selected(), model.SelectedExchange(), s.compileSvc); err != nil {
				return fmt.Errorf("failed to execute live trading: %w", err)
			}
		}
	}

	return nil
}

func (s *LiveService) GetStrategies() ([]live.Strategy, error) {
	// Discover strategies from ./strategies directory
	return live.DiscoverStrategies()
}
