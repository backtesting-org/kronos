package monitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donderom/bubblon"
)

// InstanceInfo holds display data for a running instance
type InstanceInfo struct {
	ID       string
	Status   string // running, warning, stopped, unknown
	PID      int
	Uptime   time.Duration
	PnL24h   float64
	Health   int // 0-5
	HasError bool
}

// instanceListModel displays all running strategy instances
type instanceListModel struct {
	ui.BaseModel // Embed for common key handling
	querier      monitoring.ViewQuerier
	instances    []InstanceInfo
	cursor       int
	loading      bool
	err          error
	width        int
	height       int
}

// NewInstanceListModel creates a new instance list view
func NewInstanceListModel(querier monitoring.ViewQuerier) tea.Model {
	return &instanceListModel{
		BaseModel: ui.BaseModel{IsRoot: true}, // This is launched from main menu
		querier:   querier,
		loading:   true,
	}
}

// Messages
type instancesLoadedMsg struct {
	instances []InstanceInfo
	err       error
}

type tickMsg time.Time

func (m *instanceListModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadInstances(),
		m.tickCmd(),
	)
}

func (m *instanceListModel) tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *instanceListModel) loadInstances() tea.Cmd {
	return func() tea.Msg {
		instanceIDs, err := m.querier.ListInstances()
		if err != nil {
			return instancesLoadedMsg{err: err}
		}

		var instances []InstanceInfo
		for _, id := range instanceIDs {
			info := InstanceInfo{
				ID:     id,
				Status: "unknown",
			}

			// Try to get metrics
			metrics, err := m.querier.QueryMetrics(id)
			if err == nil && metrics != nil {
				info.Status = metrics.Status
			}

			// Try to get PnL
			pnl, err := m.querier.QueryPnL(id)
			if err == nil && pnl != nil {
				info.PnL24h, _ = pnl.TotalPnL.Float64()
			}

			// Health check
			if err := m.querier.HealthCheck(id); err != nil {
				info.Health = 0
				info.HasError = true
				if info.Status == "unknown" {
					info.Status = "stopped"
				}
			} else {
				info.Health = 5
				if info.Status == "unknown" {
					info.Status = "running"
				}
			}

			instances = append(instances, info)
		}

		return instancesLoadedMsg{instances: instances}
	}
}

func (m *instanceListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case instancesLoadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.instances = msg.instances
		}
		return m, nil

	case tickMsg:
		return m, tea.Batch(
			m.loadInstances(),
			m.tickCmd(),
		)

	case tea.KeyMsg:
		// Handle common keys first (ctrl+c, q, esc)
		if handled, cmd := m.BaseModel.HandleCommonKeys(msg); handled {
			return m, cmd
		}

		switch msg.String() {

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.instances)-1 {
				m.cursor++
			}
			return m, nil

		case "enter":
			if len(m.instances) > 0 {
				selected := m.instances[m.cursor]
				detailView := NewInstanceDetailModel(m.querier, selected.ID)
				return m, bubblon.Open(detailView)
			}
			return m, nil

		case "r":
			m.loading = true
			return m, m.loadInstances()
		}
	}

	return m, nil
}

func (m *instanceListModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(ui.TitleStyle.Render("MONITOR"))
	b.WriteString("\n")

	if m.loading {
		b.WriteString(ui.SubtitleStyle.Render("Loading instances..."))
		b.WriteString("\n")
	} else if m.err != nil {
		b.WriteString(ui.StatusErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
	} else if len(m.instances) == 0 {
		b.WriteString(m.renderEmpty())
	} else {
		b.WriteString(m.renderTable())
	}

	// Help
	b.WriteString("\n")
	b.WriteString(ui.HelpStyle.Render("[↑↓] Navigate • [Enter] View Details • [R] Refresh • [Q] Back"))

	return b.String()
}

func (m *instanceListModel) renderEmpty() string {
	box := ui.BoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			"",
			ui.SubtitleStyle.Render("No Running Strategies"),
			"",
			ui.HelpStyle.Render("Start a strategy to begin monitoring"),
			"",
		),
	)
	return "\n" + box + "\n"
}

func (m *instanceListModel) renderTable() string {
	var b strings.Builder

	// Table header
	header := fmt.Sprintf("  %-8s %-18s %-8s %-10s %-14s %-10s",
		"STATUS", "STRATEGY", "PID", "UPTIME", "PNL", "HEALTH")
	b.WriteString(TableHeaderStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("─", 72)))
	b.WriteString("\n")

	// Table rows
	for i, inst := range m.instances {
		row := m.renderInstanceRow(inst, i == m.cursor)
		b.WriteString(row)
		b.WriteString("\n")
	}

	return b.String()
}

func (m *instanceListModel) renderInstanceRow(inst InstanceInfo, selected bool) string {
	icon := GetStatusIcon(inst.Status)

	statusLen := len(inst.Status)
	if statusLen > 3 {
		statusLen = 3
	}
	statusText := GetStatusStyle(inst.Status).Render(strings.ToUpper(inst.Status[:statusLen]))

	pid := "-"
	if inst.PID > 0 {
		pid = fmt.Sprintf("%d", inst.PID)
	}

	uptime := "-"
	if inst.Uptime > 0 {
		uptime = formatDuration(inst.Uptime)
	}

	pnl := FormatPnL(inst.PnL24h)
	health := FormatHealthBar(inst.Health)

	row := fmt.Sprintf("  %s %-4s %-18s %-8s %-10s %-14s %s",
		icon, statusText, inst.ID, pid, uptime, pnl, health)

	if selected {
		return TableRowSelectedStyle.Render(row)
	}
	return TableRowStyle.Render(row)
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}
