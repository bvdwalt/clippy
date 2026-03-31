package ui

import (
	"fmt"
	"log"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/search"
	"github.com/bvdwalt/clippy/internal/ui/styles"
	"github.com/bvdwalt/clippy/internal/ui/table"
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
	lastClipboard  string
	height         int
	width          int
	previewHeight  int
}

// NewModel creates a new UI model
func NewModel(historyManager *history.Manager) Model {
	ti := textinput.New()
	ti.Placeholder = "Search clipboard history..."
	ti.CharLimit = 50
	ti.SetWidth(50)

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

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return Tick()
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
				return m, nil
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
						if err := clipboard.WriteAll(items[selectedRow].Item); err != nil {
							log.Printf("Failed to write to clipboard: %v", err)
						}
						if err := m.historyManager.IncrementItemCount(selectedRow); err != nil {
							log.Printf("Failed to increment item count: %v", err)
						}
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
									m.lastClipboard = itemToDelete.Item
									m.updateTable()
								}
								break
							}
						}
					}
				}
			case "r":
				// Refresh/clear search and reload from database
				m.mode = TableView
				m.textInput.SetValue("")
				m.filtered = nil
				if err := m.historyManager.LoadFromDB(); err != nil {
					log.Printf("Failed to load from database: %v", err)
				}
				m.updateTable()
			default:
				// Handle table navigation (arrow keys, etc.)
				tbl := m.tableManager.GetTable()
				updatedTable, cmd := tbl.Update(msg)
				m.tableManager.SetTable(&updatedTable)
				return m, cmd
			}
		}

	case TickMsg:
		// Check for new clipboard content
		content, err := clipboard.ReadAll()
		if err == nil && len(content) > 0 {
			if content != m.lastClipboard {
				m.historyManager.AddItem(content)
				m.lastClipboard = content
			}
			m.updateTable()
		}
		return m, Tick()

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

		// Split available height: ~2/3 table, ~1/3 preview.
		// Overhead: title(2) + status(1) + help(2) + preview label(1) + preview borders(2) + doc margin(2) = 10
		available := max(msg.Height-10, 6)
		previewH := max(available/3, 3)
		m.previewHeight = previewH
		m.tableManager.SetSize(msg.Width, available-previewH)
	}

	return m, cmd
}

// View renders the UI
func (m Model) View() tea.View {
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
		v := tea.NewView(m.theme.Doc.Render(content.String()))
		v.AltScreen = true
		v.WindowTitle = "Clippy"
		return v
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

	// Preview pane
	if m.previewHeight > 0 {
		previewContent := ""
		if selected := m.tableManager.GetSelectedItem(); selected != nil {
			previewContent = selected.Item
		}
		previewWidth := max(m.width-8, 10) // doc margin (4 each side) + border (1 each side) + padding (1 each side)
		content.WriteString(m.theme.Help.Render("Preview") + "\n")
		content.WriteString(m.theme.Preview.Width(previewWidth).Height(m.previewHeight).Render(previewContent) + "\n")
	}

	// Status and help
	var status string
	if m.filtered != nil {
		status = fmt.Sprintf("Showing %d of %d items", len(m.filtered), m.historyManager.Count())
	} else {
		status = fmt.Sprintf("Total items: %d", len(items))
	}

	help := "Keys: \u2191/k \u2193/j navigate \u2022 Enter/c copy \u2022 d delete \u2022 / search \u2022 r refresh \u2022 q quit"
	if m.filtered != nil {
		help += " \u2022 esc clear search"
	}

	content.WriteString("\n" + status + "\n")
	content.WriteString(m.theme.Help.Render(help))

	v := tea.NewView(m.theme.Doc.Render(content.String()))
	v.AltScreen = true
	v.WindowTitle = "Clippy"
	return v
}

// GetCursor returns the current cursor position for testing
func (m Model) GetCursor() int {
	return m.tableManager.GetCursor()
}

// UpdateTable is a public wrapper for updateTable for testing purposes
func (m *Model) UpdateTable() {
	m.updateTable()
}
