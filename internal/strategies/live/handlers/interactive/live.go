package interactive

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/strategies/live/types"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/internal/ui/router"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LiveInteractive interface {
	Run() error
}

type live struct {
	service types.LiveService
}

func NewTUIHandler(service types.LiveService) LiveInteractive {
	return &live{
		service: service,
	}
}

// Run executes the TUI flow - this is where Tea orchestration lives
func (l *live) Run() error {
	// 1. Load data from service
	strategies, err := l.service.FindStrategies()
	if err != nil {
		return err
	}

	model := NewSelectionModel(strategies, l.service)
	program := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// 3. Handle result
	result, ok := finalModel.(SelectionModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if result.Err() != nil {
		return result.Err()
	}

	// 4. Execute if user selected something
	if result.Selected() != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle Ctrl+C
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			fmt.Println("\n\n🛑 Stopping strategy...")
			cancel()
		}()

		// Service handles all exchange/connector logic
		return l.service.ExecuteStrategy(ctx, result.Selected(), nil)
	}

	return nil
}

// NewSelectionModel creates a new strategy selection model (view only)
func NewSelectionModel(strategies []strategy.Strategy, service types.LiveService) SelectionModel {
	// If no strategies, start with empty state screen
	initialScreen := ScreenSelection
	if len(strategies) == 0 {
		initialScreen = ScreenEmptyState
	}

	return SelectionModel{
		strategies:    strategies,
		cursor:        0,
		scrollOffset:  0,
		selected:      nil,
		currentScreen: initialScreen,
		service:       service,
	}
}

// Err returns any error from the model
func (m SelectionModel) Err() error {
	return m.err
}

// Selected returns the selected strategy
func (m SelectionModel) Selected() *strategy.Strategy {
	return m.selected
}

// CurrentScreen returns the current screen
func (m SelectionModel) CurrentScreen() Screen {
	return m.currentScreen
}

func (m SelectionModel) Init() tea.Cmd {
	return nil
}

func (m SelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.currentScreen {
		case ScreenEmptyState:
			return m.updateEmptyState(msg)
		case ScreenSelection:
			return m.updateSelection(msg)
		}
	}

	return m, nil
}

func (m SelectionModel) updateEmptyState(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// User wants to initialize a new project
		if m.router != nil {
			// Navigate to init using router
			return m, m.router.Navigate(router.RouteInit)
		}
		// Fallback: set error flag that will be checked after TUI exits
		m.err = fmt.Errorf("INIT_PROJECT_REQUESTED")
		return m, tea.Quit

	case "q", "Q", "ctrl+c", "esc":
		return m, tea.Quit
	}

	return m, nil
}

func (m SelectionModel) updateSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If validation error is showing, handle specially
	if m.validationErr != "" {
		// Allow quit even with error showing
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Any other key clears the error
		m.validationErr = ""
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			// Scroll up if cursor goes above visible area
			if m.cursor < m.scrollOffset {
				m.scrollOffset = m.cursor
			}
		}

	case "down", "j":
		if m.cursor < len(m.strategies)-1 {
			m.cursor++
			// Scroll down if cursor goes below visible area
			if m.cursor >= m.scrollOffset+visibleStrategies {
				m.scrollOffset = m.cursor - (visibleStrategies - 1)
			}
		}

	case "enter", " ":
		// Check if the selected strategy has an error
		if m.strategies[m.cursor].Status == strategy.StatusError {
			// Don't allow selection of strategies with errors
			return m, nil
		}

		// Validate connectors BEFORE quitting the TUI
		selectedStrategy := &m.strategies[m.cursor]

		if err := m.service.ValidateStrategy(selectedStrategy); err != nil {
			// Validation failed - show error in TUI instead of quitting
			m.validationErr = err.Error()
			return m, nil
		}

		// Validation passed - select and quit
		m.selected = selectedStrategy
		return m, tea.Quit
	}

	return m, nil
}

func (m SelectionModel) View() string {
	switch m.currentScreen {
	case ScreenEmptyState:
		return m.renderEmptyState()
	case ScreenSelection:
		return m.renderSelection()
	default:
		return "Unknown screen"
	}
}

func (m SelectionModel) renderEmptyState() string {
	var b strings.Builder

	// Title
	title := ui.TitleStyle.Render("🚀 KRONOS LIVE TRADING")
	subtitle := ui.SubtitleStyle.Render("No strategies found")

	b.WriteString("\n\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, title))
	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, subtitle))
	b.WriteString("\n\n")

	// Message box
	message := []string{
		"",
		ui.ConfirmFieldStyle.Render("No strategies found in ./strategies/"),
		"",
		ui.StrategyMetaStyle.Render("To deploy strategies to live trading, you need to:"),
		"",
		ui.StrategyDescStyle.Render("  1. Initialize a new Kronos project"),
		ui.StrategyDescStyle.Render("  2. Create strategies in ./strategies/ directory"),
		ui.StrategyDescStyle.Render("  3. Configure exchanges.yml with your credentials"),
		"",
		"",
		ui.ConfirmFieldStyle.Render("Would you like to initialize a new Kronos project here?"),
		"",
	}

	b.WriteString(strings.Join(message, "\n"))

	// Options
	b.WriteString("\n")
	yesOption := ui.StrategyNameSelectedStyle.Render("Y") + ui.StrategyDescStyle.Render(" Yes, initialize new project")
	noOption := ui.StrategyMetaStyle.Render("Q") + ui.StrategyDescStyle.Render(" No, quit")

	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, yesOption))
	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, noOption))
	b.WriteString("\n\n")

	// Help
	help := ui.HelpStyle.Render("Y Initialize  Q Quit")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, help))

	// Wrap in box
	boxed := ui.BoxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxed)
}

func (m SelectionModel) renderSelection() string {
	var b strings.Builder

	// Title
	title := ui.TitleStyle.Render("🚀 KRONOS LIVE TRADING")
	subtitle := ui.SubtitleStyle.Render("Select a strategy to deploy")

	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, title))
	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, subtitle))
	b.WriteString("\n\n")

	// Strategy list - show only visible window of strategies
	visibleStart := m.scrollOffset
	visibleEnd := m.scrollOffset + visibleStrategies
	if visibleEnd > len(m.strategies) {
		visibleEnd = len(m.strategies)
	}

	// Show scroll indicators
	if m.scrollOffset > 0 {
		b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, ui.HelpStyle.Render("↑ More strategies above")))
		b.WriteString("\n")
	}

	for i := visibleStart; i < visibleEnd; i++ {
		strat := m.strategies[i]
		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
		}

		// Build strategy item
		statusIndicator := ui.GetStatusIndicator(strat.Status)

		name := strat.Name
		if m.cursor == i {
			name = ui.StrategyNameSelectedStyle.Render(name)
		} else {
			name = ui.StrategyNameStyle.Render(name)
		}

		description := ui.StrategyDescStyle.Render(strat.Description)

		// Build exchanges info
		var exchangesInfo string
		if len(strat.Exchanges) > 0 {
			exchangeNames := strings.Join(strat.Exchanges, ", ")
			exchangesInfo = ui.StrategyMetaStyle.Render(fmt.Sprintf("Exchanges: %s", exchangeNames))
		}

		itemContent := fmt.Sprintf(
			"%s  %s\n%s",
			statusIndicator,
			name,
			description,
		)

		// Add exchanges info
		if exchangesInfo != "" {
			itemContent += "\n" + exchangesInfo
		}

		// Add error message if strategy has an error and is selected
		if strat.Status == strategy.StatusError && strat.Error != "" && m.cursor == i {
			errorMsg := ui.StatusErrorStyle.Render(fmt.Sprintf("⚠ Error: %s", strat.Error))
			itemContent += "\n" + errorMsg
		}

		var item string
		if m.cursor == i {
			item = ui.StrategyItemSelectedStyle.Render(cursor + itemContent)
		} else {
			item = ui.StrategyItemStyle.Render(cursor + itemContent)
		}

		// Center the item
		b.WriteString(lipgloss.Place(m.width, lipgloss.Height(item), lipgloss.Center, lipgloss.Top, item))
		b.WriteString("\n")
	}

	// Show scroll indicator if more below
	if visibleEnd < len(m.strategies) {
		b.WriteString("\n")
		b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, ui.HelpStyle.Render("↓ More strategies below")))
	}

	// Help text
	help := ui.HelpStyle.Render("↑↓/jk Navigate  ↵ Select  q Quit")

	// Show validation error if present
	if m.validationErr != "" {
		// Build a nice error box
		errorLines := []string{
			"",
			ui.StatusErrorStyle.Render("⚠  Cannot Start Strategy"),
			"",
			ui.StrategyDescStyle.Render(m.validationErr),
			"",
			ui.HelpStyle.Render("Press any key to go back..."),
			"",
		}

		errorContent := strings.Join(errorLines, "\n")
		errorBox := ui.BoxStyle.Render(errorContent)

		b.WriteString("\n\n")
		b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, errorBox))
		return b.String()
	}

	// Check if current strategy has error to show additional help
	if m.cursor < len(m.strategies) && m.strategies[m.cursor].Status == strategy.StatusError {
		help += "\n" + ui.StatusErrorStyle.Render("  ⚠ Cannot select strategy with errors")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, help))

	return b.String()
}
