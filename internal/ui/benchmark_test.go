package ui

import (
	"strings"
	"testing"
)

func BenchmarkNewModel(b *testing.B) {
	historyManager, cleanup := setupTestHistoryManagerForBench(b)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewModel(historyManager)
	}
}

func BenchmarkModelView(b *testing.B) {
	historyManager, cleanup := setupTestHistoryManagerForBench(b)
	defer cleanup()
	model := NewModel(historyManager)

	// Add some items for benchmarking
	for i := 0; i < 100; i++ {
		historyManager.AddItem(generateTestContent(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.View()
	}
}

func BenchmarkModelViewLargeHistory(b *testing.B) {
	historyManager, cleanup := setupTestHistoryManagerForBench(b)
	defer cleanup()
	model := NewModel(historyManager)

	// Add many items to test performance with large history
	for i := 0; i < 1000; i++ {
		historyManager.AddItem(generateTestContent(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.View()
	}
}

func BenchmarkModelViewLongContent(b *testing.B) {
	historyManager, cleanup := setupTestHistoryManagerForBench(b)
	defer cleanup()
	model := NewModel(historyManager)

	// Add items with very long content
	longContent := strings.Repeat("This is a very long piece of content that exceeds the display limit. ", 20)
	for i := 0; i < 50; i++ {
		historyManager.AddItem(longContent + generateTestContent(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.View()
	}
}

func BenchmarkTick(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Tick()
	}
}

// Helper function to generate test content
func generateTestContent(i int) string {
	contents := []string{
		"Simple text content",
		"Content with\nnewlines\nand\ttabs",
		"Unicode content: 你好世界 🚀 🎉",
		"JSON-like: {\"key\": \"value\", \"number\": 123}",
		"Code snippet: func main() { fmt.Println(\"Hello\") }",
		"URL: https://example.com/path?param=value",
		"Email: user@example.com",
		"Long text that will definitely exceed the sixty character limit and should be truncated",
	}
	return contents[i%len(contents)]
}
