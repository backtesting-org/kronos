package handlers

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// StrategyBrowser handles browsing strategies and selecting actions
type StrategyBrowser interface {
	Handle(cmd *cobra.Command, args []string) error
}

type strategyBrowser struct {
	strategyService strategy.StrategyConfig
}

func NewStrategyBrowser(strategyService strategy.StrategyConfig) StrategyBrowser {
	return &strategyBrowser{
		strategyService: strategyService,
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

	// Create and run the strategy list model
	m := newStrategyListModel(strategies)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	result := finalModel.(strategyListModel)
	if result.err != nil {
		return result.err
	}

	// TODO: When strategy is selected, show action menu
	return nil
}

// strategyListModel represents the strategy list view (STRATEGIES screen)
type strategyListModel struct {
	strategies []strategy.Strategy
	cursor     int
	pageSize   int
	pageNum    int
	err        error
}

func newStrategyListModel(strategies []strategy.Strategy) strategyListModel {
	return strategyListModel{
		strategies: strategies,
		cursor:     0,
		pageSize:   10,
		pageNum:    1,
	}
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
			// Transition to strategy actions screen
			selectedStrategy := m.strategies[m.cursor]
			actionsModel := newStrategyActionsModel(&selectedStrategy)
			return actionsModel, nil
		}
	}
	return m, nil
}

func (m strategyListModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		PaddingTop(1).
		PaddingBottom(1).
		Align(lipgloss.Center)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Width(70)

	itemStyle := lipgloss.NewStyle().
		PaddingLeft(2)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true).
		PaddingLeft(0)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	title := titleStyle.Render("STRATEGIES")

	if len(m.strategies) == 0 {
		return title + "\n\n" + mutedStyle.Render("No strategies found. Create a new one to get started.")
	}

	var content string
	content += title + "\n\n"
	content += mutedStyle.Render("Use arrow keys to navigate, Enter to select, q to quit") + "\n\n"

	// Display current page
	for i, strat := range m.strategies {
		exchanges := fmt.Sprintf("[%s]", fmt.Sprintf("%v", strat.Exchanges))
		if i == m.cursor {
			content += selectedStyle.Render("▶ "+strat.Name+" "+exchanges) + "\n"
		} else {
			content += itemStyle.Render("  "+strat.Name+" "+exchanges) + "\n"
		}
	}

	// Show pagination info
	totalPages := (len(m.strategies) + m.pageSize - 1) / m.pageSize
	content += "\n" + mutedStyle.Render(fmt.Sprintf("Page %d/%d", m.pageNum, totalPages))

	return boxStyle.Render(content)
}

// strategyActionsModel represents the strategy detail view with action options (STRATEGY screen)
type strategyActionsModel struct {
	strategy *strategy.Strategy
	actions  []string
	cursor   int
	err      error
}

func newStrategyActionsModel(strat *strategy.Strategy) strategyActionsModel {
	return strategyActionsModel{
		strategy: strat,
		actions:  []string{"Compile", "Backtest", "Edit", "Delete"},
		cursor:   0,
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
			// Go back to list
			return m, tea.Quit
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
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00D9FF")).
		PaddingTop(1).
		PaddingBottom(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Width(70)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D9FF")).
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#D1D5DB"))

	title := titleStyle.Render(m.strategy.Name)

	var content string
	content += title + "\n\n"

	// Display strategy metadata
	if m.strategy.Description != "" {
		content += infoStyle.Render("📝 "+m.strategy.Description) + "\n"
	}

	content += infoStyle.Render(fmt.Sprintf("🔗 Exchanges: %v", m.strategy.Exchanges)) + "\n"

	if len(m.strategy.Parameters) > 0 {
		content += infoStyle.Render(fmt.Sprintf("⚙️  Parameters: %d", len(m.strategy.Parameters))) + "\n"
	}

	content += "\n" + mutedStyle.Render("Select action:") + "\n\n"

	for i, action := range m.actions {
		if i == m.cursor {
			content += selectedStyle.Render("▶ "+action) + "\n"
		} else {
			content += "  " + action + "\n"
		}
	}

	content += "\n" + mutedStyle.Render("Enter to select, q to back, ctrl+c to quit")

	return boxStyle.Render(content)
}
