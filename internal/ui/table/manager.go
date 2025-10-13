package table

import (
	"fmt"
	"strings"

	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/ui/styles"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// Manager handles table creation and updates
type Manager struct {
	table table.Model
	theme styles.TableTheme
}

// NewManager creates a new table manager
func NewManager(theme styles.TableTheme) *Manager {
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
		BorderForeground(theme.HeaderBorderColor).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(theme.SelectedFg).
		Background(theme.SelectedBg).
		Bold(false)
	t.SetStyles(s)

	return &Manager{
		table: t,
		theme: theme,
	}
}

// GetTable returns the underlying table model
func (tm *Manager) GetTable() table.Model {
	return tm.table
}

func (tm *Manager) SetTable(t table.Model) {
	tm.table = t
}

// UpdateRows updates the table with clipboard history items
func (tm *Manager) UpdateRows(items []history.ClipboardHistory) {
	rows := make([]table.Row, len(items))
	for i, item := range items {
		content := item.Item
		content = strings.ReplaceAll(content, "\r\n", " ")
		content = strings.ReplaceAll(content, "\n", " ")
		content = strings.ReplaceAll(content, "\r", " ")
		content = strings.ReplaceAll(content, "\t", " ")

		if len(content) > 60 {
			content = content[:57] + "..."
		}

		rows[i] = table.Row{
			fmt.Sprintf("%d", i+1),
			content,
			item.TimeStamp.Format("2006-01-02 15:04:05"),
		}
	}

	tm.table.SetRows(rows)
}

// SetSize updates the table dimensions
func (tm *Manager) SetSize(width, height int) {
	tm.table.SetWidth(width - 4)
	tm.table.SetHeight(height - 8)
}

// GetCursor returns the current cursor position
func (tm *Manager) GetCursor() int {
	return tm.table.Cursor()
}

// View returns the table view
func (tm *Manager) View() string {
	return tm.table.View()
}
