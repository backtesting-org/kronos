package settings

import (
	"fmt"
	"strings"

	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/config/settings/connectors"
	"github.com/backtesting-org/kronos-cli/internal/router"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
)

// ConnectorFormModel represents the connector detail/edit view
type ConnectorFormModel struct {
	form          *huh.Form
	connector     settings.Connector
	config        settings.Configuration
	connectorSvc  connectors.ConnectorService
	router        router.Router
	deleteFactory DeleteConfirmViewFactory
	isEditMode    bool
	originalName  string
	err           error

	// UI state
	showingDetail bool // true = show detail view, false = show edit form

	// Form field values
	exchangeName string
	network      string
	enabled      bool
	credentials  map[string]string
	assets       []string
}

// NewConnectorFormView creates a new connector form view with Huh forms
func NewConnectorFormView(
	config settings.Configuration,
	connectorSvc connectors.ConnectorService,
	r router.Router,
	deleteFactory DeleteConfirmViewFactory,
	connectorName string,
	isEdit bool,
) tea.Model {
	m := &ConnectorFormModel{
		config:        config,
		connectorSvc:  connectorSvc,
		router:        r,
		deleteFactory: deleteFactory,
		isEditMode:    isEdit,
		originalName:  connectorName,
		credentials:   make(map[string]string),
		enabled:       true,
	}

	if isEdit && connectorName != "" {
		// Load existing connector
		connectorList, err := config.GetConnectors()
		if err != nil {
			m.err = err
			return m
		}

		for _, conn := range connectorList {
			if conn.Name == connectorName {
				m.connector = conn
				m.exchangeName = conn.Name
				m.network = conn.Network
				m.enabled = conn.Enabled
				m.assets = conn.Assets
				m.credentials = conn.Credentials
				break
			}
		}

		if m.connector.Name == "" {
			m.err = fmt.Errorf("connector '%s' not found", connectorName)
			return m
		}

		// Show detail view first for editing
		m.showingDetail = true
	} else {
		// Adding new - go straight to form
		m.showingDetail = false
		m.form = m.buildForm()
	}

	return m
}

// buildForm creates the Huh form focused on credentials
func (m *ConnectorFormModel) buildForm() *huh.Form {
	var groups []*huh.Group

	// Header
	var headerFields []huh.Field
	if m.isEditMode {
		headerFields = append(headerFields,
			huh.NewNote().
				Title("✏️  "+m.exchangeName).
				Description("Update API credentials"),
		)
	} else if m.exchangeName != "" {
		headerFields = append(headerFields,
			huh.NewNote().
				Title("➕ "+m.exchangeName).
				Description("Configure API credentials for this connector"),
		)
	} else {
		// No pre-selection - show exchange dropdown
		availableExchanges := m.connectorSvc.GetAvailableConnectorNames()
		exchangeOptions := make([]huh.Option[string], len(availableExchanges))
		for i, ex := range availableExchanges {
			exchangeOptions[i] = huh.NewOption(ex, ex)
		}

		headerFields = append(headerFields,
			huh.NewSelect[string]().
				Title("Select Exchange").
				Options(exchangeOptions...).
				Value(&m.exchangeName),
		)
	}

	groups = append(groups, huh.NewGroup(headerFields...))

	// Get required credential fields from SDK (e.g., hyperliquid needs "private_key" and "account_address")
	requiredFields := m.connectorSvc.GetRequiredCredentialFields(m.exchangeName)
	if len(requiredFields) == 0 {
		// Fallback to common fields if SDK doesn't provide
		requiredFields = []string{"api_key", "api_secret"}
	}

	// Build credential input fields dynamically based on what SDK requires
	var credFields []huh.Field

	for _, fieldName := range requiredFields {
		// Initialize field value
		fieldValue := ""
		fieldDesc := fmt.Sprintf("Enter your %s", formatFieldName(fieldName))

		// If editing, show current value masked
		if m.isEditMode && len(m.credentials) > 0 {
			if existing, exists := m.credentials[fieldName]; exists && len(existing) > 3 {
				masked := existing[:3] + strings.Repeat("•", minInt(len(existing)-3, 20))
				fieldDesc = fmt.Sprintf("Current: %s", masked)
				fieldValue = existing
			}
		}

		// Determine echo mode (mask secrets/keys, show addresses plainly)
		echoMode := huh.EchoModeNormal
		if strings.Contains(strings.ToLower(fieldName), "key") ||
			strings.Contains(strings.ToLower(fieldName), "secret") {
			echoMode = huh.EchoModePassword
		}

		credFields = append(credFields,
			huh.NewInput().
				Title(formatFieldName(fieldName)).
				Description(fieldDesc).
				Placeholder("...").
				EchoMode(echoMode).
				Value(&fieldValue).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("%s is required", fieldName)
					}
					return nil
				}),
		)

		// Store the pointer so we can access it after completion
		m.credentials[fieldName] = fieldValue
	}

	// Only show enable toggle if editing (less prominent)
	if m.isEditMode {
		credFields = append(credFields,
			huh.NewConfirm().
				Title("Enabled?").
				Value(&m.enabled),
		)
	}

	groups = append(groups, huh.NewGroup(credFields...))

	// Set defaults
	if m.network == "" {
		m.network = "mainnet" // Default, but we don't ask about it
	}
	if !m.isEditMode {
		m.enabled = true // Default to enabled for new connectors
	}

	return huh.NewForm(groups...).
		WithTheme(huh.ThemeCharm()).
		WithShowHelp(true).
		WithShowErrors(true)
}

// minInt returns the minimum of two ints
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// formatFieldName converts snake_case to Title Case for display
func formatFieldName(field string) string {
	// Replace underscores with spaces and capitalize each word
	parts := strings.Split(field, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func (m *ConnectorFormModel) Init() tea.Cmd {
	if m.form != nil {
		return m.form.Init()
	}
	return nil
}

func (m *ConnectorFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle error state
	if m.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "q" || msg.String() == "esc" {
				return m, m.router.Back()
			}
		}
		return m, nil
	}

	// Handle Ctrl+C to quit
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	// If showing detail view
	if m.showingDetail {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				return m, m.router.Back()
			case "e":
				// Switch to edit mode
				m.showingDetail = false
				m.form = m.buildForm()
				return m, m.form.Init()
			case " ":
				// Quick toggle enabled
				m.connector.Enabled = !m.connector.Enabled
				if err := m.config.UpdateConnector(m.connector); err != nil {
					m.err = err
					return m, nil
				}
				// Stay on detail view to see the change
				return m, nil
			case "d":
				// Show delete confirmation dialog
				deleteView := m.deleteFactory(m.connector.Name)
				return m, bubblon.Open(deleteView)
			}
		}
		return m, nil
	}

	// Editing mode - handle form
	var cmd tea.Cmd
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form is complete
	if m.form.State == huh.StateCompleted {
		// Build connector from form values
		m.connector = settings.Connector{
			Name:        m.exchangeName,
			Network:     m.network,
			Enabled:     m.enabled,
			Assets:      m.assets,
			Credentials: m.credentials,
		}

		// Save the connector
		if err := m.saveConnector(); err != nil {
			m.err = err
			return m, nil
		}

		// Success - go back to list
		return m, m.router.Back()
	}

	// Check if form was aborted (Esc pressed)
	if m.form.State == huh.StateAborted {
		if m.isEditMode {
			// Go back to detail view
			m.showingDetail = true
			return m, nil
		}
		// Adding new - go back to list
		return m, m.router.Back()
	}

	return m, cmd
}

func (m *ConnectorFormModel) saveConnector() error {
	if m.isEditMode {
		return m.config.UpdateConnector(m.connector)
	}
	return m.config.AddConnector(m.connector)
}

func (m *ConnectorFormModel) View() string {
	if m.err != nil {
		errorBox := lipgloss.NewStyle().
			Foreground(ui.ColorDanger).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ui.ColorDanger).
			Padding(1, 2).
			Render("❌ " + m.err.Error() + "\n\nPress 'q' to go back")
		return errorBox
	}

	// Show detail view or edit form
	if m.showingDetail {
		return m.renderDetailView()
	}

	if m.form == nil {
		return "Loading..."
	}

	return m.form.View()
}

// renderDetailView shows a beautiful detail card for the connector
func (m *ConnectorFormModel) renderDetailView() string {
	var content strings.Builder

	// Title
	title := ui.TitleStyle.Render("⚙️  " + m.connector.Name)
	content.WriteString(title)
	content.WriteString("\n\n")

	// Status badge
	var statusBadge string
	if m.connector.Enabled {
		statusBadge = ui.StatusReadyStyle.Render("● ENABLED")
	} else {
		statusBadge = lipgloss.NewStyle().
			Foreground(ui.ColorMuted).
			Bold(true).
			Render("○ DISABLED")
	}
	content.WriteString(statusBadge)
	content.WriteString("\n\n")

	// Detail box
	detailStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(1, 2).
		Width(68)

	var details strings.Builder

	// Exchange type
	labelStyle := lipgloss.NewStyle().
		Foreground(ui.ColorMuted).
		Width(15)
	valueStyle := lipgloss.NewStyle().
		Foreground(ui.ColorPrimary).
		Bold(true)

	details.WriteString(labelStyle.Render("Exchange:"))
	details.WriteString(" ")
	details.WriteString(valueStyle.Render(m.connector.Name))
	details.WriteString("\n\n")

	// Network
	details.WriteString(labelStyle.Render("Network:"))
	details.WriteString(" ")
	networkValue := m.connector.Network
	if networkValue == "" {
		networkValue = "mainnet"
	}
	networkStyle := valueStyle
	if networkValue == "testnet" {
		networkStyle = lipgloss.NewStyle().
			Foreground(ui.ColorWarning).
			Bold(true)
	}
	details.WriteString(networkStyle.Render(networkValue))
	details.WriteString("\n\n")

	// Credentials section
	credHeaderStyle := lipgloss.NewStyle().
		Foreground(ui.ColorSecondary).
		Bold(true)
	details.WriteString(credHeaderStyle.Render("Credentials"))
	details.WriteString("\n\n")

	// Get required fields from SDK for this exchange
	requiredFields := m.connectorSvc.GetRequiredCredentialFields(m.connector.Name)
	if len(requiredFields) == 0 {
		requiredFields = []string{"api_key", "api_secret"}
	}

	// Show each credential field dynamically
	for _, fieldName := range requiredFields {
		fieldLabel := formatFieldName(fieldName) + ":"
		details.WriteString(labelStyle.Render(fieldLabel))
		details.WriteString(" ")

		if value, exists := m.connector.Credentials[fieldName]; exists && len(value) > 3 {
			// Mask private keys, show addresses plainly
			if strings.Contains(strings.ToLower(fieldName), "key") ||
				strings.Contains(strings.ToLower(fieldName), "secret") {
				masked := value[:3] + strings.Repeat("•", minInt(len(value)-3, 20))
				details.WriteString(lipgloss.NewStyle().
					Foreground(ui.ColorSuccess).
					Render(masked))
			} else {
				// Show addresses/usernames plainly
				details.WriteString(lipgloss.NewStyle().
					Foreground(ui.ColorSuccess).
					Render(value))
			}
		} else {
			details.WriteString(lipgloss.NewStyle().
				Foreground(ui.ColorDanger).
				Render("Not set"))
		}
		details.WriteString("\n")
	}

	content.WriteString(detailStyle.Render(details.String()))
	content.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(ui.ColorMuted)
	keyStyle := lipgloss.NewStyle().
		Foreground(ui.ColorPrimary).
		Bold(true)

	help := fmt.Sprintf(
		"%s Edit  %s Toggle  %s Delete  %s Back",
		keyStyle.Render("e"),
		keyStyle.Render("Space"),
		keyStyle.Render("d"),
		keyStyle.Render("q"),
	)
	content.WriteString(helpStyle.Render(help))

	return content.String()
}
