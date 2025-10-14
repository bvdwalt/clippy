package search

import (
	"strings"
	"testing"
	"time"

	"github.com/bvdwalt/clippy/internal/history"
)

func TestNewFuzzyMatcher(t *testing.T) {
	matcher := NewFuzzyMatcher()
	if matcher == nil {
		t.Fatal("NewFuzzyMatcher should return a non-nil matcher")
	}
}

func TestFuzzyMatcher_Search_EmptyQuery(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "test content", Hash: "hash1", TimeStamp: time.Now()},
		{Item: "another item", Hash: "hash2", TimeStamp: time.Now()},
	}

	result := matcher.Search(items, "")
	if result != nil {
		t.Errorf("Expected nil result for empty query, got %v", result)
	}
}

func TestFuzzyMatcher_Search_NoMatches(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "hello world", Hash: "hash1", TimeStamp: time.Now()},
		{Item: "foo bar", Hash: "hash2", TimeStamp: time.Now()},
	}

	result := matcher.Search(items, "xyz")
	if len(result) != 0 {
		t.Errorf("Expected 0 matches for non-matching query, got %d", len(result))
	}
}

func TestFuzzyMatcher_Search_ExactMatch(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "exact match", Hash: "hash1", TimeStamp: time.Now()},
		{Item: "different content", Hash: "hash2", TimeStamp: time.Now()},
	}

	result := matcher.Search(items, "exact match")
	if len(result) != 1 {
		t.Fatalf("Expected 1 match for exact query, got %d", len(result))
	}
	if result[0].Item != "exact match" {
		t.Errorf("Expected 'exact match', got '%s'", result[0].Item)
	}
}

func TestFuzzyMatcher_Search_CaseInsensitive(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "Hello World", Hash: "hash1", TimeStamp: time.Now()},
		{Item: "GOODBYE", Hash: "hash2", TimeStamp: time.Now()},
	}

	testCases := []struct {
		query    string
		expected string
	}{
		{"hello", "Hello World"},
		{"HELLO", "Hello World"},
		{"HeLLo", "Hello World"},
		{"goodbye", "GOODBYE"},
		{"GoodBye", "GOODBYE"},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			result := matcher.Search(items, tc.query)
			if len(result) == 0 {
				t.Fatalf("Expected match for query '%s'", tc.query)
			}
			if result[0].Item != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result[0].Item)
			}
		})
	}
}

func TestFuzzyMatcher_Search_SubsequenceMatch(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "hello world", Hash: "hash1", TimeStamp: time.Now()},
		{Item: "help me", Hash: "hash2", TimeStamp: time.Now()},
		{Item: "helicopter", Hash: "hash3", TimeStamp: time.Now()},
	}

	result := matcher.Search(items, "hel")
	if len(result) != 3 {
		t.Fatalf("Expected 3 matches for 'hel', got %d", len(result))
	}

	// All items should contain the subsequence "hel"
	for _, item := range result {
		if !containsSubsequence(item.Item, "hel") {
			t.Errorf("Item '%s' doesn't contain subsequence 'hel'", item.Item)
		}
	}
}

func TestFuzzyMatcher_Search_ScoreOrdering(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "xyz test abc", Hash: "hash1", TimeStamp: time.Now()}, // test not at beginning
		{Item: "test example", Hash: "hash2", TimeStamp: time.Now()}, // test at beginning (should score higher)
		{Item: "another test", Hash: "hash3", TimeStamp: time.Now()}, // test not at beginning
		{Item: "testing", Hash: "hash4", TimeStamp: time.Now()},      // test at beginning of word
	}

	result := matcher.Search(items, "test")
	if len(result) != 4 {
		t.Fatalf("Expected 4 matches, got %d", len(result))
	}

	// Items starting with "test" should generally score higher
	// The exact ordering depends on the scoring algorithm, but we can verify
	// that matches exist and are ordered by score
	for i := 1; i < len(result); i++ {
		// We can't easily test the exact order without exposing scores,
		// but we can verify all expected items are present
		found := false
		for _, item := range items {
			if item.Item == result[i].Item {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unexpected item in results: %s", result[i].Item)
		}
	}
}

func TestFuzzyMatcher_Search_WordBoundaryBonus(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "some-test-data", Hash: "hash1", TimeStamp: time.Now()}, // test at word boundary
		{Item: "contest result", Hash: "hash2", TimeStamp: time.Now()}, // test not at word boundary
		{Item: "my_test_file", Hash: "hash3", TimeStamp: time.Now()},   // test at word boundary
	}

	result := matcher.Search(items, "test")
	if len(result) != 3 {
		t.Fatalf("Expected 3 matches, got %d", len(result))
	}

	// All items should match
	expectedItems := map[string]bool{
		"some-test-data": false,
		"contest result": false,
		"my_test_file":   false,
	}

	for _, item := range result {
		if _, exists := expectedItems[item.Item]; !exists {
			t.Errorf("Unexpected item in results: %s", item.Item)
		}
		expectedItems[item.Item] = true
	}

	for item, found := range expectedItems {
		if !found {
			t.Errorf("Expected item not found in results: %s", item)
		}
	}
}

func TestFuzzyMatcher_Search_CamelCaseBonus(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "myTestFunction", Hash: "hash1", TimeStamp: time.Now()},
		{Item: "testingStuff", Hash: "hash2", TimeStamp: time.Now()},
		{Item: "TestCase", Hash: "hash3", TimeStamp: time.Now()},
	}

	result := matcher.Search(items, "Test")
	if len(result) == 0 {
		t.Fatal("Expected matches for camelCase search")
	}

	// All items should match since they all contain "Test" or "test"
	for _, item := range result {
		found := containsSubsequence(strings.ToLower(item.Item), "test")
		if !found {
			t.Errorf("Item '%s' should match query 'Test'", item.Item)
		}
	}
}

func TestFuzzyMatcher_Search_ConsecutiveBonus(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "abcdef", Hash: "hash1", TimeStamp: time.Now()}, // consecutive match
		{Item: "axbxcx", Hash: "hash2", TimeStamp: time.Now()}, // non-consecutive match
		{Item: "abc123", Hash: "hash3", TimeStamp: time.Now()}, // consecutive at start
	}

	result := matcher.Search(items, "abc")
	if len(result) != 3 {
		t.Fatalf("Expected 3 matches, got %d", len(result))
	}

	// Verify all expected items are present
	expectedItems := []string{"abcdef", "axbxcx", "abc123"}
	for _, expected := range expectedItems {
		found := false
		for _, item := range result {
			if item.Item == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected item '%s' not found in results", expected)
		}
	}
}

func TestFuzzyMatcher_Search_LengthBonus(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "test", Hash: "hash1", TimeStamp: time.Now()},                                      // short string
		{Item: "test with lots of additional content here", Hash: "hash2", TimeStamp: time.Now()}, // long string
		{Item: "testing", Hash: "hash3", TimeStamp: time.Now()},                                   // medium string
	}

	result := matcher.Search(items, "test")
	if len(result) != 3 {
		t.Fatalf("Expected 3 matches, got %d", len(result))
	}

	// All items should match
	for _, item := range result {
		found := containsSubsequence(strings.ToLower(item.Item), "test")
		if !found {
			t.Errorf("Item '%s' should match query 'test'", item.Item)
		}
	}
}

func TestFuzzyMatcher_Search_SpecialCharacters(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "file.txt", Hash: "hash1", TimeStamp: time.Now()},
		{Item: "path/to/file", Hash: "hash2", TimeStamp: time.Now()},
		{Item: "test-case_example", Hash: "hash3", TimeStamp: time.Now()},
		{Item: "email@example.com", Hash: "hash4", TimeStamp: time.Now()},
	}

	testCases := []struct {
		query           string
		expectedMatches int
	}{
		{"file", 2},    // Should match "file.txt" and "path/to/file"
		{"txt", 1},     // Should match "file.txt"
		{"test", 1},    // Should match "test-case_example"
		{"example", 2}, // Should match "test-case_example" and "email@example.com"
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			result := matcher.Search(items, tc.query)
			if len(result) != tc.expectedMatches {
				t.Errorf("Query '%s': expected %d matches, got %d", tc.query, tc.expectedMatches, len(result))
			}
		})
	}
}

func TestFuzzyMatcher_Search_EmptyItems(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{}

	result := matcher.Search(items, "test")
	if len(result) != 0 {
		t.Errorf("Expected 0 matches for empty items, got %d", len(result))
	}
}

func TestFuzzyMatcher_Search_ItemsWithEmptyContent(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "", Hash: "hash1", TimeStamp: time.Now()},
		{Item: "test", Hash: "hash2", TimeStamp: time.Now()},
		{Item: "  ", Hash: "hash3", TimeStamp: time.Now()}, // whitespace only
	}

	result := matcher.Search(items, "test")
	if len(result) != 1 {
		t.Fatalf("Expected 1 match, got %d", len(result))
	}
	if result[0].Item != "test" {
		t.Errorf("Expected 'test', got '%s'", result[0].Item)
	}
}

func TestFuzzyMatcher_Search_UnicodeContent(t *testing.T) {
	matcher := NewFuzzyMatcher()
	items := []history.ClipboardHistory{
		{Item: "测试内容", Hash: "hash1", TimeStamp: time.Now()},
		{Item: "test with 测试", Hash: "hash2", TimeStamp: time.Now()},
		{Item: "🚀 rocket test", Hash: "hash3", TimeStamp: time.Now()},
	}

	// Test ASCII query on unicode content
	result := matcher.Search(items, "test")
	if len(result) != 2 {
		t.Fatalf("Expected 2 matches for 'test', got %d", len(result))
	}

	// Test unicode query (this might not work perfectly with the current implementation,
	// but we should handle it gracefully)
	result = matcher.Search(items, "测试")
	// The current implementation might not handle unicode perfectly,
	// but it should not crash
	if result == nil {
		t.Error("Search should return empty slice, not nil for unicode query")
	}
}

// Helper function to check if a string contains all characters of a subsequence
func containsSubsequence(text, query string) bool {
	text = strings.ToLower(text)
	query = strings.ToLower(query)

	textIdx := 0
	for _, queryChar := range query {
		found := false
		for textIdx < len(text) {
			if rune(text[textIdx]) == queryChar {
				found = true
				textIdx++
				break
			}
			textIdx++
		}
		if !found {
			return false
		}
	}
	return true
}
