package table

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/ui/styles"
)

// Manager handles table creation and updates
type Manager struct {
	table        *table.Model
	theme        styles.TableTheme
	lastItems    []history.ClipboardHistory // lastItems holds the items currently displayed (for stable selection)
	contentWidth int
}

// NewManager creates a new table manager
func NewManager(theme styles.TableTheme) *Manager {
	columns := []table.Column{
		{Title: "#", Width: 5},
		{Title: "Content", Width: 60},
		{Title: "Pin", Width: 5},
		{Title: "Time", Width: 19},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
		table.WithWidth(80),
	)

	// Use centralized conversion for table styles
	s := styles.TableStyles(theme)
	// table.New returns a value; take its address to use pointer receivers
	t.SetStyles(s)
	return &Manager{
		table:        &t,
		theme:        theme,
		lastItems:    nil,
		contentWidth: 60,
	}
}

// GetTable returns the underlying table model
func (tm *Manager) GetTable() *table.Model {
	return tm.table
}

func (tm *Manager) SetTable(t *table.Model) {
	// When replacing the underlying table, clear lastItems to avoid mismatches
	tm.table = t
	tm.lastItems = nil
}

// UpdateRows updates the table with clipboard history items
func (tm *Manager) UpdateRows(items []history.ClipboardHistory) {
	if tm.table == nil {
		return
	}

	// Capture previous selected item's hash for stable selection
	prevCursor := tm.table.Cursor()
	var prevHash string
	if prevCursor >= 0 && tm.lastItems != nil && prevCursor < len(tm.lastItems) {
		prevHash = tm.lastItems[prevCursor].Hash
	}

	rows := make([]table.Row, len(items))
	for i, item := range items {
		content := item.Item
		content = strings.ReplaceAll(content, "\r\n", " ")
		content = strings.ReplaceAll(content, "\n", " ")
		content = strings.ReplaceAll(content, "\r", " ")
		content = strings.ReplaceAll(content, "\t", " ")

		if tm.contentWidth > 3 && len(content) > tm.contentWidth {
			content = content[:tm.contentWidth-3] + "..."
		}

		pin := ""
		if item.Pinned {
			pin = "📌"
		}
		rows[i] = table.Row{
			strconv.Itoa(i + 1),
			content,
			pin,
			item.TimeStamp.Format("2006-01-02 15:04:05"),
		}
	}

	// Update stored items before restoring selection so we can search the new list
	tm.lastItems = make([]history.ClipboardHistory, len(items))
	copy(tm.lastItems, items)

	// Apply rows to table
	tm.table.SetRows(rows)

	// Restore selection by hash if possible, otherwise clamp previous cursor
	if prevHash != "" {
		found := -1
		for i, it := range items {
			if it.Hash == prevHash {
				found = i
				break
			}
		}
		if found >= 0 {
			tm.table.SetCursor(found)
		} else {
			// fallback: clamp prevCursor into range
			if len(rows) == 0 {
				tm.table.SetCursor(0)
			} else {
				if prevCursor < 0 {
					prevCursor = 0
				}
				if prevCursor > len(rows)-1 {
					prevCursor = len(rows) - 1
				}
				tm.table.SetCursor(prevCursor)
			}
		}
	} else {
		// No previous hash: just clamp prevCursor
		if len(rows) == 0 {
			tm.table.SetCursor(0)
		} else {
			if prevCursor < 0 {
				prevCursor = 0
			}
			if prevCursor > len(rows)-1 {
				prevCursor = len(rows) - 1
			}
			tm.table.SetCursor(prevCursor)
		}
	}

	// Recompute viewport content after row updates.
	tm.table.UpdateViewport()
}

// SetSize updates the table dimensions
func (tm *Manager) SetSize(width, height int) {
	if tm.table == nil {
		return
	}

	tableWidth := width - 4
	contentWidth := tableWidth - 29 - 4
	contentWidth = max(contentWidth, 20)
	tm.contentWidth = contentWidth

	tm.table.SetColumns([]table.Column{
		{Title: "#", Width: 5},
		{Title: "Content", Width: contentWidth},
		{Title: "Pin", Width: 5},
		{Title: "Time", Width: 19},
	})
	tm.table.SetWidth(tableWidth)
	tm.table.SetHeight(height)

	if tm.lastItems != nil {
		tm.UpdateRows(tm.lastItems)
	}
	// Ensure viewport matches the new size.
	tm.table.UpdateViewport()
}

// GetCursor returns the current cursor position
func (tm *Manager) GetCursor() int {
	if tm.table == nil {
		return 0
	}
	cursor := tm.table.Cursor()
	if cursor < 0 {
		return 0
	}
	return cursor
}

// GetSelectedItem returns the currently selected clipboard item, or nil if none.
func (tm *Manager) GetSelectedItem() *history.ClipboardHistory {
	if tm.table == nil || len(tm.lastItems) == 0 {
		return nil
	}
	cursor := tm.table.Cursor()
	if cursor < 0 || cursor >= len(tm.lastItems) {
		return nil
	}
	item := tm.lastItems[cursor]
	return &item
}

// View returns the table view
func (tm *Manager) View() string {
	if tm.table == nil {
		return ""
	}
	return tm.table.View()
}
