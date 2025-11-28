package handlers

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/ui"
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
	router          router.Router
}

func NewStrategyBrowser(strategyService strategy.StrategyConfig, r router.Router) StrategyBrowser {
	return &strategyBrowser{
		strategyService: strategyService,
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

	// Create and run the orchestrator model
	m := newStrategiesModel(strategies, h.router)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	result := finalModel.(strategiesModel)
	if result.err != nil {
		return result.err
	}

	return nil
}

// strategiesModel is the top-level model that orchestrates navigation
type strategiesModel struct {
	list   strategyListModel
	detail strategyActionsModel
	screen string // "list" or "detail"
	router router.Router
	err    error
}

func newStrategiesModel(strategies []strategy.Strategy, r router.Router) strategiesModel {
	return strategiesModel{
		list:   strategyListModel{strategies: strategies, cursor: 0, pageSize: 10, pageNum: 1},
		screen: "list",
		router: r,
	}
}

func (m strategiesModel) Init() tea.Cmd {
	return nil
}

func (m strategiesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case router.NavigateMsg:
		// Handle navigation messages
		switch msg.Route {
		case router.RouteMenu:
			return m, tea.Quit
		}
	}

	// Route to current screen's update
	if m.screen == "list" {
		updated, cmd := m.list.Update(msg)
		if listModel, ok := updated.(strategyListModel); ok {
			m.list = listModel

			// Check if transitioned to detail screen
			if listModel.transitionToDetail {
				m.detail = newStrategyActionsModel(&listModel.strategies[listModel.cursor])
				m.screen = "detail"
				m.list.transitionToDetail = false
			}
		}
		return m, cmd
	} else if m.screen == "detail" {
		updated, cmd := m.detail.Update(msg)
		if detailModel, ok := updated.(strategyActionsModel); ok {
			m.detail = detailModel

			// Check if going back to list
			if detailModel.backToList {
				m.screen = "list"
				m.detail.backToList = false
			}
		}
		return m, cmd
	}

	return m, nil
}

func (m strategiesModel) View() string {
	if m.screen == "list" {
		return m.list.View()
	} else if m.screen == "detail" {
		return m.detail.View()
	}
	return ""
}

// strategyListModel represents the strategy list view (STRATEGIES screen)
type strategyListModel struct {
	strategies         []strategy.Strategy
	cursor             int
	pageSize           int
	pageNum            int
	transitionToDetail bool
	err                error
}

func (m strategyListModel) Init() tea.Cmd {
	return nil
}

func (m strategyListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.strategies)-1 {
				m.cursor++
			}
		case "enter":
			// Mark transition to detail screen
			m.transitionToDetail = true
		}
	}
	return m, nil
}

func (m strategyListModel) View() string {
	if len(m.strategies) == 0 {
		return ui.TitleStyle.Render("STRATEGIES") + "\n\n" + ui.SubtitleStyle.Render("No strategies found. Create a new one to get started.")
	}

	var content string
	content += ui.TitleStyle.Render("STRATEGIES") + "\n"
	content += ui.SubtitleStyle.Render("Use arrow keys to navigate, Enter to select, q to quit") + "\n\n"

	// Display current page
	for i, strat := range m.strategies {
		exchanges := fmt.Sprintf("[%v]", strat.Exchanges)
		if i == m.cursor {
			content += ui.StrategyNameSelectedStyle.Render("▶ "+strat.Name+" "+exchanges) + "\n"
		} else {
			content += ui.StrategyNameStyle.Render("  "+strat.Name+" "+exchanges) + "\n"
		}
	}

	// Show pagination info
	totalPages := (len(m.strategies) + m.pageSize - 1) / m.pageSize
	content += "\n" + ui.SubtitleStyle.Render(fmt.Sprintf("Page %d/%d", m.pageNum, totalPages))

	return ui.BoxStyle.Render(content)
}

// strategyActionsModel represents the strategy detail view with action options (STRATEGY screen)
type strategyActionsModel struct {
	strategy   *strategy.Strategy
	actions    []string
	cursor     int
	backToList bool
	err        error
}

func newStrategyActionsModel(strat *strategy.Strategy) strategyActionsModel {
	return strategyActionsModel{
		strategy:   strat,
		actions:    []string{"Compile", "Backtest", "Edit", "Delete"},
		cursor:     0,
		backToList: false,
	}
}

func (m strategyActionsModel) Init() tea.Cmd {
	return nil
}

func (m strategyActionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			// Mark to go back to list
			m.backToList = true
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			}
		case "enter":
			// TODO: Execute the selected action
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m strategyActionsModel) View() string {
	var content string
	content += ui.TitleStyle.Render(m.strategy.Name) + "\n"

	// Display strategy metadata
	if m.strategy.Description != "" {
		content += ui.StrategyDescStyle.Render("📝 "+m.strategy.Description) + "\n"
	}

	content += ui.StrategyMetaStyle.Render(fmt.Sprintf("🔗 Exchanges: %v", m.strategy.Exchanges)) + "\n"

	if len(m.strategy.Parameters) > 0 {
		content += ui.StrategyMetaStyle.Render(fmt.Sprintf("⚙️  Parameters: %d", len(m.strategy.Parameters))) + "\n"
	}

	content += "\n" + ui.SubtitleStyle.Render("Select action:") + "\n\n"

	for i, action := range m.actions {
		if i == m.cursor {
			content += ui.StrategyNameSelectedStyle.Render("▶ "+action) + "\n"
		} else {
			content += "  " + action + "\n"
		}
	}

	content += "\n" + ui.SubtitleStyle.Render("Enter to select, q to back, ctrl+c to quit")

	return ui.BoxStyle.Render(content)
}
