package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bvdwalt/clippy/internal/history"
)

// setupTestHistoryManager creates an isolated history manager for testing
func setupTestHistoryManager(t *testing.T) (*history.Manager, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "clippy_ui_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	manager, err := history.NewManagerWithPath(dbPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create test manager: %v", err)
	}

	cleanup := func() {
		manager.Close()
		os.RemoveAll(tempDir)
	}

	return manager, cleanup
}

// setupTestHistoryManagerForBench creates an isolated history manager for benchmarking
func setupTestHistoryManagerForBench(b *testing.B) (*history.Manager, func()) {
	b.Helper()

	tempDir, err := os.MkdirTemp("", "clippy_ui_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "bench.db")
	manager, err := history.NewManagerWithPath(dbPath)
	if err != nil {
		os.RemoveAll(tempDir)
		b.Fatalf("Failed to create benchmark manager: %v", err)
	}

	cleanup := func() {
		manager.Close()
		os.RemoveAll(tempDir)
	}

	return manager, cleanup
}
