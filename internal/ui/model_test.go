package ui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestNewModel(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
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
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Expected Init() to return a non-nil command")
	}

	// We can't easily test the exact commands returned by tea.Batch
	// since they're internal to bubbletea, but we can verify it doesn't panic
}

func TestModelUpdateKeyMessages(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

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
		downMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
		updatedModel, _ := model.Update(downMsg)
		model = updatedModel.(Model)

		if model.GetCursor() != 1 {
			t.Errorf("Expected cursor 1 after down movement, got %d", model.GetCursor())
		}

		// Simulate up key press to move cursor back
		upMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
		updatedModel, _ = model.Update(upMsg)
		model = updatedModel.(Model)

		if model.GetCursor() != 0 {
			t.Errorf("Expected cursor 0 after up movement, got %d", model.GetCursor())
		}
	})
}

func TestModelUpdateTickMessage(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
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
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
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
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	// Test empty history
	view := model.View()
	viewStr := view.Content
	if !contains(viewStr, "No clipboard history yet...") {
		t.Errorf("Expected empty view to contain 'No clipboard history yet...', got:\n%s", viewStr)
	}

	// Test with items
	historyManager.AddItem("first item")
	historyManager.AddItem("second item")
	model.UpdateTable() // Update table with new items

	view = model.View()
	viewStr = view.Content

	// Check that view contains expected elements (table format)
	expectedContents := []string{
		"📋 Clippy Clipboard History",
		"first item",
		"second item",
		"Total items: 2",
	}

	for _, expected := range expectedContents {
		if !contains(viewStr, expected) {
			t.Errorf("Expected view to contain %q, got:\n%s", expected, viewStr)
		}
	}
}

func TestModelViewLongContent(t *testing.T) {
	// ...existing code, but change view uses to view.Content
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	longContent := "This is a very long piece of content that should be truncated when displayed in the UI because it exceeds sixty characters"
	historyManager.AddItem(longContent)
	model.UpdateTable()

	view := model.View()
	viewStr := view.Content

	expectedTruncated := longContent[:57] + "..."
	if !contains(viewStr, expectedTruncated) {
		t.Errorf("Expected view to contain truncated content %q", expectedTruncated)
	}

	if contains(viewStr, longContent) {
		t.Error("View should not contain full long content")
	}
}

func TestModelViewNewlineReplacement(t *testing.T) {
	// ...existing code with view.Content
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	contentWithNewlines := "line1\nline2\nline3"
	historyManager.AddItem(contentWithNewlines)
	model.UpdateTable()

	view := model.View()
	viewStr := view.Content

	expectedReplaced := "line1 line2 line3"
	if !contains(viewStr, expectedReplaced) {
		t.Errorf("Expected view to contain %q with newlines replaced", expectedReplaced)
	}

	if contains(viewStr, "line1\nline2") {
		t.Error("View should not contain literal newlines in content")
	}
}

func TestModelViewCursorMovement(t *testing.T) {
	// ...existing code, change view to view.Content
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	items := []string{"item1", "item2", "item3"}
	for _, item := range items {
		historyManager.AddItem(item)
	}
	model.UpdateTable()

	view := model.View()
	viewStr := view.Content

	for _, item := range items {
		if !contains(viewStr, item) {
			t.Errorf("Expected view to contain item %q", item)
		}
	}

	// Test cursor movement through key events
	downMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	updatedModel, _ := model.Update(downMsg)
	model = updatedModel.(Model)

	if model.GetCursor() == 0 && len(items) > 1 {
		t.Error("Expected cursor to move down from initial position")
	}
}

func TestModelEnterKeyWithValidItem(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
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
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	// Test the GetItem logic with invalid cursor (empty history)
	_, ok := model.historyManager.GetItem(5)
	if ok {
		t.Error("Expected GetItem to return false for invalid cursor position")
	}
}

func TestModelPreviewHeight(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	windowMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(Model)

	// available = max(40-10, 6) = 30, previewH = max(30/3, 3) = 10
	if model.previewHeight != 10 {
		t.Errorf("Expected previewHeight 10, got %d", model.previewHeight)
	}
}

func TestModelPreviewHeightSmallWindow(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	// Very small window: available = max(12-10, 6) = 6, previewH = max(6/3, 3) = 3
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 12}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(Model)

	if model.previewHeight != 3 {
		t.Errorf("Expected minimum previewHeight 3, got %d", model.previewHeight)
	}
}

func TestModelPreviewPane(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("preview content here")
	model := NewModel(historyManager)

	// Trigger window resize to set previewHeight > 0
	windowMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(Model)

	view := model.View()

	if !contains(view, "Preview") {
		t.Error("Expected view to contain 'Preview' label")
	}
	if !contains(view, "preview content here") {
		t.Error("Expected view to contain selected item content in preview pane")
	}
}

func TestModelPreviewPaneNoSelection(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	windowMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(Model)

	// No items — preview pane should still render without panicking
	view := model.View()
	if !contains(view, "Preview") {
		t.Error("Expected 'Preview' label even with no items")
	}
}

func TestModelUnknownKeyMessage(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	initialHeight := model.height
	initialManager := model.historyManager

	// Test with an unknown key message (use KeyPressMsg with Text)
	unknownMsg := tea.KeyPressMsg(tea.Key{Text: "x"})
	updatedModel, _ := model.Update(unknownMsg)
	model = updatedModel.(Model)

	if model.height != initialHeight {
		t.Error("Model height should remain unchanged for unknown operations")
	}
	if model.historyManager != initialManager {
		t.Error("Model historyManager should remain unchanged for unknown operations")
	}
}

func TestModelFilterItems(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("hello world")
	historyManager.AddItem("foo bar")
	historyManager.AddItem("hello go")

	t.Run("Non-empty query sets filtered", func(t *testing.T) {
		model := NewModel(historyManager)
		model.filterItems("hello")

		if model.filtered == nil {
			t.Fatal("Expected filtered to be set after non-empty query")
		}
		for _, item := range model.filtered {
			if !contains(item.Item, "hello") {
				t.Errorf("Unexpected item in filtered results: %q", item.Item)
			}
		}
	})

	t.Run("Empty query clears filtered", func(t *testing.T) {
		model := NewModel(historyManager)
		model.filterItems("hello")
		model.filterItems("")

		if model.filtered != nil {
			t.Error("Expected filtered to be nil after empty query")
		}
	})
}

func TestModelGetDisplayItems(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("item one")
	historyManager.AddItem("item two")

	t.Run("Returns all items when no filter", func(t *testing.T) {
		model := NewModel(historyManager)
		items := model.getDisplayItems()
		if len(items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(items))
		}
	})

	t.Run("Returns filtered items when filter set", func(t *testing.T) {
		model := NewModel(historyManager)
		model.filterItems("one")
		items := model.getDisplayItems()
		if len(items) != 1 {
			t.Errorf("Expected 1 filtered item, got %d", len(items))
		}
		if items[0].Item != "item one" {
			t.Errorf("Expected 'item one', got %q", items[0].Item)
		}
	})
}

func TestModelSearchModeToggle(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	if model.mode != TableView {
		t.Fatal("Expected initial mode to be TableView")
	}

	// Press "/" to enter search mode
	slashMsg := tea.KeyPressMsg(tea.Key{Text: "/"})
	newModel, _ := model.Update(slashMsg)
	model = newModel.(Model)

	if model.mode != SearchView {
		t.Error("Expected mode to be SearchView after pressing '/'")
	}

	// Press "esc" to exit search mode
	escMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	newModel, _ = model.Update(escMsg)
	model = newModel.(Model)

	if model.mode != TableView {
		t.Error("Expected mode to return to TableView after pressing 'esc'")
	}
	if model.filtered != nil {
		t.Error("Expected filter to be cleared after pressing 'esc'")
	}
}

func TestModelSearchModeTyping(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	// Enter search mode
	slashMsg := tea.KeyPressMsg(tea.Key{Text: "/"})
	newModel, _ := model.Update(slashMsg)
	model = newModel.(Model)

	// Type a character — should be handled by textInput
	aMsg := tea.KeyPressMsg(tea.Key{Text: "a"})
	newModel, _ = model.Update(aMsg)
	model = newModel.(Model)

	if model.mode != SearchView {
		t.Error("Expected mode to remain SearchView while typing")
	}
}

func TestModelSearchModeEnter(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("hello world")
	historyManager.AddItem("foo bar")
	model := NewModel(historyManager)

	// Enter search mode
	newModel, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "/"}))
	model = newModel.(Model)

	// Type a search term directly into the model's textInput
	model.textInput.SetValue("hello")

	// Press enter to apply the search
	enterMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	newModel, _ = model.Update(enterMsg)
	model = newModel.(Model)

	if model.mode != TableView {
		t.Error("Expected mode to return to TableView after search enter")
	}
	if model.filtered == nil {
		t.Error("Expected filtered to be set after search enter")
	}
}

func TestModelQuitKey(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	_, cmd := model.Update(tea.KeyPressMsg(tea.Key{Text: "q"}))
	if cmd == nil {
		t.Error("Expected quit command after pressing 'q'")
	}
}

func TestModelDeleteKey(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("item to delete")
	historyManager.AddItem("item to keep")
	model := NewModel(historyManager)

	initialCount := historyManager.Count()

	dMsg := tea.KeyPressMsg(tea.Key{Text: "d"})
	newModel, _ := model.Update(dMsg)
	model = newModel.(Model)

	if historyManager.Count() != initialCount-1 {
		t.Errorf("Expected item count to decrease by 1, got %d (was %d)", historyManager.Count(), initialCount)
	}
	_ = model
}

func TestModelRefreshKey(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("some item")
	model := NewModel(historyManager)

	// Set a filter first
	model.filterItems("some")
	if model.filtered == nil {
		t.Fatal("Expected filter to be set before refresh")
	}

	rMsg := tea.KeyPressMsg(tea.Key{Text: "r"})
	newModel, _ := model.Update(rMsg)
	model = newModel.(Model)

	if model.filtered != nil {
		t.Error("Expected filter to be cleared after 'r' refresh")
	}
	if model.mode != TableView {
		t.Error("Expected mode to be TableView after refresh")
	}
}

func TestModelViewSearchMode(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)

	// Switch to search mode
	newModel, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "/"}))
	model = newModel.(Model)

	view := model.View()
	if !contains(view, "Search") {
		t.Error("Expected SearchView to contain 'Search'")
	}
	// Table should not be rendered in search mode
	if contains(view, "Total items") {
		t.Error("Expected table status line to be absent in SearchView")
	}
}

func TestModelViewFilteredStatus(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("hello world")
	historyManager.AddItem("hello go")
	historyManager.AddItem("foo bar")
	model := NewModel(historyManager)

	// Apply a filter directly
	model.filterItems("hello")
	model.updateTable()

	view := model.View()
	if !contains(view, "Showing") {
		t.Error("Expected filtered view to show 'Showing X of Y items'")
	}
}

func TestModelViewNoResultsMessage(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("hello world")
	model := NewModel(historyManager)

	// Apply a filter that matches nothing
	model.filterItems("zzznomatch")
	model.updateTable()

	view := model.View()
	if !contains(view, "No results found") {
		t.Error("Expected 'No results found' when filter matches nothing")
	}
}

// Helper function to check if a string (or tea.View) contains a substring
func contains(hay any, substr string) bool {
	var s string
	switch v := hay.(type) {
	case string:
		s = v
	case tea.View:
		s = v.Content
	default:
		return false
	}

	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	// Strip ANSI escape sequences (lipgloss output) so tests can match plain text
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	s = re.ReplaceAllString(s, "")

	return strings.Contains(s, substr)
}

func TestModelPinKey(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("item to pin")
	model := NewModel(historyManager)

	newModel, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "p"}))
	model = newModel.(Model)

	item, ok := historyManager.GetItem(0)
	if !ok {
		t.Fatal("expected item at index 0")
	}
	if !item.Pinned {
		t.Error("expected item to be pinned after 'p'")
	}

	// Toggle off
	newModel, _ = model.Update(tea.KeyPressMsg(tea.Key{Text: "p"}))
	model = newModel.(Model)
	item, _ = historyManager.GetItem(0)
	if item.Pinned {
		t.Error("expected item to be unpinned after second 'p'")
	}
	_ = model
}

func TestModelDeletePinnedItemConfirmY(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("pinned item")
	if err := historyManager.TogglePin(0); err != nil {
		t.Fatalf("TogglePin: %v", err)
	}
	model := NewModel(historyManager)

	// 'd' on a pinned item should set confirmDelete, not delete immediately
	newModel, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "d"}))
	model = newModel.(Model)

	if historyManager.Count() != 1 {
		t.Error("expected item to still exist after 'd' on pinned item")
	}
	if !model.confirmDelete {
		t.Error("expected confirmDelete to be true")
	}

	// View should show the confirmation prompt
	if !contains(model.View(), "y/n") {
		t.Error("expected confirmation prompt in view")
	}

	// 'y' should confirm the delete
	newModel, _ = model.Update(tea.KeyPressMsg(tea.Key{Text: "y"}))
	model = newModel.(Model)

	if historyManager.Count() != 0 {
		t.Error("expected item to be deleted after 'y' confirmation")
	}
	if model.confirmDelete {
		t.Error("expected confirmDelete to be cleared after 'y'")
	}
	_ = model
}

func TestModelDeletePinnedItemConfirmN(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("pinned item")
	if err := historyManager.TogglePin(0); err != nil {
		t.Fatalf("TogglePin: %v", err)
	}
	model := NewModel(historyManager)

	newModel, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "d"}))
	model = newModel.(Model)

	// 'n' should cancel
	newModel, _ = model.Update(tea.KeyPressMsg(tea.Key{Text: "n"}))
	model = newModel.(Model)

	if historyManager.Count() != 1 {
		t.Error("expected item to still exist after 'n' cancel")
	}
	if model.confirmDelete {
		t.Error("expected confirmDelete to be cleared after 'n'")
	}
	_ = model
}

func TestModelDeletePinnedItemConfirmEsc(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()

	historyManager.AddItem("pinned item")
	if err := historyManager.TogglePin(0); err != nil {
		t.Fatalf("TogglePin: %v", err)
	}
	model := NewModel(historyManager)

	newModel, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "d"}))
	model = newModel.(Model)

	newModel, _ = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	model = newModel.(Model)

	if historyManager.Count() != 1 {
		t.Error("expected item to still exist after esc cancel")
	}
	if model.confirmDelete {
		t.Error("expected confirmDelete to be cleared after esc")
	}
	_ = model
}
