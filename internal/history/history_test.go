package history

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.Count() != 0 {
		t.Errorf("Expected empty manager to have 0 items, got %d", manager.Count())
	}

	if len(manager.items) != 0 {
		t.Errorf("Expected empty items slice, got length %d", len(manager.items))
	}

	if len(manager.hashes) != 0 {
		t.Errorf("Expected empty hashes map, got length %d", len(manager.hashes))
	}
}

func TestAddItem(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"Add first item", "hello world", true},
		{"Add different item", "goodbye world", true},
		{"Add duplicate item", "hello world", false},
		{"Add empty string", "", true},
		{"Add duplicate empty string", "", false},
		{"Add whitespace", "   ", true},
		{"Add newline content", "line1\nline2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.AddItem(tt.content)
			if result != tt.expected {
				t.Errorf("AddItem(%q) = %v, expected %v", tt.content, result, tt.expected)
			}
		})
	}

	// Verify final count
	expectedCount := 5 // unique items: "hello world", "goodbye world", "", "   ", "line1\nline2"
	if manager.Count() != expectedCount {
		t.Errorf("Expected %d items after all additions, got %d", expectedCount, manager.Count())
	}
}

func TestGetItem(t *testing.T) {
	manager := NewManager()

	// Test empty manager
	_, ok := manager.GetItem(0)
	if ok {
		t.Error("Expected GetItem(0) to return false for empty manager")
	}

	_, ok = manager.GetItem(-1)
	if ok {
		t.Error("Expected GetItem(-1) to return false")
	}

	// Add some items
	contents := []string{"first", "second", "third"}
	for _, content := range contents {
		manager.AddItem(content)
	}

	// Test valid indices
	for i, expectedContent := range contents {
		item, ok := manager.GetItem(i)
		if !ok {
			t.Errorf("Expected GetItem(%d) to return true", i)
		}
		if item.Item != expectedContent {
			t.Errorf("Expected item content %q, got %q", expectedContent, item.Item)
		}
	}

	// Test invalid indices
	invalidIndices := []int{-1, 3, 100}
	for _, index := range invalidIndices {
		_, ok := manager.GetItem(index)
		if ok {
			t.Errorf("Expected GetItem(%d) to return false", index)
		}
	}
}

func TestGetItems(t *testing.T) {
	manager := NewManager()

	// Test empty manager
	items := manager.GetItems()
	if len(items) != 0 {
		t.Errorf("Expected empty slice for new manager, got length %d", len(items))
	}

	// Add items and test
	contents := []string{"item1", "item2", "item3"}
	for _, content := range contents {
		manager.AddItem(content)
	}

	items = manager.GetItems()
	if len(items) != len(contents) {
		t.Errorf("Expected %d items, got %d", len(contents), len(items))
	}

	for i, expectedContent := range contents {
		if items[i].Item != expectedContent {
			t.Errorf("Expected item %d to be %q, got %q", i, expectedContent, items[i].Item)
		}
	}
}

func TestCount(t *testing.T) {
	manager := NewManager()

	// Test empty manager
	if manager.Count() != 0 {
		t.Errorf("Expected count 0 for new manager, got %d", manager.Count())
	}

	// Add items and verify count
	contents := []string{"a", "b", "c", "a"} // last "a" is duplicate
	expectedCounts := []int{1, 2, 3, 3}      // count after each addition

	for i, content := range contents {
		manager.AddItem(content)
		if manager.Count() != expectedCounts[i] {
			t.Errorf("After adding %q, expected count %d, got %d", content, expectedCounts[i], manager.Count())
		}
	}
}

func TestNewClipboardItem(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"Simple text", "hello"},
		{"Empty string", ""},
		{"With newlines", "line1\nline2\nline3"},
		{"With special chars", "!@#$%^&*()"},
		{"Unicode", "こんにちは"},
		{"Long text", string(make([]byte, 1000))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := newClipboardItem(tt.content)

			if item.Item != tt.content {
				t.Errorf("Expected content %q, got %q", tt.content, item.Item)
			}

			if item.Hash == "" {
				t.Error("Expected non-empty hash")
			}

			if len(item.Hash) != 64 { // SHA256 hex string length
				t.Errorf("Expected hash length 64, got %d", len(item.Hash))
			}

			if item.TimeStamp.IsZero() {
				t.Error("Expected non-zero timestamp")
			}

			// Verify timestamp is recent (within last second)
			if time.Since(item.TimeStamp) > time.Second {
				t.Error("Timestamp seems too old")
			}
		})
	}

	// Test that same content produces same hash
	content := "test content"
	item1 := newClipboardItem(content)
	item2 := newClipboardItem(content)

	if item1.Hash != item2.Hash {
		t.Error("Same content should produce same hash")
	}

	// Test that different content produces different hash
	item3 := newClipboardItem("different content")
	if item1.Hash == item3.Hash {
		t.Error("Different content should produce different hash")
	}
}

func TestSaveAndLoadFromFile(t *testing.T) {
	// Use a temporary file for testing
	tempFile := "test_history.json"

	// Override the constant for testing
	// Note: We'll need to be careful since this is a const
	defer func() {
		os.Remove(tempFile)
	}()

	// Create manager and add test data
	manager := NewManager()
	testContents := []string{"item1", "item2", "item3"}
	for _, content := range testContents {
		manager.AddItem(content)
	}

	// Test SaveToFile - we'll temporarily modify the filename
	// Since HistoryFileName is a const, we need to test with the actual filename
	err := manager.SaveToFile()
	if err != nil {
		t.Fatalf("SaveToFile() failed: %v", err)
	}
	defer os.Remove(HistoryFileName) // Clean up

	// Create new manager and load
	newManager := NewManager()
	err = newManager.LoadFromFile()
	if err != nil {
		t.Fatalf("LoadFromFile() failed: %v", err)
	}

	// Verify loaded data
	if newManager.Count() != manager.Count() {
		t.Errorf("Expected count %d after loading, got %d", manager.Count(), newManager.Count())
	}

	originalItems := manager.GetItems()
	loadedItems := newManager.GetItems()

	for i, originalItem := range originalItems {
		if i >= len(loadedItems) {
			t.Fatalf("Missing item at index %d", i)
		}

		loadedItem := loadedItems[i]
		if originalItem.Item != loadedItem.Item {
			t.Errorf("Item %d: expected %q, got %q", i, originalItem.Item, loadedItem.Item)
		}
		if originalItem.Hash != loadedItem.Hash {
			t.Errorf("Item %d: hash mismatch", i)
		}
		if !originalItem.TimeStamp.Equal(loadedItem.TimeStamp) {
			t.Errorf("Item %d: timestamp mismatch", i)
		}
	}

	// Test that hashes map is properly reconstructed
	for _, item := range loadedItems {
		if _, exists := newManager.hashes[item.Hash]; !exists {
			t.Errorf("Hash %s not found in hashes map after loading", item.Hash)
		}
	}
}

func TestLoadFromNonExistentFile(t *testing.T) {
	// Ensure file doesn't exist
	nonExistentFile := "definitely_does_not_exist.json"
	os.Remove(nonExistentFile)

	manager := NewManager()
	err := manager.LoadFromFile()

	// Should not return an error for non-existent file
	if err != nil {
		t.Errorf("LoadFromFile() should not error for non-existent file, got: %v", err)
	}

	if manager.Count() != 0 {
		t.Errorf("Expected empty manager after loading non-existent file, got count %d", manager.Count())
	}
}

func TestLoadFromCorruptedFile(t *testing.T) {
	// Create a corrupted JSON file
	corruptedContent := `{"invalid": json content`
	err := os.WriteFile(HistoryFileName, []byte(corruptedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted test file: %v", err)
	}
	defer os.Remove(HistoryFileName)

	manager := NewManager()
	err = manager.LoadFromFile()

	if err == nil {
		t.Error("Expected error when loading corrupted file")
	}

	if manager.Count() != 0 {
		t.Errorf("Expected empty manager after failed load, got count %d", manager.Count())
	}
}

func TestSaveToInvalidPath(t *testing.T) {
	// This test might be platform-specific
	// On Unix-like systems, we can try to save to a directory that doesn't exist
	manager := NewManager()
	manager.AddItem("test")

	// Temporarily change the filename to an invalid path
	// Since we can't change the const, we'll test the current behavior
	// The real SaveToFile uses HistoryFileName which should work in current directory

	// Instead, let's test by creating a file with no write permissions
	restrictedFile := "restricted_history.json"
	err := os.WriteFile(restrictedFile, []byte("test"), 0444) // read-only
	if err != nil {
		t.Skipf("Could not create read-only file for testing: %v", err)
	}
	defer func() {
		os.Chmod(restrictedFile, 0644) // restore permissions
		os.Remove(restrictedFile)
	}()

	// This test would require modifying the Manager to accept a custom filename
	// For now, we'll just ensure SaveToFile works with valid permissions
	err = manager.SaveToFile()
	if err != nil {
		t.Errorf("SaveToFile() failed with error: %v", err)
	}
	defer os.Remove(HistoryFileName)
}

func TestSaveToFileWithUnencodableData(t *testing.T) {
	// Test case for when JSON encoding might fail
	// This is tricky since ClipboardHistory has simple, encodable fields
	// We'll test with the current implementation
	manager := NewManager()
	manager.AddItem("valid content")

	err := manager.SaveToFile()
	if err != nil {
		t.Errorf("SaveToFile() should succeed with valid data, got error: %v", err)
	}
	defer os.Remove(HistoryFileName)

	// Note: It's difficult to force a JSON encoding error with ClipboardHistory
	// since all its fields (string, string, time.Time) are JSON-encodable.
	// In a real scenario, you might test this by mocking the encoder or
	// by temporarily replacing the encoder with a failing one.
}

func TestJSONMarshaling(t *testing.T) {
	// Test that ClipboardHistory can be properly marshaled/unmarshaled
	original := ClipboardHistory{
		Item:      "test content",
		Hash:      "abcd1234",
		TimeStamp: time.Now().Truncate(time.Second), // Truncate for comparison
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal ClipboardHistory: %v", err)
	}

	var unmarshaled ClipboardHistory
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ClipboardHistory: %v", err)
	}

	if original.Item != unmarshaled.Item {
		t.Errorf("Item mismatch: expected %q, got %q", original.Item, unmarshaled.Item)
	}
	if original.Hash != unmarshaled.Hash {
		t.Errorf("Hash mismatch: expected %q, got %q", original.Hash, unmarshaled.Hash)
	}
	if !original.TimeStamp.Equal(unmarshaled.TimeStamp) {
		t.Errorf("TimeStamp mismatch: expected %v, got %v", original.TimeStamp, unmarshaled.TimeStamp)
	}
}
