package settings

import (
	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/config/settings/connectors"
	"github.com/backtesting-org/kronos-cli/internal/router"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/fx"
)

var Module = fx.Module("settings",
	fx.Provide(
		// TUI Views (factories)
		NewSettingsListViewFactory,
		NewConnectorFormViewFactory,
		NewDeleteConfirmViewFactory,
	),
)

// Factory types for DI
type SettingsListViewFactory func() tea.Model
type ConnectorFormViewFactory func(connectorName string, isEdit bool) tea.Model
type DeleteConfirmViewFactory func(connectorName string) tea.Model

// NewSettingsListViewFactory creates a factory function for the settings list view
func NewSettingsListViewFactory(
	config settings.Configuration,
	connectorSvc connectors.ConnectorService,
	r router.Router,
	formFactory ConnectorFormViewFactory,
	deleteFactory DeleteConfirmViewFactory,
) SettingsListViewFactory {
	return func() tea.Model {
		return NewSettingsListView(config, connectorSvc, r, formFactory, deleteFactory)
	}
}

// NewConnectorFormViewFactory creates a factory function for the connector form view
func NewConnectorFormViewFactory(
	config settings.Configuration,
	connectorSvc connectors.ConnectorService,
	r router.Router,
	deleteFactory DeleteConfirmViewFactory,
) ConnectorFormViewFactory {
	return func(connectorName string, isEdit bool) tea.Model {
		return NewConnectorFormView(config, connectorSvc, r, deleteFactory, connectorName, isEdit)
	}
}

// NewDeleteConfirmViewFactory creates a factory function for the delete confirmation view
func NewDeleteConfirmViewFactory(
	config settings.Configuration,
	r router.Router,
) DeleteConfirmViewFactory {
	return func(connectorName string) tea.Model {
		return NewDeleteConfirmView(config, r, connectorName)
	}
}
