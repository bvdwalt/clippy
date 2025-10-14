package history

import (
	"fmt"
	"testing"
)

func BenchmarkAddItem(b *testing.B) {
	manager := NewManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		content := fmt.Sprintf("item_%d", i)
		manager.AddItem(content)
	}
}

func BenchmarkAddItemDuplicates(b *testing.B) {
	manager := NewManager()
	content := "duplicate content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.AddItem(content)
	}
}

func BenchmarkGetItem(b *testing.B) {
	manager := NewManager()

	// Pre-populate with 1000 items
	for i := 0; i < 1000; i++ {
		manager.AddItem(fmt.Sprintf("item_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetItem(i % 1000)
	}
}

func BenchmarkNewClipboardItem(b *testing.B) {
	content := "benchmark test content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newClipboardItem(content)
	}
}

func BenchmarkNewClipboardItemLarge(b *testing.B) {
	// Test with larger content (1KB)
	content := string(make([]byte, 1024))
	for i := range content {
		content = content[:i] + "A" + content[i+1:]
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newClipboardItem(content)
	}
}

func BenchmarkSaveToFile(b *testing.B) {
	manager := NewManager()

	// Pre-populate with some items
	for i := 0; i < 100; i++ {
		manager.AddItem(fmt.Sprintf("item_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.SaveToFile()
	}

	// Clean up
	b.StopTimer()
	// Note: This will leave the last file, but that's okay for benchmarking
}

func BenchmarkLoadFromFile(b *testing.B) {
	// Create a test file first
	setupManager := NewManager()
	for i := 0; i < 100; i++ {
		setupManager.AddItem(fmt.Sprintf("item_%d", i))
	}
	setupManager.SaveToFile()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewManager()
		manager.LoadFromFile()
	}
}
