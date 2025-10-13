package history

import (
	"encoding/json"
	"testing"
	"time"
)

func TestClipboardHistoryJSONSerialization(t *testing.T) {
	// Test basic JSON serialization
	original := ClipboardHistory{
		Item:      "test content",
		Hash:      "abc123",
		TimeStamp: time.Date(2023, 10, 13, 12, 0, 0, 0, time.UTC),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal ClipboardHistory: %v", err)
	}

	// Unmarshal from JSON
	var restored ClipboardHistory
	err = json.Unmarshal(jsonData, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal ClipboardHistory: %v", err)
	}

	// Verify all fields are preserved
	if original.Item != restored.Item {
		t.Errorf("Item field not preserved: expected %q, got %q", original.Item, restored.Item)
	}
	if original.Hash != restored.Hash {
		t.Errorf("Hash field not preserved: expected %q, got %q", original.Hash, restored.Hash)
	}
	if !original.TimeStamp.Equal(restored.TimeStamp) {
		t.Errorf("TimeStamp field not preserved: expected %v, got %v", original.TimeStamp, restored.TimeStamp)
	}
}

func TestClipboardHistoryJSONFieldNames(t *testing.T) {
	// Test that JSON field names are as expected
	ch := ClipboardHistory{
		Item:      "content",
		Hash:      "hash123",
		TimeStamp: time.Date(2023, 10, 13, 12, 0, 0, 0, time.UTC),
	}

	jsonData, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(jsonData)

	// Check that JSON uses the expected field names
	expectedFields := []string{`"item":`, `"hash":`, `"timeStamp":`}
	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("Expected JSON to contain field %s, got: %s", field, jsonStr)
		}
	}
}

func TestClipboardHistoryEmptyValues(t *testing.T) {
	// Test serialization with empty/zero values
	ch := ClipboardHistory{
		Item:      "",
		Hash:      "",
		TimeStamp: time.Time{},
	}

	jsonData, err := json.Marshal(ch)
	if err != nil {
		t.Fatalf("Failed to marshal empty ClipboardHistory: %v", err)
	}

	var restored ClipboardHistory
	err = json.Unmarshal(jsonData, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal empty ClipboardHistory: %v", err)
	}

	if restored.Item != "" {
		t.Errorf("Expected empty Item, got %q", restored.Item)
	}
	if restored.Hash != "" {
		t.Errorf("Expected empty Hash, got %q", restored.Hash)
	}
	if !restored.TimeStamp.IsZero() {
		t.Errorf("Expected zero TimeStamp, got %v", restored.TimeStamp)
	}
}

func TestClipboardHistorySpecialCharacters(t *testing.T) {
	// Test with various special characters that might cause JSON issues
	testCases := []struct {
		name    string
		content string
	}{
		{"Newlines", "line1\nline2\nline3"},
		{"Tabs", "col1\tcol2\tcol3"},
		{"Quotes", `text with "quotes" and 'apostrophes'`},
		{"Backslashes", `path\to\file and \\network\share`},
		{"Unicode", "Unicode: 你好, مرحبا, こんにちは, 🚀🎉"},
		{"JSON-like", `{"key": "value", "number": 123}`},
		{"HTML", "<div>HTML content with &amp; entities</div>"},
		{"Mixed", "Mixed content\nwith \"quotes\"\tand\tunicode: 🎯"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			original := ClipboardHistory{
				Item:      tc.content,
				Hash:      "hash123",
				TimeStamp: time.Now(),
			}

			jsonData, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Failed to marshal content %q: %v", tc.content, err)
			}

			var restored ClipboardHistory
			err = json.Unmarshal(jsonData, &restored)
			if err != nil {
				t.Fatalf("Failed to unmarshal content %q: %v", tc.content, err)
			}

			if original.Item != restored.Item {
				t.Errorf("Content not preserved for %s:\nExpected: %q\nGot: %q", tc.name, original.Item, restored.Item)
			}
		})
	}
}

func TestClipboardHistoryLargeContent(t *testing.T) {
	// Test with large content
	largeContent := string(make([]byte, 10000)) // 10KB of null bytes
	for i := range largeContent {
		largeContent = largeContent[:i] + "A" + largeContent[i+1:]
	}

	original := ClipboardHistory{
		Item:      largeContent,
		Hash:      "largehash",
		TimeStamp: time.Now(),
	}

	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal large content: %v", err)
	}

	var restored ClipboardHistory
	err = json.Unmarshal(jsonData, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal large content: %v", err)
	}

	if len(restored.Item) != len(original.Item) {
		t.Errorf("Large content length mismatch: expected %d, got %d", len(original.Item), len(restored.Item))
	}
	if original.Item != restored.Item {
		t.Error("Large content not preserved correctly")
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
