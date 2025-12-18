package settings

import (
	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/config/settings/connectors"
	"github.com/backtesting-org/kronos-cli/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// ConnectorFormModel represents the create/edit connector form
type ConnectorFormModel struct {
	step               int
	connector          settings.Connector
	availableExchanges []string
	selectedExchange   string

	// Field editing
	focusedField int
	fieldValues  map[string]string
	fieldErrors  map[string]string

	config       settings.Configuration
	connectorSvc connectors.ConnectorService
	router       router.Router

	isEditMode   bool
	originalName string
}

// NewConnectorFormView creates a new connector form view
func NewConnectorFormView(
	config settings.Configuration,
	connectorSvc connectors.ConnectorService,
	r router.Router,
	connectorName string,
	isEdit bool,
) tea.Model {
	// Fetch available exchanges from SDK via connector service
	availableNames := connectorSvc.GetAvailableConnectorNames()

	return &ConnectorFormModel{
		config:             config,
		connectorSvc:       connectorSvc,
		router:             r,
		availableExchanges: availableNames,
		fieldValues:        make(map[string]string),
		fieldErrors:        make(map[string]string),
		isEditMode:         isEdit,
		originalName:       connectorName,
	}
}

func (m *ConnectorFormModel) Init() tea.Cmd {
	return nil
}

func (m *ConnectorFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			// Go back to list
			return m, m.router.Back()
		}
	}
	return m, nil
}

func (m *ConnectorFormModel) View() string {
	s := "📝 Connector Form (Placeholder)\n"
	s += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"
	s += "This view will be implemented in Sprint 3.\n\n"
	s += "Press 'q' or 'Esc' to go back.\n"
	return s
}
