package router

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Route represents a navigation destination
type Route string

const (
	RouteInit             Route = "init"
	RouteLive             Route = "live"
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

// NavigateMsg is sent when navigating to a new route
type NavigateMsg struct {
	Route Route
	Data  interface{}
	View  tea.Model // The view is created externally with its dependencies
}

// Router is the main Tea model that manages view navigation
// It does NOT create views or store services
type Router interface {
	tea.Model

	// SetInitialView sets the starting view
	SetInitialView(view tea.Model)

	// Navigate creates a navigation message
	Navigate(route Route) tea.Cmd

	// NavigateWithData creates a navigation message with context
	NavigateWithData(route Route, data interface{}) tea.Cmd
}

type router struct {
	currentRoute Route
	currentView  tea.Model
}

// NewRouter creates a simple router - no services, no factories
func NewRouter() Router {
	return &router{
		currentRoute: RouteStrategyList,
	}
}

// SetInitialView sets the starting view
func (r *router) SetInitialView(view tea.Model) {
	r.currentView = view
}

// Init implements tea.Model
func (r *router) Init() tea.Cmd {
	if r.currentView != nil {
		return r.currentView.Init()
	}
	return nil
}

// Update implements tea.Model
func (r *router) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle navigation messages
	if navMsg, ok := msg.(NavigateMsg); ok {
		// Navigation message contains the new view already created with its dependencies
		if navMsg.View != nil {
			r.currentRoute = navMsg.Route
			r.currentView = navMsg.View
			return r, r.currentView.Init()
		}
	}

	// Pass other messages to current view
	if r.currentView == nil {
		return r, nil
	}

	updated, cmd := r.currentView.Update(msg)
	r.currentView = updated
	return r, cmd
}

// View implements tea.Model
func (r *router) View() string {
	if r.currentView == nil {
		return ""
	}
	return r.currentView.View()
}

// Navigate creates a navigation command
func (r *router) Navigate(route Route) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Route: route}
	}
}

// NavigateWithData creates a navigation command with data
func (r *router) NavigateWithData(route Route, data interface{}) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{
			Route: route,
			Data:  data,
		}
	}
}
