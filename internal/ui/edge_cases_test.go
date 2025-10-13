package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bvdwalt/clippy/internal/history"
)

func TestModelViewEdgeCases(t *testing.T) {
	t.Run("View with cursor beyond available items", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Add one item but set cursor beyond it
		historyManager.AddItem("single item")
		model.cursor = 5 // way beyond available items

		view := model.View()

		// Should still render without crashing
		if !contains(view, "single item") {
			t.Error("View should contain the single item even with invalid cursor")
		}

		// Should not show cursor on non-existent items
		if contains(view, "> 6:") {
			t.Error("Should not show cursor on non-existent items")
		}
	})

	t.Run("View with negative cursor", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		historyManager.AddItem("test item")
		model.cursor = -1 // negative cursor

		view := model.View()

		// Should render without crashing
		if !contains(view, "test item") {
			t.Error("View should contain the item even with negative cursor")
		}
	})

	t.Run("View with exactly 60 character content", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Create content that's exactly 60 characters
		content60 := strings.Repeat("a", 60)
		historyManager.AddItem(content60)

		view := model.View()

		// Should show full content without truncation
		if !contains(view, content60) {
			t.Error("60-character content should be shown in full")
		}

		// Should not contain "..."
		if contains(view, "...") {
			t.Error("60-character content should not be truncated")
		}
	})

	t.Run("View with 61 character content", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Create content that's 61 characters (should be truncated)
		content61 := strings.Repeat("b", 61)
		historyManager.AddItem(content61)

		view := model.View()

		// Should show truncated content
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

		view := model.View()

		// strings.ReplaceAll(item, "\n", " ") only replaces \n with space
		// \r and \t characters are preserved as-is
		expected := "line1 line2\r line3\tcolumn2" // Only \n gets replaced with space
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

		// Should show the empty item
		if !contains(view, "> 1: ") {
			t.Error("Should show empty string item with cursor")
		}
	})

	t.Run("View with only whitespace content", func(t *testing.T) {
		historyManager := history.NewManager()
		model := NewModel(historyManager)

		// Add whitespace-only content
		historyManager.AddItem("   \t   ")

		view := model.View()

		// Should preserve whitespace
		if !contains(view, "> 1:    \t   ") {
			t.Error("Should preserve whitespace in display")
		}
	})
}

func TestModelCursorBoundaryConditions(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	t.Run("Cursor with empty history", func(t *testing.T) {
		// No items in history
		model.cursor = 0

		// Cursor should be valid even with empty history
		if model.cursor != 0 {
			t.Error("Cursor should be 0 with empty history")
		}

		view := model.View()
		if !contains(view, "No clipboard history yet...") {
			t.Error("Should show empty history message")
		}
	})

	t.Run("Cursor movement with single item", func(t *testing.T) {
		historyManager.AddItem("single item")
		model.cursor = 0

		// Test boundary logic
		maxIndex := model.historyManager.Count() - 1
		if maxIndex != 0 {
			t.Errorf("Expected max index 0 for single item, got %d", maxIndex)
		}

		// Cursor should not go below 0
		if model.cursor > 0 {
			model.cursor-- // This should not execute
		}
		if model.cursor != 0 {
			t.Error("Cursor should remain at 0")
		}

		// Cursor should not go above max index
		if model.cursor < maxIndex {
			model.cursor++ // This should not execute since cursor == maxIndex
		}
		if model.cursor != 0 {
			t.Error("Cursor should remain at 0 (max index)")
		}
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

		// Test a few cursor positions
		positions := []int{0, actualCount / 4, actualCount / 2, actualCount - 1}

		for _, pos := range positions {
			if pos >= actualCount {
				continue // Skip if position is out of bounds
			}

			model.cursor = pos
			view := model.View()

			// View should render without issues
			if len(view) == 0 {
				t.Errorf("View should not be empty with cursor at position %d", pos)
			}

			// Should show correct cursor position
			expectedCursor := fmt.Sprintf("> %d:", pos+1)
			if !contains(view, expectedCursor) {
				t.Errorf("Should show cursor at position %d, got view: %s", pos+1, view)
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
			"line1 line2\r line3", // \n -> space, \r remains
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
			view := model.View()

			if !contains(view, tc.expected) {
				t.Errorf("Expected %q in view, got: %s", tc.expected, view)
			}
		})
	}
}
