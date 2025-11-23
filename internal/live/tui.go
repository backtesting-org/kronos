package live

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen represents which screen we're on
type Screen int

const (
	ScreenSelection Screen = iota
	ScreenConfirmation
	ScreenDeploying
	ScreenSuccess
)

// SelectionModel is the Bubble Tea model for strategy selection
type SelectionModel struct {
	strategies    []Strategy
	cursor        int
	scrollOffset  int
	selected      *Strategy
	currentScreen Screen
	confirmInput  string
	width         int
	height        int
	err           error
}

// NewSelectionModel creates a new strategy selection model
func NewSelectionModel(strategies []Strategy) SelectionModel {
	return SelectionModel{
		strategies:    strategies,
		cursor:        0,
		currentScreen: ScreenSelection,
		width:         80,
		height:        24,
	}
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
		case ScreenSelection:
			return m.updateSelection(msg)
		case ScreenConfirmation:
			return m.updateConfirmation(msg)
		case ScreenSuccess:
			if msg.String() == "q" || msg.String() == "enter" {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m SelectionModel) updateSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
			// Show max 3 strategies at a time
			if m.cursor >= m.scrollOffset+3 {
				m.scrollOffset = m.cursor - 2
			}
		}

	case "enter", " ":
		// Select the current strategy
		m.selected = &m.strategies[m.cursor]
		m.currentScreen = ScreenConfirmation
		m.confirmInput = ""
	}

	return m, nil
}

func (m SelectionModel) updateConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentScreen = ScreenSelection
		m.selected = nil
		m.confirmInput = ""

	case "ctrl+c":
		return m, tea.Quit

	case "backspace":
		if len(m.confirmInput) > 0 {
			m.confirmInput = m.confirmInput[:len(m.confirmInput)-1]
		}

	case "enter":
		// Proceed to success screen
		m.currentScreen = ScreenSuccess
		return m, nil
	}

	return m, nil
}

func (m SelectionModel) View() string {
	switch m.currentScreen {
	case ScreenSelection:
		return m.renderSelection()
	case ScreenConfirmation:
		return m.renderConfirmation()
	case ScreenSuccess:
		return m.renderSuccess()
	default:
		return "Unknown screen"
	}
}

func (m SelectionModel) renderSelection() string {
	var b strings.Builder

	// Title
	title := TitleStyle.Render("🚀 KRONOS LIVE TRADING")
	subtitle := SubtitleStyle.Render("Select a strategy to deploy")

	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, title))
	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, subtitle))
	b.WriteString("\n\n")

	// Strategy list - show only visible window of strategies
	visibleStart := m.scrollOffset
	visibleEnd := m.scrollOffset + 3
	if visibleEnd > len(m.strategies) {
		visibleEnd = len(m.strategies)
	}

	// Show scroll indicators
	if m.scrollOffset > 0 {
		b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, HelpStyle.Render("↑ More strategies above")))
		b.WriteString("\n")
	}

	for i := visibleStart; i < visibleEnd; i++ {
		strategy := m.strategies[i]
		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
		}

		// Build strategy item
		statusIndicator := GetStatusIndicator(strategy.Status)

		name := strategy.Name
		if m.cursor == i {
			name = StrategyNameSelectedStyle.Render(name)
		} else {
			name = StrategyNameStyle.Render(name)
		}

		description := StrategyDescStyle.Render(strategy.Description)

		// Exchange and asset info
		exchangeNames := []string{}
		assetCount := 0
		for _, ex := range strategy.Exchanges {
			if ex.Enabled {
				exchangeNames = append(exchangeNames, ex.Name)
				assetCount += len(ex.Assets)
			}
		}
		meta := StrategyMetaStyle.Render(fmt.Sprintf(
			"Exchanges: %s | Assets: %d",
			strings.Join(exchangeNames, ", "),
			assetCount,
		))

		itemContent := fmt.Sprintf(
			"%s  %s\n%s\n%s",
			statusIndicator,
			name,
			description,
			meta,
		)

		var item string
		if m.cursor == i {
			item = StrategyItemSelectedStyle.Render(cursor + itemContent)
		} else {
			item = StrategyItemStyle.Render(cursor + itemContent)
		}

		// Center the item
		b.WriteString(lipgloss.Place(m.width, lipgloss.Height(item), lipgloss.Center, lipgloss.Top, item))
		b.WriteString("\n")
	}

	// Show scroll indicator if more below
	if visibleEnd < len(m.strategies) {
		b.WriteString("\n")
		b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, HelpStyle.Render("↓ More strategies below")))
	}

	// Help text
	help := HelpStyle.Render("↑↓/jk Navigate  ↵ Select  q Quit")
	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, help))

	return b.String()
}

func (m SelectionModel) renderConfirmation() string {
	if m.selected == nil {
		return "No strategy selected"
	}

	var b strings.Builder

	// Title
	title := ConfirmTitleStyle.Render("⚠️  CONFIRM DEPLOYMENT")
	b.WriteString("\n\n")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Strategy details
	details := []string{}

	// Strategy name
	details = append(details, fmt.Sprintf("%s  %s",
		ConfirmFieldStyle.Render("Strategy:"),
		ConfirmValueStyle.Render(m.selected.Name),
	))

	// Exchanges
	for _, ex := range m.selected.Exchanges {
		if ex.Enabled {
			details = append(details, fmt.Sprintf("%s  %s",
				ConfirmFieldStyle.Render("Exchange:"),
				ConfirmValueStyle.Render(ex.Name),
			))
			details = append(details, fmt.Sprintf("%s  %s",
				ConfirmFieldStyle.Render("Assets:"),
				ConfirmValueStyle.Render(strings.Join(ex.Assets, ", ")),
			))
		}
	}

	// Mode
	modeIndicator := GetModeIndicator(m.selected.Config.Execution.DryRun)
	details = append(details, fmt.Sprintf("%s  %s",
		ConfirmFieldStyle.Render("Mode:"),
		modeIndicator,
	))

	// Risk limits
	details = append(details, "")
	details = append(details, ConfirmFieldStyle.Render("Risk Limits:"))
	details = append(details, fmt.Sprintf("  Max Position: %s",
		ConfirmValueStyle.Render(fmt.Sprintf("$%.0f", m.selected.Config.Risk.MaxPositionSize)),
	))
	details = append(details, fmt.Sprintf("  Max Daily Loss: %s",
		ConfirmValueStyle.Render(fmt.Sprintf("$%.0f", m.selected.Config.Risk.MaxDailyLoss)),
	))

	detailsText := strings.Join(details, "\n")
	b.WriteString(detailsText)
	b.WriteString("\n")

	// Warning if live trading
	if !m.selected.Config.Execution.DryRun {
		warning := ConfirmWarningStyle.Render("🔴 This will execute real trades with real money!")
		b.WriteString("\n")
		b.WriteString(warning)
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	help := HelpStyle.Render("↵ Proceed  esc Cancel  ctrl+c Quit")
	b.WriteString(help)

	// Wrap in box
	boxed := ConfirmBoxStyle.Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxed)
}

func (m SelectionModel) renderSuccess() string {
	if m.selected == nil {
		return "No strategy selected"
	}

	var b strings.Builder

	// Success message
	successIcon := StatusReadyStyle.Render("✓")
	title := StrategyNameSelectedStyle.Render(fmt.Sprintf("Strategy '%s' deployed successfully!", m.selected.Name))

	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%s  %s", successIcon, title))
	b.WriteString("\n\n")

	// What would happen next
	nextSteps := []string{
		"",
		"Command that would be executed:",
		"",
		SubtitleStyle.Render(fmt.Sprintf("  live-trading --exchange %s \\", m.selected.Exchanges[0].Name)),
		SubtitleStyle.Render(fmt.Sprintf("               --strategy %s/strategy.so \\", m.selected.Path)),
		SubtitleStyle.Render(fmt.Sprintf("               --config %s/live.yml \\", m.selected.Path)),
		SubtitleStyle.Render(fmt.Sprintf("               --mode %s", m.selected.Config.Execution.Mode)),
		"",
		"",
		StrategyMetaStyle.Render("(This is a demo - no actual deployment yet)"),
		"",
	}

	b.WriteString(strings.Join(nextSteps, "\n"))

	help := HelpStyle.Render("↵ Continue  q Quit")
	b.WriteString("\n")
	b.WriteString(help)

	boxed := BoxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxed)
}

// RunSelectionTUI runs the strategy selection TUI
func RunSelectionTUI() error {
	// Get mock strategies for demo
	strategies := GetMockStrategies()

	// Create model
	m := NewSelectionModel(strategies)

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	// Check for errors in final model
	if model, ok := finalModel.(SelectionModel); ok {
		if model.err != nil {
			return model.err
		}
	}

	return nil
}

// runSimpleMenu is a fallback for environments without TTY (like IDEs)
func runSimpleMenu(strategies []Strategy) error {
	fmt.Println("\n🚀 KRONOS LIVE TRADING")
	fmt.Println("======================\n")
	fmt.Println("Select a strategy to deploy:\n")

	for i, strategy := range strategies {
		mode := "📝 PAPER"
		if !strategy.Config.Execution.DryRun {
			mode = "🔴 LIVE"
		}

		exchangeNames := []string{}
		assetCount := 0
		for _, ex := range strategy.Exchanges {
			if ex.Enabled {
				exchangeNames = append(exchangeNames, ex.Name)
				assetCount += len(ex.Assets)
			}
		}

		fmt.Printf("  [%d] %s\n", i+1, strategy.Name)
		fmt.Printf("      %s\n", strategy.Description)
		fmt.Printf("      Exchanges: %s | Assets: %d | Mode: %s\n\n",
			strings.Join(exchangeNames, ", "), assetCount, mode)
	}

	fmt.Print("Enter selection (1-3) or 'q' to quit: ")

	var input string
	fmt.Scanln(&input)

	if input == "q" || input == "Q" {
		return fmt.Errorf("cancelled by user")
	}

	var selection int
	_, err := fmt.Sscanf(input, "%d", &selection)
	if err != nil || selection < 1 || selection > len(strategies) {
		return fmt.Errorf("invalid selection")
	}

	selected := &strategies[selection-1]

	// Show confirmation
	fmt.Println("\n⚠️  CONFIRM DEPLOYMENT")
	fmt.Println("======================\n")
	fmt.Printf("Strategy:  %s\n", selected.Name)

	for _, ex := range selected.Exchanges {
		if ex.Enabled {
			fmt.Printf("Exchange:  %s\n", ex.Name)
			fmt.Printf("Assets:    %s\n", strings.Join(ex.Assets, ", "))
		}
	}

	if selected.Config.Execution.DryRun {
		fmt.Println("Mode:      📝 PAPER TRADING")
	} else {
		fmt.Println("Mode:      🔴 LIVE TRADING")
		fmt.Println("\n🔴 This will execute real trades with real money!")
	}

	fmt.Printf("\nRisk Limits:\n")
	fmt.Printf("  Max Position: $%.0f\n", selected.Config.Risk.MaxPositionSize)
	fmt.Printf("  Max Daily Loss: $%.0f\n", selected.Config.Risk.MaxDailyLoss)

	fmt.Print("\nType 'CONFIRM' to proceed: ")
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToUpper(confirm) != "CONFIRM" {
		return fmt.Errorf("deployment cancelled")
	}

	// Show success
	fmt.Printf("\n✓ Strategy '%s' deployed successfully!\n\n", selected.Name)
	fmt.Println("Command that would be executed:")
	fmt.Printf("  live-trading --exchange %s \\\n", selected.Exchanges[0].Name)
	fmt.Printf("               --strategy %s/strategy.so \\\n", selected.Path)
	fmt.Printf("               --config %s/live.yml \\\n", selected.Path)
	fmt.Printf("               --mode %s\n\n", selected.Config.Execution.Mode)
	fmt.Println("(This is a demo - no actual deployment yet)\n")

	return nil
}
