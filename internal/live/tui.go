package live

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen represents which screen we're on
type Screen int

const (
	ScreenSelection Screen = iota
	ScreenExchangeSelection
	ScreenCredentials
	ScreenConfirmation
	ScreenDeploying
	ScreenSuccess
)

const (
	// visibleStrategies is the maximum number of strategies shown at once
	visibleStrategies = 3
)

// SelectionModel is the Bubble Tea model for strategy selection
type SelectionModel struct {
	strategies         []Strategy
	cursor             int
	scrollOffset       int
	selected           *Strategy
	selectedExchange   *ExchangeConfig
	globalExchanges    *GlobalExchangesConfig  // Global exchanges config
	currentScreen      Screen
	confirmInput       string
	width              int
	height             int
	err                error

	// Exchange selection
	exchangeCursor     int

	// Credential input fields
	credentialFields   []string
	currentField       int
	fieldInputs        map[string]string
	showPassword       bool
}

// NewSelectionModel creates a new strategy selection model
func NewSelectionModel(strategies []Strategy, globalExchanges *GlobalExchangesConfig) SelectionModel {
	return SelectionModel{
		strategies:      strategies,
		globalExchanges: globalExchanges,
		cursor:          0,
		currentScreen:   ScreenSelection,
		width:           80,
		height:          24,
		fieldInputs:     make(map[string]string),
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
		case ScreenExchangeSelection:
			return m.updateExchangeSelection(msg)
		case ScreenCredentials:
			return m.updateCredentials(msg)
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
			if m.cursor >= m.scrollOffset+visibleStrategies {
				m.scrollOffset = m.cursor - (visibleStrategies - 1)
			}
		}

	case "enter", " ":
		// Select the current strategy
		m.selected = &m.strategies[m.cursor]

		// Get available exchanges for this strategy from global config
		availableExchanges := m.getAvailableExchangesForStrategy()

		if len(availableExchanges) == 1 {
			// Only one exchange, auto-select it and move to credentials
			m.selectedExchange = availableExchanges[0]
			m.setupCredentialFields()
			m.currentScreen = ScreenCredentials
		} else if len(availableExchanges) > 1 {
			// Multiple exchanges, let user choose
			m.currentScreen = ScreenExchangeSelection
			m.exchangeCursor = 0
		} else {
			// No exchanges configured
			m.err = fmt.Errorf("no exchanges configured for this strategy")
		}
	}

	return m, nil
}

func (m SelectionModel) updateExchangeSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	availableExchanges := m.getAvailableExchangesForStrategy()

	switch msg.String() {
	case "esc":
		m.currentScreen = ScreenSelection
		m.selected = nil

	case "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.exchangeCursor > 0 {
			m.exchangeCursor--
		}

	case "down", "j":
		if m.exchangeCursor < len(availableExchanges)-1 {
			m.exchangeCursor++
		}

	case "enter", " ":
		// Select the current exchange
		if m.exchangeCursor < len(availableExchanges) {
			m.selectedExchange = availableExchanges[m.exchangeCursor]
			m.setupCredentialFields()
			m.currentScreen = ScreenCredentials
		}
	}

	return m, nil
}

func (m SelectionModel) updateCredentials(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Go back to exchange selection or strategy selection
		if len(m.selected.Config.Exchanges) > 1 {
			m.currentScreen = ScreenExchangeSelection
		} else {
			m.currentScreen = ScreenSelection
		}
		m.selectedExchange = nil
		m.fieldInputs = make(map[string]string)

	case "ctrl+c":
		return m, tea.Quit

	case "tab", "down":
		// Move to next field
		if m.currentField < len(m.credentialFields)-1 {
			m.currentField++
		}

	case "shift+tab", "up":
		// Move to previous field
		if m.currentField > 0 {
			m.currentField--
		}

	case "backspace":
		// Delete character from current field
		currentFieldName := m.credentialFields[m.currentField]
		if len(m.fieldInputs[currentFieldName]) > 0 {
			m.fieldInputs[currentFieldName] = m.fieldInputs[currentFieldName][:len(m.fieldInputs[currentFieldName])-1]
		}

	case "enter":
		// Move to next field, or if on last field, proceed to confirmation
		if m.currentField < len(m.credentialFields)-1 {
			m.currentField++
		} else {
			// Save credentials to the exchange config
			for field, value := range m.fieldInputs {
				if m.selectedExchange != nil && m.selectedExchange.Credentials != nil {
					m.selectedExchange.Credentials[field] = value
				}
			}
			m.currentScreen = ScreenConfirmation
			m.confirmInput = ""
		}

	default:
		// Add typed character to current field input
		currentFieldName := m.credentialFields[m.currentField]
		m.fieldInputs[currentFieldName] += msg.String()
	}

	return m, nil
}

// getAvailableExchangesForStrategy returns the list of exchange configs available for the selected strategy
func (m *SelectionModel) getAvailableExchangesForStrategy() []*ExchangeConfig {
	if m.selected == nil || m.globalExchanges == nil {
		return []*ExchangeConfig{}
	}

	var available []*ExchangeConfig
	for _, exchangeName := range m.selected.Config.Exchanges {
		// Find this exchange in global config
		for i := range m.globalExchanges.Exchanges {
			if m.globalExchanges.Exchanges[i].Name == exchangeName && m.globalExchanges.Exchanges[i].Enabled {
				available = append(available, &m.globalExchanges.Exchanges[i])
				break
			}
		}
	}

	return available
}

// setupCredentialFields determines what credential fields are needed based on the selected exchange
func (m *SelectionModel) setupCredentialFields() {
	if m.selectedExchange == nil {
		return
	}

	m.currentField = 0
	m.fieldInputs = make(map[string]string)

	// Determine required fields based on exchange
	switch m.selectedExchange.Name {
	case "paradex":
		m.credentialFields = []string{
			"account_address",
			"eth_private_key",
			"l2_private_key",
		}
		// Pre-fill with existing values if any
		if m.selectedExchange.Credentials != nil {
			for _, field := range m.credentialFields {
				if val, ok := m.selectedExchange.Credentials[field]; ok {
					m.fieldInputs[field] = val
				} else {
					m.fieldInputs[field] = ""
				}
			}
		}

	case "bybit", "binance", "kraken":
		m.credentialFields = []string{
			"api_key",
			"api_secret",
		}
		// Pre-fill with existing values if any
		if m.selectedExchange.Credentials != nil {
			for _, field := range m.credentialFields {
				if val, ok := m.selectedExchange.Credentials[field]; ok {
					m.fieldInputs[field] = val
				} else {
					m.fieldInputs[field] = ""
				}
			}
		}

	default:
		// Generic API key/secret for unknown exchanges
		m.credentialFields = []string{
			"api_key",
			"api_secret",
		}
		m.fieldInputs["api_key"] = ""
		m.fieldInputs["api_secret"] = ""
	}
}

func (m SelectionModel) updateConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Go back to credentials screen
		m.currentScreen = ScreenCredentials
		m.confirmInput = ""

	case "ctrl+c":
		return m, tea.Quit

	case "backspace":
		if len(m.confirmInput) > 0 {
			m.confirmInput = m.confirmInput[:len(m.confirmInput)-1]
		}

	case "enter":
		// Validate confirmation input before proceeding
		if strings.ToUpper(m.confirmInput) == "CONFIRM" {
			m.currentScreen = ScreenSuccess
			return m, nil
		}

	default:
		// Add typed character to confirmation input
		m.confirmInput += msg.String()
	}

	return m, nil
}

func (m SelectionModel) View() string {
	switch m.currentScreen {
	case ScreenSelection:
		return m.renderSelection()
	case ScreenExchangeSelection:
		return m.renderExchangeSelection()
	case ScreenCredentials:
		return m.renderCredentials()
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
	visibleEnd := m.scrollOffset + visibleStrategies
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

func (m SelectionModel) renderExchangeSelection() string {
	if m.selected == nil {
		return "No strategy selected"
	}

	availableExchanges := m.getAvailableExchangesForStrategy()

	var b strings.Builder

	// Title
	title := TitleStyle.Render("SELECT EXCHANGE")
	subtitle := SubtitleStyle.Render(fmt.Sprintf("Choose exchange for %s strategy", m.selected.Name))

	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, title))
	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, subtitle))
	b.WriteString("\n\n")

	// Exchange list
	for i, exConfig := range availableExchanges {
		cursor := "  "
		if m.exchangeCursor == i {
			cursor = "▶ "
		}

		// Build exchange item
		name := exConfig.Name
		if m.exchangeCursor == i {
			name = StrategyNameSelectedStyle.Render(name)
		} else {
			name = StrategyNameStyle.Render(name)
		}

		// Get assets for this exchange from strategy config
		assets := m.selected.Config.Assets[exConfig.Name]
		assetInfo := StrategyMetaStyle.Render(fmt.Sprintf("Assets: %s", strings.Join(assets, ", ")))

		networkInfo := ""
		if exConfig.Network != "" {
			networkInfo = StrategyMetaStyle.Render(fmt.Sprintf("Network: %s", exConfig.Network))
		}

		itemContent := fmt.Sprintf("%s\n%s", name, assetInfo)
		if networkInfo != "" {
			itemContent += "\n" + networkInfo
		}

		var item string
		if m.exchangeCursor == i {
			item = StrategyItemSelectedStyle.Render(cursor + itemContent)
		} else {
			item = StrategyItemStyle.Render(cursor + itemContent)
		}

		b.WriteString(lipgloss.Place(m.width, lipgloss.Height(item), lipgloss.Center, lipgloss.Top, item))
		b.WriteString("\n")
	}

	// Help text
	help := HelpStyle.Render("↑↓/jk Navigate  ↵ Select  esc Back  ctrl+c Quit")
	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, help))

	return b.String()
}

func (m SelectionModel) renderCredentials() string {
	if m.selected == nil || m.selectedExchange == nil {
		return "No exchange selected"
	}

	var b strings.Builder

	// Title
	title := TitleStyle.Render(fmt.Sprintf("🔐 %s CREDENTIALS", strings.ToUpper(m.selectedExchange.Name)))
	subtitle := SubtitleStyle.Render("Enter your API credentials")

	b.WriteString("\n\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, title))
	b.WriteString("\n")
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, subtitle))
	b.WriteString("\n\n")

	// Render each credential field
	for i, field := range m.credentialFields {
		// Format field name for display
		displayName := strings.ReplaceAll(field, "_", " ")
		displayName = strings.Title(displayName)

		label := ConfirmFieldStyle.Render(displayName + ":")

		// Get current input value
		inputValue := m.fieldInputs[field]

		// Mask sensitive fields
		if strings.Contains(field, "key") || strings.Contains(field, "secret") {
			if len(inputValue) > 0 {
				inputValue = strings.Repeat("*", len(inputValue))
			}
		}

		// Highlight current field
		if i == m.currentField {
			inputValue = ConfirmValueStyle.Render(inputValue + "█") // Cursor
		} else {
			inputValue = StrategyMetaStyle.Render(inputValue)
		}

		b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top, label+" "+inputValue))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	help := HelpStyle.Render("↑↓/tab Navigate  ↵ Next/Continue  esc Back  ctrl+c Quit")
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

	// Selected exchange and assets
	if m.selectedExchange != nil {
		details = append(details, fmt.Sprintf("%s  %s",
			ConfirmFieldStyle.Render("Exchange:"),
			ConfirmValueStyle.Render(m.selectedExchange.Name),
		))

		if m.selectedExchange.Network != "" {
			details = append(details, fmt.Sprintf("%s  %s",
				ConfirmFieldStyle.Render("Network:"),
				ConfirmValueStyle.Render(m.selectedExchange.Network),
			))
		}

		// Get assets from strategy config
		assets := m.selected.Config.Assets[m.selectedExchange.Name]
		details = append(details, fmt.Sprintf("%s  %s",
			ConfirmFieldStyle.Render("Assets:"),
			ConfirmValueStyle.Render(strings.Join(assets, ", ")),
		))
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

	// Confirmation input prompt
	b.WriteString("\n")
	b.WriteString(ConfirmFieldStyle.Render("Type 'CONFIRM' to proceed: "))
	b.WriteString(ConfirmValueStyle.Render(m.confirmInput))
	b.WriteString("\n")

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

	// Build the actual command that will be executed
	nextSteps := []string{
		"",
		"Command to be executed:",
		"",
	}

	if m.selectedExchange != nil {
		// Get the .so file path
		strategyName := filepath.Base(m.selected.Path)
		soPath := filepath.Join(m.selected.Path, strategyName+".so")

		// Build command based on exchange type
		if m.selectedExchange.Name == "paradex" {
			nextSteps = append(nextSteps,
				SubtitleStyle.Render("  kronos-live run \\"),
				SubtitleStyle.Render(fmt.Sprintf("    --exchange %s \\", m.selectedExchange.Name)),
				SubtitleStyle.Render(fmt.Sprintf("    --strategy %s \\", soPath)),
			)

			// Add Paradex-specific flags
			if accountAddr, ok := m.selectedExchange.Credentials["account_address"]; ok && accountAddr != "" {
				nextSteps = append(nextSteps, SubtitleStyle.Render(fmt.Sprintf("    --paradex-account-address %s \\", accountAddr)))
			}
			if ethKey, ok := m.selectedExchange.Credentials["eth_private_key"]; ok && ethKey != "" {
				masked := ethKey
				if len(masked) > 10 {
					masked = masked[:6] + "..." + masked[len(masked)-4:]
				}
				nextSteps = append(nextSteps, SubtitleStyle.Render(fmt.Sprintf("    --paradex-eth-private-key %s \\", masked)))
			}
			if l2Key, ok := m.selectedExchange.Credentials["l2_private_key"]; ok && l2Key != "" {
				masked := l2Key
				if len(masked) > 10 {
					masked = masked[:6] + "..." + masked[len(masked)-4:]
				}
				nextSteps = append(nextSteps, SubtitleStyle.Render(fmt.Sprintf("    --paradex-l2-private-key %s \\", masked)))
			}
			if m.selectedExchange.Network != "" {
				nextSteps = append(nextSteps, SubtitleStyle.Render(fmt.Sprintf("    --paradex-network %s", m.selectedExchange.Network)))
			}
		} else {
			// Generic exchange command
			nextSteps = append(nextSteps,
				SubtitleStyle.Render("  kronos-live run \\"),
				SubtitleStyle.Render(fmt.Sprintf("    --exchange %s \\", m.selectedExchange.Name)),
				SubtitleStyle.Render(fmt.Sprintf("    --strategy %s", soPath)),
			)
		}
	}

	nextSteps = append(nextSteps,
		"",
		"",
		StrategyMetaStyle.Render("Press Enter to start live trading, or Q to quit"),
		"",
	)

	b.WriteString(strings.Join(nextSteps, "\n"))

	help := HelpStyle.Render("↵ Start Live Trading  q Quit")
	b.WriteString("\n")
	b.WriteString(help)

	boxed := BoxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxed)
}

// RunSelectionTUI runs the strategy selection TUI
func RunSelectionTUI() error {
	// Try to discover strategies from ./strategies directory
	strategies, err := DiscoverStrategies()
	if err != nil {
		// Fall back to mock strategies if discovery fails
		strategies = GetMockStrategies()
	}

	// Load global exchanges config
	exchangesConfigPath := "./exchanges.yml"
	globalExchanges, err := LoadGlobalExchangesConfig(exchangesConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load global exchanges config: %w", err)
	}

	// Create model
	m := NewSelectionModel(strategies, globalExchanges)

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

		// Save credentials back to global exchanges.yml if a strategy was successfully configured
		if model.selected != nil && model.selectedExchange != nil && model.currentScreen == ScreenSuccess {
			if err := SaveGlobalExchangesConfig(exchangesConfigPath, model.globalExchanges); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			// Execute live trading
			fmt.Println("\n🚀 Starting live trading...\n")
			if err := ExecuteLiveTrading(model.selected, model.selectedExchange); err != nil {
				return fmt.Errorf("failed to execute live trading: %w", err)
			}
		}
	}

	return nil
}
