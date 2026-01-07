package monitor

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/monitoring"
	tea "github.com/charmbracelet/bubbletea"
)

// MonitorViewFactory creates a new monitor view
type MonitorViewFactory func() tea.Model

// NewMonitorViewFactory creates the factory for monitor views
func NewMonitorViewFactory(querier monitoring.ViewQuerier) MonitorViewFactory {
	return func() tea.Model {
		return NewInstanceListModel(querier)
	}
}
