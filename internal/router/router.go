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
	RouteSettingsList     Route = "settings-list"
	RouteSettingsEdit     Route = "settings-edit"
	RouteSettingsCreate   Route = "settings-create"
	RouteSettingsDelete   Route = "settings-delete"
)

// ViewFactory creates a view for a given route
type ViewFactory func() tea.Model

// Router manages navigation between views using path-based routing
type Router interface {
	tea.Model

	// RegisterRoute associates a path with a view factory
	RegisterRoute(route Route, factory ViewFactory)

	// NavigateTo pushes a new view onto the stack by route name
	NavigateTo(route Route) tea.Cmd

	// Back pops the current view off the stack
	Back() tea.Cmd

	// SetInitialView sets the starting view
	SetInitialView(view tea.Model)
}

type router struct {
	controller bubblon.Controller
	routes     map[Route]ViewFactory
}

// NewRouter creates a router with route registration
func NewRouter() (Router, error) {
	dummyModel := &dummyModel{}
	ctrl, err := bubblon.New(dummyModel)
	if err != nil {
		return nil, err
	}
	return &router{
		controller: ctrl,
		routes:     make(map[Route]ViewFactory),
	}, nil
}

// dummyModel is a placeholder for when router hasn't been initialized with a real view yet
type dummyModel struct{}

func (d *dummyModel) Init() tea.Cmd                           { return nil }
func (d *dummyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return d, nil }
func (d *dummyModel) View() string                            { return "" }

// RegisterRoute associates a route path with a view factory
func (r *router) RegisterRoute(route Route, factory ViewFactory) {
	r.routes[route] = factory
}

// NavigateTo creates a view from the route factory and pushes it onto the stack
func (r *router) NavigateTo(route Route) tea.Cmd {
	factory, exists := r.routes[route]
	if !exists {
		return nil // Route not found, do nothing
	}
	view := factory()
	return bubblon.Open(view)
}

// Back pops the current view off the stack
func (r *router) Back() tea.Cmd {
	return bubblon.Cmd(bubblon.Close())
}

// SetInitialView sets the starting view (call before running the program)
func (r *router) SetInitialView(view tea.Model) {
	r.controller, _ = bubblon.New(view)
}

// Init implements tea.Model
func (r *router) Init() tea.Cmd {
	return r.controller.Init()
}

// Update implements tea.Model
// Handles navigation messages and delegates to bubblon controller
func (r *router) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle navigation messages
	switch msg := msg.(type) {
	case NavigateMsg:
		// Navigate to the requested route
		cmd := r.NavigateTo(msg.Route)
		return r, cmd

	case BackMsg:
		// Go back
		cmd := r.Back()
		return r, cmd
	}

	// Delegate to bubblon controller for everything else
	updated, cmd := r.controller.Update(msg)
	r.controller = updated.(bubblon.Controller)
	return r, cmd
}

// View implements tea.Model
func (r *router) View() string {
	return r.controller.View()
}

// Navigation message types that views can send to trigger routing
type NavigateMsg struct {
	Route Route
}

type BackMsg struct{}
