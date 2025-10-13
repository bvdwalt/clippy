package ui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/bvdwalt/clippy/internal/history"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	TableView ViewMode = iota
	SearchView
)

// Model represents the UI state
type Model struct {
	historyManager *history.Manager
	table          table.Model
	textInput      textinput.Model
	mode           ViewMode
	filtered       []history.ClipboardHistory
	height         int
	width          int
}

// Styles for the enhanced UI
var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Margin(1, 0)

	searchStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Width(50)
)

// NewModel creates a new UI model
func NewModel(historyManager *history.Manager) Model {
	// Initialize text input for search
	ti := textinput.New()
	ti.Placeholder = "Search clipboard history..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 50

	// Initialize table
	columns := []table.Column{
		{Title: "#", Width: 4},
		{Title: "Content", Width: 60},
		{Title: "Time", Width: 19},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := Model{
		historyManager: historyManager,
		table:          t,
		textInput:      ti,
		mode:           TableView,
	}

	m.updateTable()
	return m
}

// updateTable refreshes the table with current (filtered) history items
func (m *Model) updateTable() {
	items := m.getDisplayItems()

	rows := make([]table.Row, len(items))
	for i, item := range items {
		content := item.Item
		// First, replace whitespace characters (\r\n should become single space)
		content = strings.ReplaceAll(content, "\r\n", " ")
		content = strings.ReplaceAll(content, "\n", " ")
		content = strings.ReplaceAll(content, "\r", " ")
		content = strings.ReplaceAll(content, "\t", " ")

		// Then truncate if needed (table column width is 60)
		if len(content) > 60 {
			content = content[:57] + "..."
		}

		rows[i] = table.Row{
			fmt.Sprintf("%d", i+1),
			content,
			item.TimeStamp.Format("2006-01-02 15:04:05"),
		}
	}

	m.table.SetRows(rows)
}

// getDisplayItems returns the items to display (filtered or all)
func (m *Model) getDisplayItems() []history.ClipboardHistory {
	if m.mode == SearchView && len(m.filtered) > 0 {
		return m.filtered
	}
	return m.historyManager.GetItems()
}

// filterItems filters history items based on search query
func (m *Model) filterItems(query string) {
	if query == "" {
		m.filtered = nil
		return
	}

	query = strings.ToLower(query)
	allItems := m.historyManager.GetItems()
	m.filtered = make([]history.ClipboardHistory, 0)

	for _, item := range allItems {
		if strings.Contains(strings.ToLower(item.Item), query) {
			m.filtered = append(m.filtered, item)
		}
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global shortcuts that work in any mode
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "/":
			// Toggle search mode
			if m.mode == TableView {
				m.mode = SearchView
				m.textInput.Focus()
				return m, textinput.Blink
			}
		case "esc":
			// Exit search mode
			if m.mode == SearchView {
				m.mode = TableView
				m.textInput.Blur()
				m.textInput.SetValue("")
				m.filtered = nil
				m.updateTable()
				return m, nil
			}
		}

		// Mode-specific key handling
		switch m.mode {
		case SearchView:
			switch msg.String() {
			case "enter":
				// Apply search filter
				m.filterItems(m.textInput.Value())
				m.updateTable()
				m.mode = TableView
				m.textInput.Blur()
				return m, nil
			default:
				// Handle text input
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		case TableView:
			switch msg.String() {
			case "enter", "c":
				// Copy selected item
				items := m.getDisplayItems()
				if len(items) > 0 {
					selectedRow := m.table.Cursor()
					if selectedRow < len(items) {
						clipboard.WriteAll(items[selectedRow].Item)
					}
				}
			case "d":
				// Delete selected item
				items := m.getDisplayItems()
				if len(items) > 0 {
					selectedRow := m.table.Cursor()
					if selectedRow < len(items) {
						// Find the original index in history manager
						itemToDelete := items[selectedRow]
						allItems := m.historyManager.GetItems()
						for i, item := range allItems {
							if item.Hash == itemToDelete.Hash {
								if m.historyManager.DeleteItem(i) {
									m.updateTable()
								}
								break
							}
						}
					}
				}
			case "r":
				// Refresh/clear search
				m.mode = TableView
				m.textInput.SetValue("")
				m.filtered = nil
				m.updateTable()
			default:
				// Handle table navigation
				m.table, cmd = m.table.Update(msg)
			}
		}

	case TickMsg:
		// Check for new clipboard content
		content, err := clipboard.ReadAll()
		if err == nil && len(content) > 0 {
			m.historyManager.AddItem(content)
			m.updateTable()
		}
		return m, Tick()

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		// Update table size
		m.table.SetWidth(msg.Width - 4)
		m.table.SetHeight(msg.Height - 8)
	}

	return m, cmd
}

// View renders the UI
func (m Model) View() string {
	var content strings.Builder

	// Title
	title := titleStyle.Render("📋 Clippy Clipboard History")
	content.WriteString(title + "\n\n")

	// Search mode UI
	if m.mode == SearchView {
		searchBox := searchStyle.Render(
			fmt.Sprintf("🔍 Search:\n\n%s\n\n%s",
				m.textInput.View(),
				helpStyle.Render("Press Enter to search, Esc to cancel")))
		content.WriteString(searchBox + "\n")
		return docStyle.Render(content.String())
	}

	// Table view
	items := m.getDisplayItems()
	if len(items) == 0 {
		if m.filtered != nil {
			content.WriteString("No results found for your search.\n")
		} else {
			content.WriteString("No clipboard history yet...\n")
		}
	} else {
		content.WriteString(m.table.View() + "\n")
	}

	// Status and help
	var status string
	if m.filtered != nil {
		status = fmt.Sprintf("Showing %d of %d items", len(m.filtered), m.historyManager.Count())
	} else {
		status = fmt.Sprintf("Total items: %d", len(items))
	}

	help := "Keys: ↑/k ↓/j navigate • Enter/c copy • d delete • / search • r refresh • q quit"
	if m.filtered != nil {
		help += " • esc clear search"
	}

	content.WriteString("\n" + status + "\n")
	content.WriteString(helpStyle.Render(help))

	return docStyle.Render(content.String())
}

// GetCursor returns the current cursor position for testing
func (m Model) GetCursor() int {
	return m.table.Cursor()
}

// SetCursor sets the cursor position for testing
func (m *Model) SetCursor(pos int) {
	// We'll need to simulate key presses to move the cursor
	// For now, this is a placeholder for test compatibility
}

// UpdateTable is a public wrapper for updateTable for testing purposes
func (m *Model) UpdateTable() {
	m.updateTable()
}
