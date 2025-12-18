package settings

import (
	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/router"
	tea "github.com/charmbracelet/bubbletea"
)

// DeleteConfirmModel represents the delete confirmation view
type DeleteConfirmModel struct {
	connectorName string
	config        settings.Configuration
	router        router.Router
	confirmed     bool
}

// NewDeleteConfirmView creates a new delete confirmation view
func NewDeleteConfirmView(
	config settings.Configuration,
	r router.Router,
	connectorName string,
) tea.Model {
	return &DeleteConfirmModel{
		config:        config,
		router:        r,
		connectorName: connectorName,
	}
}

func (m *DeleteConfirmModel) Init() tea.Cmd {
	return nil
}

func (m *DeleteConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "n":
			// Cancel - go back
			return m, m.router.Back()
		case "y", "enter":
			// Confirm delete
			if err := m.config.RemoveConnector(m.connectorName); err != nil {
				// TODO: Show error
				return m, m.router.Back()
			}
			return m, m.router.Back()
		}
	}
	return m, nil
}

func (m *DeleteConfirmModel) View() string {
	s := "🗑️  Delete Connector\n"
	s += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"
	s += "Are you sure you want to delete connector '" + m.connectorName + "'?\n\n"
	s += "This action cannot be undone.\n\n"
	s += "Press 'y' to confirm, 'n' or 'Esc' to cancel.\n"
	return s
}
