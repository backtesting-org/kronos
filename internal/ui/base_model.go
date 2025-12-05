package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/donderom/bubblon"
)

// BaseModel provides common keyboard handling for all views
// Embed this in your models to get consistent quit/back behavior
type BaseModel struct {
	// Set to true if this is a root-level view (launched from main menu)
	IsRoot bool
}

// HandleCommonKeys processes common keyboard shortcuts
// Returns (handled, cmd) - if handled is true, the key was processed
func (b *BaseModel) HandleCommonKeys(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Always quit the entire program
		return true, tea.Quit

	case "q", "esc":
		if b.IsRoot {
			// At root level, quit entirely
			return true, tea.Quit
		}
		// Otherwise, close this view and go back
		return true, bubblon.Close
	}

	return false, nil
}

// WrapUpdate wraps your Update function to handle common keys first
// Usage in your model's Update:
//
//	func (m *myModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	    if keyMsg, ok := msg.(tea.KeyMsg); ok {
//	        if handled, cmd := m.BaseModel.HandleCommonKeys(keyMsg); handled {
//	            return m, cmd
//	        }
//	    }
//	    // ... rest of your update logic
//	}
