package monitor

import (
	"fmt"
	"strings"

	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/monitor/tabs"
	"github.com/backtesting-org/kronos-cli/internal/ui"
	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	tea "github.com/charmbracelet/bubbletea"
)

// Tab represents a detail view tab
type Tab int

const (
	TabOverview Tab = iota
	TabPositions
	TabOrderbook
	TabTrades
	TabPnL
)

var tabNames = []string{"Overview", "Positions", "Orderbook", "Trades", "PnL"}

// instanceDetailModel shows detailed view of a single instance
type instanceDetailModel struct {
	ui.BaseModel // Embed for common key handling
	querier      monitoring.ViewQuerier
	instanceID   string
	activeTab    Tab
	width        int
	height       int

	// Tab models - each manages its own data
	overviewTab  *tabs.OverviewModel
	positionsTab *tabs.PositionsModel
	orderbookTab *tabs.OrderbookModel
	tradesTab    *tabs.TradesModel
	pnlTab       *tabs.PnLModel
}

// NewInstanceDetailModel creates a detail view for an instance
func NewInstanceDetailModel(querier monitoring.ViewQuerier, instanceID string) tea.Model {
	return &instanceDetailModel{
		BaseModel:    ui.BaseModel{IsRoot: false},
		querier:      querier,
		instanceID:   instanceID,
		activeTab:    TabOverview,
		overviewTab:  tabs.NewOverviewModel(querier, instanceID),
		positionsTab: tabs.NewPositionsModel(querier, instanceID),
		orderbookTab: tabs.NewOrderbookModel(querier, instanceID),
		tradesTab:    tabs.NewTradesModel(querier, instanceID),
		pnlTab:       tabs.NewPnLModel(querier, instanceID),
	}
}

func (m *instanceDetailModel) Init() tea.Cmd {
	// Initialize all tabs
	return tea.Batch(
		m.overviewTab.Init(),
		m.positionsTab.Init(),
		m.orderbookTab.Init(),
		m.tradesTab.Init(),
		m.pnlTab.Init(),
	)
}

func (m *instanceDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if handled, cmd := m.BaseModel.HandleCommonKeys(msg); handled {
			return m, cmd
		}

		switch msg.String() {
		case "left", "h":
			if m.activeTab > 0 {
				m.activeTab--
			}
			return m, nil

		case "right", "l":
			if m.activeTab < TabPnL {
				m.activeTab++
			}
			return m, nil

		case "1":
			m.activeTab = TabOverview
			return m, nil
		case "2":
			m.activeTab = TabPositions
			return m, nil
		case "3":
			m.activeTab = TabOrderbook
			return m, nil
		case "4":
			m.activeTab = TabTrades
			return m, nil
		case "5":
			m.activeTab = TabPnL
			return m, nil
		}

		// Forward key messages only to active tab
		var cmd tea.Cmd
		switch m.activeTab {
		case TabOverview:
			_, cmd = m.overviewTab.Update(msg)
		case TabPositions:
			_, cmd = m.positionsTab.Update(msg)
		case TabOrderbook:
			_, cmd = m.orderbookTab.Update(msg)
		case TabTrades:
			_, cmd = m.tradesTab.Update(msg)
		case TabPnL:
			_, cmd = m.pnlTab.Update(msg)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	// Forward ALL other messages (ticks, data, etc.) to ALL tabs
	// This ensures each tab's tick cycle continues even when not active
	var cmd tea.Cmd

	_, cmd = m.overviewTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	_, cmd = m.positionsTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	_, cmd = m.orderbookTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	_, cmd = m.tradesTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	_, cmd = m.pnlTab.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}

func (m *instanceDetailModel) View() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	// Content from active tab
	switch m.activeTab {
	case TabOverview:
		b.WriteString(m.overviewTab.View())
	case TabPositions:
		b.WriteString(m.positionsTab.View())
	case TabOrderbook:
		b.WriteString(m.orderbookTab.View())
	case TabTrades:
		b.WriteString(m.tradesTab.View())
	case TabPnL:
		b.WriteString(m.pnlTab.View())
	}

	// Help
	b.WriteString("\n\n")
	helpText := "[←→] Switch Tab • [1-5] Jump to Tab • [R] Refresh • [Q] Back"
	if m.activeTab == TabOrderbook {
		helpText = "[←→] Switch Tab • [D] Toggle Depth • [R] Refresh • [Q] Back"
	}
	b.WriteString(ui.HelpStyle.Render(helpText))

	return b.String()
}

func (m *instanceDetailModel) renderHeader() string {
	title := ui.TitleStyle.Render(strings.ToUpper(m.instanceID))
	return title
}

func (m *instanceDetailModel) renderTabs() string {
	var renderedTabs []string
	for i, name := range tabNames {
		if Tab(i) == m.activeTab {
			renderedTabs = append(renderedTabs, TabActiveStyle.Render(fmt.Sprintf("[%s]", name)))
		} else {
			renderedTabs = append(renderedTabs, TabStyle.Render(name))
		}
	}
	return strings.Join(renderedTabs, "  ")
}
