package router

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

// Route represents a navigation destination
type Route string

const (
	RouteInit             Route = "init"
	RouteLive             Route = "live"
	RouteMonitor          Route = "monitor"
	RouteBacktest         Route = "backtest"
	RouteAnalyze          Route = "analyze"
	RouteMenu             Route = "menu"
	RouteStrategyList     Route = "strategy-list"
	RouteStrategyDetail   Route = "strategy-detail"
	RouteStrategyCompile  Route = "strategy-compile"
	RouteStrategyBacktest Route = "strategy-backtest"
	RouteStrategyEdit     Route = "strategy-edit"
	RouteStrategyDelete   Route = "strategy-delete"
)

// Router is the main Tea model that manages view navigation using Bubblon's Controller
// It delegates all navigation through Bubblon's Open/Close/Replace commands
type Router interface {
	tea.Model

	// SetInitialView sets the starting view
	SetInitialView(view tea.Model)
}

type router struct {
	controller bubblon.Controller
}

// NewRouter creates an empty router ready to have views pushed
func NewRouter() (Router, error) {
	// Create an empty controller with a dummy model initially
	// This will be replaced when SetInitialView is called
	dummyModel := &dummyModel{}
	ctrl, err := bubblon.New(dummyModel)
	if err != nil {
		return nil, err
	}
	return &router{
		controller: ctrl,
	}, nil
}

// dummyModel is a placeholder for when router hasn't been initialized with a real view yet
type dummyModel struct{}

func (d *dummyModel) Init() tea.Cmd                           { return nil }
func (d *dummyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return d, nil }
func (d *dummyModel) View() string                            { return "" }

// SetInitialView sets the starting view (call before running the program)
func (r *router) SetInitialView(view tea.Model) {
	r.controller, _ = bubblon.New(view)
}

// Init implements tea.Model
func (r *router) Init() tea.Cmd {
	return r.controller.Init()
}

// Update implements tea.Model
// Bubblon handles all navigation commands (Open/Close/Replace) internally
func (r *router) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := r.controller.Update(msg)
	r.controller = updated.(bubblon.Controller)
	return r, cmd
}

// View implements tea.Model
func (r *router) View() string {
	return r.controller.View()
}
