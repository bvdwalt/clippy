package history

import (
	"fmt"
	"testing"
)

func BenchmarkAddItem(b *testing.B) {
	manager, cleanup := setupTestManager(&testing.T{})
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		content := fmt.Sprintf("item_%d", i)
		manager.AddItem(content)
	}
}

func BenchmarkAddItemDuplicates(b *testing.B) {
	manager, cleanup := setupTestManager(&testing.T{})
	defer cleanup()
	content := "duplicate content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.AddItem(content)
	}
}

func BenchmarkGetItem(b *testing.B) {
	manager, cleanup := setupTestManager(&testing.T{})
	defer cleanup()

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

func BenchmarkLoadFromDB(b *testing.B) {
	// Create a test manager and populate it
	setupManager, setupCleanup := setupTestManager(&testing.T{})
	for i := 0; i < 100; i++ {
		setupManager.AddItem(fmt.Sprintf("item_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testManager, testCleanup := setupTestManager(&testing.T{})
		testManager.LoadFromDB()
		testCleanup()
	}

	b.StopTimer()
	setupCleanup()
}
