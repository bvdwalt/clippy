package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// TickMsg is sent periodically to check for new clipboard content
type TickMsg time.Time

// Tick returns a command that sends a TickMsg every 500ms
func Tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
