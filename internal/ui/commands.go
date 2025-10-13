package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TickMsg is sent periodically to check for new clipboard content
type TickMsg time.Time

// Tick returns a command that sends a TickMsg every 2 seconds
func Tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
