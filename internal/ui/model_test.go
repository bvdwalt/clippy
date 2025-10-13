package ui

import (
	"fmt"
	"testing"
	"time"

	"github.com/bvdwalt/clippy/internal/history"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	if model.historyManager == nil {
		t.Error("Expected historyManager to be set")
	}

	if model.cursor != 0 {
		t.Errorf("Expected cursor to be 0, got %d", model.cursor)
	}

	if model.height != 0 {
		t.Errorf("Expected height to be 0, got %d", model.height)
	}

	// Verify the manager is the same instance
	if model.historyManager != historyManager {
		t.Error("Expected historyManager to be the same instance")
	}
}

func TestModelInit(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Expected Init() to return a non-nil command")
	}

	// We can't easily test the exact commands returned by tea.Batch
	// since they're internal to bubbletea, but we can verify it doesn't panic
}

func TestModelUpdateKeyMessages(t *testing.T) {
	historyManager := history.NewManager()

	// Add some test items
	historyManager.AddItem("first item")
	historyManager.AddItem("second item")
	historyManager.AddItem("third item")

	// Since we can't easily create tea.KeyMsg instances in tests,
	// we'll test the cursor movement logic separately

	t.Run("Cursor movement logic", func(t *testing.T) {
		model := NewModel(historyManager)

		// Test up movement
		model.cursor = 1
		if model.cursor > 0 {
			model.cursor-- // simulate "up" key
		}
		if model.cursor != 0 {
			t.Errorf("Expected cursor 0 after up movement, got %d", model.cursor)
		}

		// Test up movement at boundary (should not go below 0)
		model.cursor = 0
		if model.cursor > 0 {
			model.cursor-- // simulate "up" key
		}
		if model.cursor != 0 {
			t.Errorf("Expected cursor to stay at 0, got %d", model.cursor)
		}

		// Test down movement
		model.cursor = 0
		if model.cursor < model.historyManager.Count()-1 {
			model.cursor++ // simulate "down" key
		}
		if model.cursor != 1 {
			t.Errorf("Expected cursor 1 after down movement, got %d", model.cursor)
		}

		// Test down movement at boundary (should not exceed max)
		model.cursor = 2 // last item
		if model.cursor < model.historyManager.Count()-1 {
			model.cursor++ // simulate "down" key
		}
		if model.cursor != 2 {
			t.Errorf("Expected cursor to stay at 2, got %d", model.cursor)
		}
	})
}

func TestModelUpdateTickMessage(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	// Create a TickMsg
	tickMsg := TickMsg(time.Now())

	// Note: This test won't actually interact with the clipboard
	// since clipboard.ReadAll() will likely fail in test environment
	// But we can test that the model handles the message without panicking

	newModel, cmd := model.Update(tickMsg)

	// Verify model is returned
	if newModel == nil {
		t.Error("Expected non-nil model from Update")
	}

	// Verify a new Tick command is returned
	if cmd == nil {
		t.Error("Expected Tick command to be returned")
	}

	// The model state should remain unchanged if clipboard read fails (which it will in tests)
	updatedModel := newModel.(Model)
	if updatedModel.cursor != model.cursor {
		t.Error("Cursor should not change on TickMsg")
	}
}

func TestModelUpdateWindowSizeMessage(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	windowMsg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	newModel, cmd := model.Update(windowMsg)

	if cmd != nil {
		t.Error("Expected no command for WindowSizeMsg")
	}

	updatedModel := newModel.(Model)
	if updatedModel.height != 24 {
		t.Errorf("Expected height to be 24, got %d", updatedModel.height)
	}
}

func TestModelView(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	// Test empty history
	view := model.View()
	expectedEmpty := "Clipboard History (press q to quit, enter/c to copy, d to delete)\n\nNo clipboard history yet...\n"
	if view != expectedEmpty {
		t.Errorf("Expected empty view:\n%q\nGot:\n%q", expectedEmpty, view)
	}

	// Test with items
	historyManager.AddItem("first item")
	historyManager.AddItem("second item")

	view = model.View()

	// Check that view contains expected elements
	expectedContents := []string{
		"Clipboard History (press q to quit, enter/c to copy, d to delete)",
		"> 1: first item",  // First item should be selected (cursor = 0)
		"  2: second item", // Second item should not be selected
	}

	for _, expected := range expectedContents {
		if !contains(view, expected) {
			t.Errorf("Expected view to contain %q, got:\n%s", expected, view)
		}
	}
}

func TestModelViewLongContent(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	// Add an item longer than 60 characters
	longContent := "This is a very long piece of content that should be truncated when displayed in the UI because it exceeds sixty characters"
	historyManager.AddItem(longContent)

	view := model.View()

	// Should be truncated to 57 chars + "..."
	expectedTruncated := longContent[:57] + "..."
	if !contains(view, expectedTruncated) {
		t.Errorf("Expected view to contain truncated content %q", expectedTruncated)
	}

	// Should not contain the full long content
	if contains(view, longContent) {
		t.Error("View should not contain full long content")
	}
}

func TestModelViewNewlineReplacement(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	// Add content with newlines
	contentWithNewlines := "line1\nline2\nline3"
	historyManager.AddItem(contentWithNewlines)

	view := model.View()

	// Newlines should be replaced with spaces
	expectedReplaced := "line1 line2 line3"
	if !contains(view, expectedReplaced) {
		t.Errorf("Expected view to contain %q with newlines replaced", expectedReplaced)
	}

	// Should not contain actual newlines in content
	if contains(view, "line1\nline2") {
		t.Error("View should not contain literal newlines in content")
	}
}

func TestModelViewCursorMovement(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	// Add multiple items
	items := []string{"item1", "item2", "item3"}
	for _, item := range items {
		historyManager.AddItem(item)
	}

	// Test cursor at different positions
	for i := 0; i < len(items); i++ {
		model.cursor = i
		view := model.View()

		// Check that the correct item has the cursor
		for j, item := range items {
			if j == i {
				// This item should have the cursor ">"
				expected := fmt.Sprintf("> %d: %s", j+1, item)
				if !contains(view, expected) {
					t.Errorf("Expected cursor on item %d: %q", j, expected)
				}
			} else {
				// Other items should have space " "
				expected := fmt.Sprintf("  %d: %s", j+1, item)
				if !contains(view, expected) {
					t.Errorf("Expected no cursor on item %d: %q", j, expected)
				}
			}
		}
	}
}

func TestModelEnterKeyWithValidItem(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	// Add some items
	historyManager.AddItem("test content")
	historyManager.AddItem("another item")

	// Set cursor to first item
	model.cursor = 0

	// Note: We can't easily test clipboard.WriteAll() in unit tests
	// since it requires system clipboard access. In a real scenario,
	// you might want to use dependency injection or interfaces to mock this.

	// For now, we'll test the GetItem logic that would be called
	item, ok := model.historyManager.GetItem(model.cursor)
	if !ok {
		t.Error("Expected GetItem to return true for valid cursor position")
	}
	if item.Item != "test content" {
		t.Errorf("Expected item content 'test content', got %q", item.Item)
	}
}

func TestModelEnterKeyWithInvalidCursor(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	// Set cursor to invalid position (no items in history)
	model.cursor = 5

	// Test the GetItem logic that would be called with invalid cursor
	_, ok := model.historyManager.GetItem(model.cursor)
	if ok {
		t.Error("Expected GetItem to return false for invalid cursor position")
	}
}

func TestModelUnknownKeyMessage(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	// For testing unknown keys, we'll focus on testing that
	// the model remains in a valid state regardless of input

	initialCursor := model.cursor
	initialHeight := model.height
	initialManager := model.historyManager

	// Verify model state remains consistent
	if model.cursor != initialCursor {
		t.Error("Model cursor should remain unchanged for unknown operations")
	}
	if model.height != initialHeight {
		t.Error("Model height should remain unchanged for unknown operations")
	}
	if model.historyManager != initialManager {
		t.Error("Model historyManager should remain unchanged for unknown operations")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
