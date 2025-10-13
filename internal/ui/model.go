package ui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/search"
	"github.com/bvdwalt/clippy/internal/ui/styles"
	"github.com/bvdwalt/clippy/internal/ui/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	tableManager   *table.Manager
	textInput      textinput.Model
	fuzzyMatcher   *search.FuzzyMatcher
	theme          styles.Theme
	mode           ViewMode
	filtered       []history.ClipboardHistory
	height         int
	width          int
}

// NewModel creates a new UI model
func NewModel(historyManager *history.Manager) Model {
	ti := textinput.New()
	ti.Placeholder = "Search clipboard history..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 50

	theme := styles.DefaultTheme()
	tableTheme := styles.DefaultTableTheme()
	tableManager := table.NewManager(tableTheme)
	fuzzyMatcher := search.NewFuzzyMatcher()

	m := Model{
		historyManager: historyManager,
		tableManager:   tableManager,
		textInput:      ti,
		fuzzyMatcher:   fuzzyMatcher,
		theme:          theme,
		mode:           TableView,
	}

	m.updateTable()
	return m
}

// updateTable refreshes the table with current (filtered) history items
func (m *Model) updateTable() {
	items := m.getDisplayItems()
	m.tableManager.UpdateRows(items)
}

// getDisplayItems returns the items to display (filtered or all)
func (m *Model) getDisplayItems() []history.ClipboardHistory {
	if m.filtered != nil {
		return m.filtered
	}
	return m.historyManager.GetItems()
}

// filterItems filters history items using fuzzy finding (like fzf)
func (m *Model) filterItems(query string) {
	if query == "" {
		m.filtered = nil
		return
	}

	allItems := m.historyManager.GetItems()
	m.filtered = m.fuzzyMatcher.Search(allItems, query)
}

// isLower checks if a rune is lowercase
func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

// isUpper checks if a rune is uppercase
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
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
					selectedRow := m.tableManager.GetCursor()
					if selectedRow < len(items) {
						clipboard.WriteAll(items[selectedRow].Item)
					}
				}
			case "d":
				// Delete selected item
				items := m.getDisplayItems()
				if len(items) > 0 {
					selectedRow := m.tableManager.GetCursor()
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
				table := m.tableManager.GetTable()
				table, cmd = table.Update(msg)
				m.tableManager.SetTable(table)
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
		m.tableManager.SetSize(msg.Width, msg.Height)
	}

	return m, cmd
}

// View renders the UI
func (m Model) View() string {
	var content strings.Builder

	// Title
	title := m.theme.Title.Render("📋 Clippy Clipboard History")
	content.WriteString(title + "\n\n")

	// Search mode UI
	if m.mode == SearchView {
		searchBox := m.theme.Search.Render(
			fmt.Sprintf("🔍 Search:\n\n%s\n\n%s",
				m.textInput.View(),
				m.theme.Help.Render("Press Enter to search, Esc to cancel")))
		content.WriteString(searchBox + "\n")
		return m.theme.Doc.Render(content.String())
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
		content.WriteString(m.tableManager.View() + "\n")
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
	content.WriteString(m.theme.Help.Render(help))

	return m.theme.Doc.Render(content.String())
}

// GetCursor returns the current cursor position for testing
func (m Model) GetCursor() int {
	return m.tableManager.GetCursor()
}

// SetCursor sets the cursor position for testing
func (m *Model) SetCursor(pos int) {
	// Placeholder for test compatibility
}

// UpdateTable is a public wrapper for updateTable for testing purposes
func (m *Model) UpdateTable() {
	m.updateTable()
}
