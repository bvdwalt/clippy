package ui

import (
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

	if model.GetCursor() != 0 {
		t.Errorf("Expected cursor to be 0, got %d", model.GetCursor())
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

		// With the new table-based model, we test cursor position after simulated key events
		// Test that cursor starts at 0
		if model.GetCursor() != 0 {
			t.Errorf("Expected initial cursor 0, got %d", model.GetCursor())
		}

		// Simulate down key press to move cursor
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ := model.Update(downMsg)
		model = updatedModel.(Model)

		if model.GetCursor() != 1 {
			t.Errorf("Expected cursor 1 after down movement, got %d", model.GetCursor())
		}

		// Simulate up key press to move cursor back
		upMsg := tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, _ = model.Update(upMsg)
		model = updatedModel.(Model)

		if model.GetCursor() != 0 {
			t.Errorf("Expected cursor 0 after up movement, got %d", model.GetCursor())
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
	if updatedModel.GetCursor() != model.GetCursor() {
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
	if !contains(view, "No clipboard history yet...") {
		t.Errorf("Expected empty view to contain 'No clipboard history yet...', got:\n%s", view)
	}

	// Test with items
	historyManager.AddItem("first item")
	historyManager.AddItem("second item")
	model.UpdateTable() // Update table with new items

	view = model.View()

	// Check that view contains expected elements (table format)
	expectedContents := []string{
		"📋 Clippy Clipboard History",
		"first item",
		"second item",
		"Total items: 2",
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
	model.UpdateTable() // Update table with new items

	view := model.View()

	// Should be truncated to 57 chars + "..." (content longer than 60 chars)
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
	model.UpdateTable() // Update table with new items

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
	model.UpdateTable() // Update table with new items

	// Test that view renders with items
	view := model.View()

	// Check that all items appear in the table
	for _, item := range items {
		if !contains(view, item) {
			t.Errorf("Expected view to contain item %q", item)
		}
	}

	// Test cursor movement through key events
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := model.Update(downMsg)
	model = updatedModel.(Model)

	// Verify cursor moved
	if model.GetCursor() == 0 && len(items) > 1 {
		t.Error("Expected cursor to move down from initial position")
	}
}

func TestModelEnterKeyWithValidItem(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	// Add some items
	historyManager.AddItem("test content")
	historyManager.AddItem("another item")
	model.UpdateTable() // Update table with new items

	// Note: We can't easily test clipboard.WriteAll() in unit tests
	// since it requires system clipboard access. In a real scenario,
	// you might want to use dependency injection or interfaces to mock this.

	// Test that GetItem works with current cursor position
	cursor := model.GetCursor()
	item, ok := model.historyManager.GetItem(cursor)
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

	// Test the GetItem logic with invalid cursor (empty history)
	_, ok := model.historyManager.GetItem(5)
	if ok {
		t.Error("Expected GetItem to return false for invalid cursor position")
	}
}

func TestModelUnknownKeyMessage(t *testing.T) {
	historyManager := history.NewManager()
	model := NewModel(historyManager)

	// For testing unknown keys, we'll focus on testing that
	// the model remains in a valid state regardless of input

	initialHeight := model.height
	initialManager := model.historyManager

	// Test with an unknown key message
	unknownMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	updatedModel, _ := model.Update(unknownMsg)
	model = updatedModel.(Model)

	// Verify model state remains consistent
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
