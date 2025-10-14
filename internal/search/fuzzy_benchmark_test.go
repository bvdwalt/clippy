package search

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/bvdwalt/clippy/internal/history"
)

func BenchmarkFuzzyMatcher_Search(b *testing.B) {
	matcher := NewFuzzyMatcher()

	// Create test data with various sizes
	benchmarks := []struct {
		name  string
		items int
	}{
		{"Small_10", 10},
		{"Medium_100", 100},
		{"Large_1000", 1000},
		{"XLarge_10000", 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			items := generateTestItems(bm.items)
			query := "test"

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = matcher.Search(items, query)
			}
		})
	}
}

func BenchmarkFuzzyMatcher_Search_QueryLength(b *testing.B) {
	matcher := NewFuzzyMatcher()
	items := generateTestItems(1000)

	queries := []struct {
		name  string
		query string
	}{
		{"Short_1char", "t"},
		{"Short_3char", "tes"},
		{"Medium_6char", "testing"},
		{"Long_12char", "long_test_query"},
		{"VeryLong_24char", "very_long_test_query_here"},
	}

	for _, q := range queries {
		b.Run(q.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = matcher.Search(items, q.query)
			}
		})
	}
}

func BenchmarkFuzzyMatcher_Search_MatchRatio(b *testing.B) {
	matcher := NewFuzzyMatcher()

	// Create datasets with different match ratios
	benchmarks := []struct {
		name       string
		matchRatio float64 // percentage of items that will match
	}{
		{"HighMatch_90pct", 0.9},
		{"MediumMatch_50pct", 0.5},
		{"LowMatch_10pct", 0.1},
		{"VeryLowMatch_1pct", 0.01},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			items := generateTestItemsWithMatchRatio(1000, "test", bm.matchRatio)
			query := "test"

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = matcher.Search(items, query)
			}
		})
	}
}

func BenchmarkFuzzyMatcher_FuzzyMatch(b *testing.B) {
	matcher := NewFuzzyMatcher()

	testCases := []struct {
		name  string
		text  string
		query string
	}{
		{"Short_Exact", "test", "test"},
		{"Short_Partial", "testing", "test"},
		{"Medium_Match", "this is a test string", "test"},
		{"Long_Match", "this is a very long string with test somewhere in the middle of it", "test"},
		{"Long_NoMatch", "this is a very long string without the target word anywhere in it", "xyz"},
		{"CamelCase", "myTestFunctionName", "test"},
		{"WithSpecialChars", "test-case_example.txt", "test"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = matcher.fuzzyMatch(tc.text, tc.query)
			}
		})
	}
}

func BenchmarkFuzzyMatcher_SortByScore(b *testing.B) {
	matcher := NewFuzzyMatcher()

	// Create scored items with different sizes
	benchmarks := []struct {
		name  string
		items int
	}{
		{"Small_10", 10},
		{"Medium_100", 100},
		{"Large_1000", 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			scoredItems := generateScoredItems(bm.items)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Create a copy for each iteration since sorting modifies the slice
				items := make([]ScoredItem, len(scoredItems))
				copy(items, scoredItems)
				matcher.sortByScore(items)
			}
		})
	}
}

// Helper functions for benchmarks

func generateTestItems(count int) []history.ClipboardHistory {
	items := make([]history.ClipboardHistory, count)
	rand.Seed(time.Now().UnixNano())

	// Common words that might appear in clipboard history
	words := []string{
		"test", "example", "code", "function", "variable", "string", "number",
		"file", "path", "directory", "document", "text", "content", "data",
		"user", "admin", "system", "config", "settings", "option", "value",
		"create", "update", "delete", "select", "insert", "query", "database",
		"server", "client", "request", "response", "api", "endpoint", "url",
		"error", "debug", "log", "info", "warning", "success", "failure",
	}

	for i := 0; i < count; i++ {
		// Generate random content
		numWords := rand.Intn(5) + 1 // 1-5 words
		content := ""
		for j := 0; j < numWords; j++ {
			if j > 0 {
				content += " "
			}
			content += words[rand.Intn(len(words))]
		}

		items[i] = history.ClipboardHistory{
			Item:      content,
			Hash:      fmt.Sprintf("hash%d", i),
			TimeStamp: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}

	return items
}

func generateTestItemsWithMatchRatio(count int, matchWord string, matchRatio float64) []history.ClipboardHistory {
	items := make([]history.ClipboardHistory, count)
	rand.Seed(time.Now().UnixNano())

	matchCount := int(float64(count) * matchRatio)

	words := []string{
		"example", "code", "function", "variable", "string", "number",
		"file", "path", "directory", "document", "text", "content", "data",
		"user", "admin", "system", "config", "settings", "option", "value",
	}

	for i := 0; i < count; i++ {
		var content string
		if i < matchCount {
			// Items that will match
			if rand.Float64() < 0.5 {
				// Put match word at the beginning
				content = matchWord + " " + words[rand.Intn(len(words))]
			} else {
				// Put match word in the middle or end
				content = words[rand.Intn(len(words))] + " " + matchWord
			}
		} else {
			// Items that won't match
			numWords := rand.Intn(3) + 1
			for j := 0; j < numWords; j++ {
				if j > 0 {
					content += " "
				}
				content += words[rand.Intn(len(words))]
			}
		}

		items[i] = history.ClipboardHistory{
			Item:      content,
			Hash:      fmt.Sprintf("hash%d", i),
			TimeStamp: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}

	return items
}

func generateScoredItems(count int) []ScoredItem {
	items := make([]ScoredItem, count)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		items[i] = ScoredItem{
			Item: history.ClipboardHistory{
				Item:      fmt.Sprintf("item %d", i),
				Hash:      fmt.Sprintf("hash%d", i),
				TimeStamp: time.Now(),
			},
			Score: rand.Intn(1000), // Random score 0-999
		}
	}

	return items
}
