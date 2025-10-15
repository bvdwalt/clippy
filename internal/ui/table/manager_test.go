package table

import (
	"strings"
	"testing"
	"time"

	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/ui/styles"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func TestNewManager(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	if manager == nil {
		t.Fatal("NewManager should return a non-nil manager")
	}

	// Test that the table is properly initialized (check it produces output)
	view := manager.View()
	if view == "" {
		t.Error("Expected table to be initialized and produce output")
	}

	// Test that theme is stored
	if manager.theme != theme {
		t.Error("Expected theme to be stored in manager")
	}

	// Test initial cursor position
	if manager.GetCursor() != 0 {
		t.Errorf("Expected initial cursor position to be 0, got %d", manager.GetCursor())
	}
}

func TestNewManagerWithCustomTheme(t *testing.T) {
	customTheme := styles.TableTheme{
		HeaderBorderColor: lipgloss.Color("100"),
		SelectedFg:        lipgloss.Color("200"),
		SelectedBg:        lipgloss.Color("50"),
	}

	manager := NewManager(customTheme)

	if manager.theme.HeaderBorderColor != customTheme.HeaderBorderColor {
		t.Error("Custom HeaderBorderColor not applied")
	}

	if manager.theme.SelectedFg != customTheme.SelectedFg {
		t.Error("Custom SelectedFg not applied")
	}

	if manager.theme.SelectedBg != customTheme.SelectedBg {
		t.Error("Custom SelectedBg not applied")
	}
}

func TestUpdateRows(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	// Test with empty items
	t.Run("Empty items", func(t *testing.T) {
		var items []history.ClipboardHistory
		manager.UpdateRows(items)

		view := manager.View()
		// Should not panic and should return a view
		if view == "" {
			t.Error("Expected non-empty view even with empty items")
		}
	})

	// Test with single item
	t.Run("Single item", func(t *testing.T) {
		items := []history.ClipboardHistory{
			{
				Item:      "test content",
				Hash:      "hash1",
				TimeStamp: time.Date(2023, 10, 13, 12, 0, 0, 0, time.UTC),
			},
		}

		manager.UpdateRows(items)
		view := manager.View()

		// Should contain the item content
		if !strings.Contains(view, "test content") {
			t.Error("Expected view to contain item content")
		}

		// Should contain the row number
		if !strings.Contains(view, "1") {
			t.Error("Expected view to contain row number")
		}

		// Should contain formatted timestamp
		if !strings.Contains(view, "2023-10-13 12:00:00") {
			t.Error("Expected view to contain formatted timestamp")
		}
	})

	// Test with multiple items
	t.Run("Multiple items", func(t *testing.T) {
		items := []history.ClipboardHistory{
			{
				Item:      "first item",
				Hash:      "hash1",
				TimeStamp: time.Date(2023, 10, 13, 12, 0, 0, 0, time.UTC),
			},
			{
				Item:      "second item",
				Hash:      "hash2",
				TimeStamp: time.Date(2023, 10, 13, 13, 0, 0, 0, time.UTC),
			},
			{
				Item:      "third item",
				Hash:      "hash3",
				TimeStamp: time.Date(2023, 10, 13, 14, 0, 0, 0, time.UTC),
			},
		}

		manager.UpdateRows(items)
		view := manager.View()

		// Should contain all items
		expectedContents := []string{"first item", "second item", "third item"}
		for _, content := range expectedContents {
			if !strings.Contains(view, content) {
				t.Errorf("Expected view to contain %q", content)
			}
		}

		// Should contain row numbers
		expectedNumbers := []string{"1", "2", "3"}
		for _, number := range expectedNumbers {
			if !strings.Contains(view, number) {
				t.Errorf("Expected view to contain row number %q", number)
			}
		}
	})
}

func TestUpdateRowsContentFormatting(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	testCases := []struct {
		name            string
		input           string
		expectedContent string
	}{
		{
			"Newlines replaced",
			"line1\nline2\nline3",
			"line1 line2 line3",
		},
		{
			"Carriage returns replaced",
			"line1\rline2\rline3",
			"line1 line2 line3",
		},
		{
			"Windows newlines replaced",
			"line1\r\nline2\r\nline3",
			"line1 line2 line3",
		},
		{
			"Tabs replaced",
			"col1\tcol2\tcol3",
			"col1 col2 col3",
		},
		{
			"Mixed whitespace",
			"line1\nline2\tcolumn\rend",
			"line1 line2 column end",
		},
		{
			"Short content unchanged",
			"short text",
			"short text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			items := []history.ClipboardHistory{
				{
					Item:      tc.input,
					Hash:      "hash1",
					TimeStamp: time.Now(),
				},
			}

			manager.UpdateRows(items)
			view := manager.View()

			if !strings.Contains(view, tc.expectedContent) {
				t.Errorf("Expected view to contain %q, got view: %s", tc.expectedContent, view)
			}
		})
	}
}

func TestUpdateRowsContentTruncation(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	t.Run("Content exactly 60 chars", func(t *testing.T) {
		content := strings.Repeat("a", 60) // Exactly 60 characters
		items := []history.ClipboardHistory{
			{
				Item:      content,
				Hash:      "hash1",
				TimeStamp: time.Now(),
			},
		}

		manager.UpdateRows(items)
		view := manager.View()

		// Should show full content (60 chars fits)
		if !strings.Contains(view, content) {
			t.Error("Expected view to contain full 60-character content")
		}

		// Should not contain truncation indicator
		if strings.Contains(view, "...") {
			t.Error("Expected no truncation for 60-character content")
		}
	})

	t.Run("Content longer than 60 chars", func(t *testing.T) {
		content := strings.Repeat("b", 70) // 70 characters
		expectedTruncated := strings.Repeat("b", 57) + "..."

		items := []history.ClipboardHistory{
			{
				Item:      content,
				Hash:      "hash1",
				TimeStamp: time.Now(),
			},
		}

		manager.UpdateRows(items)
		view := manager.View()

		// Should show truncated content
		if !strings.Contains(view, expectedTruncated) {
			t.Error("Expected view to contain truncated content with ellipsis")
		}

		// Should not contain full content
		if strings.Contains(view, content) {
			t.Error("Expected view not to contain full long content")
		}
	})

	t.Run("Content much longer than 60 chars", func(t *testing.T) {
		content := strings.Repeat("very long content ", 20) // Much longer than 60
		expectedTruncated := content[:57] + "..."

		items := []history.ClipboardHistory{
			{
				Item:      content,
				Hash:      "hash1",
				TimeStamp: time.Now(),
			},
		}

		manager.UpdateRows(items)
		view := manager.View()

		if !strings.Contains(view, expectedTruncated) {
			t.Error("Expected view to contain properly truncated content")
		}
	})
}

func TestSetSize(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	testCases := []struct {
		name   string
		width  int
		height int
	}{
		{"Small size", 50, 20},
		{"Medium size", 100, 40},
		{"Large size", 200, 60},
		{"Minimum size", 10, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager.SetSize(tc.width, tc.height)

			// Test that the method doesn't panic
			view := manager.View()
			if view == "" {
				t.Error("Expected non-empty view after setting size")
			}

			// We can't easily test the exact dimensions without accessing internal state,
			// but we can verify the operation completed successfully
		})
	}
}

func TestGetCursor(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	// Test initial cursor position
	initialCursor := manager.GetCursor()
	if initialCursor != 0 {
		t.Errorf("Expected initial cursor to be 0, got %d", initialCursor)
	}

	// Add some items to test cursor with content
	items := []history.ClipboardHistory{
		{Item: "item1", Hash: "hash1", TimeStamp: time.Now()},
		{Item: "item2", Hash: "hash2", TimeStamp: time.Now()},
		{Item: "item3", Hash: "hash3", TimeStamp: time.Now()},
	}

	manager.UpdateRows(items)

	// Cursor should still be 0 after adding items
	cursor := manager.GetCursor()
	if cursor < 0 {
		t.Errorf("Expected cursor to be non-negative, got %d", cursor)
	}
}

func TestView(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	// Test view with no data
	t.Run("Empty view", func(t *testing.T) {
		view := manager.View()
		if view == "" {
			t.Error("Expected non-empty view even with no data")
		}
	})

	// Test view with data
	t.Run("View with data", func(t *testing.T) {
		items := []history.ClipboardHistory{
			{
				Item:      "test item",
				Hash:      "hash1",
				TimeStamp: time.Date(2023, 10, 13, 12, 0, 0, 0, time.UTC),
			},
		}

		manager.UpdateRows(items)
		view := manager.View()

		// Should contain table headers
		if !strings.Contains(view, "#") || !strings.Contains(view, "Content") || !strings.Contains(view, "Time") {
			t.Error("Expected view to contain table headers")
		}

		// Should contain the data
		if !strings.Contains(view, "test item") {
			t.Error("Expected view to contain item content")
		}

		if !strings.Contains(view, "2023-10-13 12:00:00") {
			t.Error("Expected view to contain formatted timestamp")
		}
	})
}

func TestGetSetTable(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	// Test GetTable - verify it returns a working table
	view := manager.View()
	if view == "" {
		t.Error("Expected GetTable to return functional table")
	}

	// Test SetTable
	newTable := table.New()
	manager.SetTable(newTable)

	retrievedTable := manager.GetTable()
	// Note: We can't directly compare table.Model instances for equality,
	// but we can verify the operation completed without panic by checking it works
	view = manager.View()
	if view == "" {
		t.Error("Expected SetTable/GetTable to work correctly")
	}

	_ = retrievedTable // Use the variable to avoid unused error
}

func TestManagerWithSpecialCharacterContent(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	testCases := []struct {
		name    string
		content string
	}{
		{"Unicode emojis", "🎉 📋 🚀 Testing emojis"},
		{"Unicode text", "你好世界 こんにちは مرحبا"},
		{"Special symbols", "!@#$%^&*()_+-=[]{}|;:,.<>?"},
		{"HTML content", "<div>HTML content with &amp; entities</div>"},
		{"JSON content", `{"key": "value", "number": 123, "array": [1,2,3]}`},
		{"Code content", "func main() {\n\tfmt.Println(\"Hello\")\n}"},
		{"Mixed content", "Mixed: 🌍 with\nnewlines\tand\rtabs!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			items := []history.ClipboardHistory{
				{
					Item:      tc.content,
					Hash:      "hash1",
					TimeStamp: time.Now(),
				},
			}

			// Should not panic
			manager.UpdateRows(items)
			view := manager.View()

			// Should produce non-empty view
			if view == "" {
				t.Errorf("Expected non-empty view for content: %q", tc.content)
			}

			// Should contain some form of the content (might be formatted/truncated)
			// For content with newlines/tabs, they should be replaced with spaces
			expectedContent := strings.ReplaceAll(tc.content, "\n", " ")
			expectedContent = strings.ReplaceAll(expectedContent, "\r", " ")
			expectedContent = strings.ReplaceAll(expectedContent, "\t", " ")

			// If content is too long, it will be truncated
			if len(expectedContent) > 60 {
				expectedContent = expectedContent[:57] + "..."
			}

			if !strings.Contains(view, expectedContent) {
				t.Errorf("Expected view to contain formatted content for %q", tc.name)
			}
		})
	}
}

func TestManagerTimestampFormatting(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	testCases := []struct {
		name      string
		timestamp time.Time
		expected  string
	}{
		{
			"Standard timestamp",
			time.Date(2023, 10, 13, 12, 30, 45, 0, time.UTC),
			"2023-10-13 12:30:45",
		},
		{
			"Zero timestamp",
			time.Time{},
			"0001-01-01 00:00:00",
		},
		{
			"Edge of year",
			time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			"2023-12-31 23:59:59",
		},
		{
			"Beginning of year",
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			"2024-01-01 00:00:00",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			items := []history.ClipboardHistory{
				{
					Item:      "test content",
					Hash:      "hash1",
					TimeStamp: tc.timestamp,
				},
			}

			manager.UpdateRows(items)
			view := manager.View()

			if !strings.Contains(view, tc.expected) {
				t.Errorf("Expected view to contain timestamp %q, got view: %s", tc.expected, view)
			}
		})
	}
}

func TestManagerPerformanceWithManyItems(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	// Create many items to test performance
	items := make([]history.ClipboardHistory, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = history.ClipboardHistory{
			Item:      strings.Repeat("content", i%10+1),
			Hash:      string(rune(i)), // Simple hash for testing
			TimeStamp: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}

	// Should handle many items without significant performance issues
	manager.UpdateRows(items)
	view := manager.View()

	if view == "" {
		t.Error("Expected non-empty view with many items")
	}

	// Should still respond to operations
	cursor := manager.GetCursor()
	if cursor < 0 {
		t.Error("Expected valid cursor position with many items")
	}
}

func TestManagerEdgeCases(t *testing.T) {
	theme := styles.DefaultTableTheme()
	manager := NewManager(theme)

	t.Run("Empty content items", func(t *testing.T) {
		items := []history.ClipboardHistory{
			{Item: "", Hash: "hash1", TimeStamp: time.Now()},
			{Item: "   ", Hash: "hash2", TimeStamp: time.Now()},    // whitespace only
			{Item: "\n\n\n", Hash: "hash3", TimeStamp: time.Now()}, // newlines only
		}

		manager.UpdateRows(items)
		view := manager.View()

		// Should handle empty content gracefully
		if view == "" {
			t.Error("Expected non-empty view even with empty content items")
		}

		// Empty content should show as empty in table, whitespace should show as spaces
		if !strings.Contains(view, "   ") { // Whitespace preserved as spaces
			t.Error("Expected whitespace-only content to be preserved as spaces")
		}
	})

	t.Run("Nil slice", func(t *testing.T) {
		// Should handle nil slice gracefully
		manager.UpdateRows(nil)
		view := manager.View()

		if view == "" {
			t.Error("Expected non-empty view even with nil items")
		}
	})

	t.Run("Very large single item", func(t *testing.T) {
		largeContent := strings.Repeat("a", 10000) // 10KB of content
		items := []history.ClipboardHistory{
			{Item: largeContent, Hash: "hash1", TimeStamp: time.Now()},
		}

		manager.UpdateRows(items)
		view := manager.View()

		// Should truncate properly
		expectedTruncated := strings.Repeat("a", 57) + "..."
		if !strings.Contains(view, expectedTruncated) {
			t.Error("Expected large content to be truncated properly")
		}
	})
}

func TestManagerZeroValue(t *testing.T) {
	// Test behavior with zero-value manager (should not panic)
	var manager Manager

	// These operations should not panic even with zero-value manager
	cursor := manager.GetCursor()
	if cursor != 0 {
		t.Errorf("Expected zero-value manager cursor to be 0, got %d", cursor)
	}

	view := manager.View()
	// Zero-value manager might return empty view or minimal view
	// The important thing is it doesn't panic
	_ = view

	// SetSize should not panic
	manager.SetSize(80, 24)

	// GetTable should not panic
	table := manager.GetTable()
	_ = table
}
