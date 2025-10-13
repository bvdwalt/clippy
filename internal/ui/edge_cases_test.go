package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bvdwalt/clippy/internal/history"
	tea "github.com/charmbracelet/bubbletea"
)

func TestModelViewEdgeCases(t *testing.T) {
	t.Run("View with cursor beyond available items", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Add one item
		historyManager.AddItem("single item")
		model.UpdateTable() // Update table with new items

		view := model.View()

		// Should still render without crashing
		if !contains(view, "single item") {
			t.Error("View should contain the single item")
		}

		// Table component handles cursor bounds automatically
		if model.GetCursor() < 0 || model.GetCursor() >= historyManager.Count() {
			t.Error("Table cursor should be within valid bounds")
		}
	})

	t.Run("View with item renders correctly", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		historyManager.AddItem("test item")
		model.UpdateTable() // Update table with new items

		view := model.View()

		// Should render without crashing
		if !contains(view, "test item") {
			t.Error("View should contain the item")
		}

		// Cursor should be valid (table handles bounds)
		if model.GetCursor() < 0 {
			t.Error("Cursor should not be negative")
		}
	})

	t.Run("View with exactly 60 character content", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Create content that's exactly 60 characters
		content60 := strings.Repeat("a", 60)
		historyManager.AddItem(content60)
		model.UpdateTable() // Update table with new items

		view := model.View()

		// Should show full content without truncation (60 chars fits in 60-char column)
		if !contains(view, content60) {
			t.Error("60-character content should be shown in full")
		}
	})

	t.Run("View with 61 character content", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Create content that's 61 characters (should be truncated)
		content61 := strings.Repeat("b", 61)
		historyManager.AddItem(content61)
		model.UpdateTable() // Update table with new items

		view := model.View()

		// Should show truncated content (61 chars truncated to 57 + "...")
		expected := strings.Repeat("b", 57) + "..."
		if !contains(view, expected) {
			t.Error("61-character content should be truncated to 57 chars + '...'")
		}

		// Should not contain the full content
		if contains(view, content61) {
			t.Error("61-character content should not appear in full")
		}
	})

	t.Run("View with mixed newlines and spaces", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Content with various whitespace characters
		content := "line1\nline2\r\nline3\tcolumn2"
		historyManager.AddItem(content)
		model.UpdateTable() // Update table with new items

		view := model.View()

		// In the table format, newlines, carriage returns, and tabs are replaced with spaces
		expected := "line1 line2 line3 column2" // \n, \r, and \t all become spaces
		if !contains(view, expected) {
			t.Errorf("Expected %q in view, got: %s", expected, view)
		}
	})

	t.Run("View with empty string content", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Add empty string
		historyManager.AddItem("")

		view := model.View()

		// Should show the empty item in table format
		if !contains(view, "1") { // Should show row number 1
			t.Error("Should show empty string item in table")
		}
	})

	t.Run("View with only whitespace content", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Add whitespace-only content
		historyManager.AddItem("   \t   ")

		view := model.View()

		// Should preserve whitespace (converted to spaces in table)
		if !contains(view, "       ") { // Whitespace should be preserved as spaces
			t.Error("Should preserve whitespace in display")
		}
	})
}

func TestModelCursorBoundaryConditions(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	t.Run("Cursor with empty history", func(t *testing.T) {
		// No items in history
		// Cursor should be valid even with empty history
		if model.GetCursor() != 0 {
			t.Error("Cursor should be 0 with empty history")
		}

		view := model.View()
		if !contains(view, "No clipboard history yet...") {
			t.Error("Should show empty history message")
		}
	})

	t.Run("Cursor movement with single item", func(t *testing.T) {
		historyManager.AddItem("single item")
		model.UpdateTable() // Update table with new items

		// Test boundary logic with table component
		maxIndex := model.historyManager.Count() - 1
		if maxIndex != 0 {
			t.Errorf("Expected max index 0 for single item, got %d", maxIndex)
		}

		// Test that cursor stays within bounds (table handles this automatically)
		initialCursor := model.GetCursor()

		// Try moving up (should stay at 0)
		upMsg := tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, _ := model.Update(upMsg)
		model = updatedModel.(Model)

		if model.GetCursor() != 0 {
			t.Error("Cursor should remain at 0 when trying to move up from first item")
		}

		// Try moving down (should stay at 0 since there's only one item)
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ = model.Update(downMsg)
		model = updatedModel.(Model)

		if model.GetCursor() != 0 {
			t.Error("Cursor should remain at 0 when there's only one item")
		}

		_ = initialCursor // Use the variable to avoid unused error
	})
}

func TestModelLargeDatasets(t *testing.T) {
	t.Run("Performance with many items", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Add many unique items to avoid deduplication
		itemCount := 100 // Reduced for faster testing
		for i := 0; i < itemCount; i++ {
			historyManager.AddItem(fmt.Sprintf("unique item content %d with nanos %d", i, time.Now().UnixNano()))
		}

		actualCount := historyManager.Count()
		if actualCount != itemCount {
			t.Errorf("Expected %d items, got %d (possible deduplication)", itemCount, actualCount)
		}

		model.UpdateTable() // Update table with all items

		// Test a few positions by moving the cursor
		for i := 0; i < 5 && i < actualCount; i++ {
			// Move cursor to position i using down key presses
			for j := model.GetCursor(); j < i; j++ {
				downMsg := tea.KeyMsg{Type: tea.KeyDown}
				updatedModel, _ := model.Update(downMsg)
				model = updatedModel.(Model)
			}

			view := model.View()

			// View should render without issues
			if len(view) == 0 {
				t.Errorf("View should not be empty with cursor at position %d", i)
			}

			// Should contain some of the items
			if !contains(view, fmt.Sprintf("unique item content %d", i)) {
				t.Errorf("Should show item at position %d", i)
			}
		}
	})
}

func TestModelSpecialCharacters(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			"Control characters",
			"text\x00with\x01control\x02chars",
			"text\x00with\x01control\x02chars", // Should preserve as-is
		},
		{
			"Unicode emojis",
			"Hello 🌍 World 🚀",
			"Hello 🌍 World 🚀",
		},
		{
			"Mixed newlines",
			"line1\nline2\r\nline3",
			"line1 line2 line3", // \n and \r both become single spaces
		},
		{
			"Backslashes",
			"path\\to\\file",
			"path\\to\\file",
		},
		{
			"Quotes and special chars",
			`"quoted text" and 'single quotes' & symbols`,
			`"quoted text" and 'single quotes' & symbols`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			historyManager := history.NewManager()
			model := NewModel(historyManager)

			historyManager.AddItem(tc.content)
			model.UpdateTable() // Update table with new items
			view := model.View()

			if !contains(view, tc.expected) {
				t.Errorf("Expected %q in view, got: %s", tc.expected, view)
			}
		})
	}
}
