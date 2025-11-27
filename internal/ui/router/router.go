package router

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Route represents a navigation destination
type Route string

const (
	RouteInit     Route = "init"
	RouteLive     Route = "live"
	RouteBacktest Route = "backtest"
	RouteAnalyze  Route = "analyze"
	RouteMenu     Route = "menu"
)

// Router is a simple service for navigation
type Router interface {
	// Navigate creates a navigation message
	Navigate(route Route) tea.Cmd

	// NavigateWithData creates a navigation message with context
	NavigateWithData(route Route, data interface{}) tea.Cmd
}

type router struct{}

// NewRouter creates a new router service
func NewRouter() Router {
	return &router{}
}

// Navigate creates a command to navigate to a route
func (r *router) Navigate(route Route) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Route: route}
	}
}

// NavigateWithData creates a command to navigate with data
func (r *router) NavigateWithData(route Route, data interface{}) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{
			Route: route,
			Data:  data,
		}
	}
}

// NavigateMsg is sent when navigating to a new route
type NavigateMsg struct {
	Route Route
	Data  interface{}
}
