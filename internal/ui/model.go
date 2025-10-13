package ui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/bvdwalt/clippy/internal/history"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the UI state
type Model struct {
	historyManager *history.Manager
	cursor         int
	height         int
}

// NewModel creates a new UI model
func NewModel(historyManager *history.Manager) Model {
	return Model{
		historyManager: historyManager,
		cursor:         0,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		Tick(),
		tea.EnterAltScreen,
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < m.historyManager.Count()-1 {
				m.cursor++
			}
		case "enter":
			if item, ok := m.historyManager.GetItem(m.cursor); ok {
				clipboard.WriteAll(item.Item)
			}
		}
	case TickMsg:
		content, err := clipboard.ReadAll()
		if err == nil && len(content) > 0 {
			m.historyManager.AddItem(content)
		}
		return m, Tick()
	case tea.WindowSizeMsg:
		m.height = msg.Height
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	s := "Clipboard History (press q to quit, enter to copy)\n\n"

	items := m.historyManager.GetItems()
	if len(items) == 0 {
		s += "No clipboard history yet...\n"
		return s
	}

	for i, historyItem := range items {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		item := historyItem.Item
		if len(item) > 60 {
			item = item[:57] + "..."
		}

		item = strings.ReplaceAll(item, "\n", " ")

		s += fmt.Sprintf("%s %d: %s\n", cursor, i+1, item)
	}

	return s
}
