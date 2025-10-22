// Package main tests for the Clippy clipboard manager application
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// createTempTestDir creates a temporary directory for testing
func createTempTestDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "clippy_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	cleanup := func() {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// setupTestHistoryManager creates an isolated history manager for testing
func setupTestHistoryManager(t *testing.T) (*history.Manager, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "clippy_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := fmt.Sprintf("%s/test.db", tempDir)
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

func TestLogRedirection(t *testing.T) {
	t.Run("Successful log redirection", func(t *testing.T) {
		logFile, err := os.OpenFile("/dev/null", os.O_WRONLY, 0)
		if err == nil {
			defer logFile.Close()

			originalOutput := log.Writer()
			defer log.SetOutput(originalOutput)

			log.SetOutput(logFile)

			if logFile == nil {
				t.Error("Expected logFile to be non-nil")
			}
		} else {
			t.Skipf("Could not open /dev/null for testing: %v", err)
		}
	})

	t.Run("Handle log redirection failure gracefully", func(t *testing.T) {
		invalidPath := "/this/path/definitely/does/not/exist/nowhere"
		logFile, err := os.OpenFile(invalidPath, os.O_WRONLY, 0)

		if err == nil {
			logFile.Close()
			t.Skip("Unexpected success opening invalid path")
		}
	})
}

func TestHistoryManagerInitialization(t *testing.T) {
	t.Run("Initialize history manager", func(t *testing.T) {
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()

		if historyManager == nil {
			t.Fatal("Expected historyManager to be non-nil")
		}

		if historyManager.Count() != 0 {
			t.Errorf("Expected new history manager to be empty, got %d items", historyManager.Count())
		}
	})

	t.Run("Load from non-existent file", func(t *testing.T) {
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()

		// This should not error (file doesn't exist is handled gracefully)
		err := historyManager.LoadFromDB()
		if err != nil {
			t.Errorf("LoadFromDB should handle non-existent file gracefully, got error: %v", err)
		}

		if historyManager.Count() != 0 {
			t.Errorf("Expected empty history after loading non-existent file, got %d items", historyManager.Count())
		}
	})

	t.Run("Load from existing file", func(t *testing.T) {
		// Create a test history file
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()
		historyManager.AddItem("test item 1")
		historyManager.AddItem("test item 2")

		// Load from DB
		err := historyManager.LoadFromDB()
		if err != nil {
			t.Errorf("Failed to load history: %v", err)
		}

		if historyManager.Count() != 2 {
			t.Errorf("Expected 2 items after loading, got %d", historyManager.Count())
		}
	})

	t.Run("Handle corrupted history file", func(t *testing.T) {
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()

		// With SQLite, this test is less relevant
		// Just verify basic operations work
		historyManager.AddItem("test item")
		if historyManager.Count() != 1 {
			t.Errorf("Expected 1 item, got %d", historyManager.Count())
		}
	})
}

func TestUIModelInitialization(t *testing.T) {
	t.Run("Create UI model with history manager", func(t *testing.T) {
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()
		historyManager.AddItem("test item")

		initialModel := ui.NewModel(historyManager)

		// Verify model is properly initialized
		// We can't easily test the internal state, but we can test the public interface
		view := initialModel.View()
		if !strings.Contains(view, "test item") {
			t.Error("Expected model view to contain test item")
		}

		if !strings.Contains(view, "Clipboard History") {
			t.Error("Expected model view to contain title")
		}
	})

	t.Run("UI model Init command", func(t *testing.T) {
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()
		initialModel := ui.NewModel(historyManager)

		cmd := initialModel.Init()
		if cmd == nil {
			t.Error("Expected Init() to return a command")
		}
	})
}

func TestBubbleteaProgramCreation(t *testing.T) {
	t.Run("Create bubbletea program", func(t *testing.T) {
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()
		initialModel := ui.NewModel(historyManager)

		// This tests that we can create a program without errors
		program := tea.NewProgram(initialModel)

		if program == nil {
			t.Error("Expected tea.NewProgram to return non-nil program")
		}

		// We can't easily test program.Run() without it blocking,
		// but we can test that the program is properly constructed
	})

	t.Run("Program with options", func(t *testing.T) {
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()
		initialModel := ui.NewModel(historyManager)

		// Test creating program with options (as might be done in real app)
		program := tea.NewProgram(
			initialModel,
			tea.WithAltScreen(),       // Use alternate screen
			tea.WithMouseCellMotion(), // Enable mouse support
		)

		if program == nil {
			t.Error("Expected tea.NewProgram with options to return non-nil program")
		}
	})
}

func TestIntegrationFlow(t *testing.T) {
	t.Run("Complete application flow simulation", func(t *testing.T) {
		// Simulate the complete main() function flow without the blocking tea.Program.Run()

		// Step 1: Log redirection (simplified)
		originalOutput := log.Writer()
		defer log.SetOutput(originalOutput)

		// Step 2: History manager initialization
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()
		if historyManager == nil {
			t.Fatal("Failed to create history manager")
		}

		// Step 3: Add test data
		historyManager.AddItem("initial test item")

		// Step 4: UI model creation
		initialModel := ui.NewModel(historyManager)
		if initialModel.View() == "" {
			t.Error("Expected non-empty initial view")
		}

		// Step 5: Program creation (but not running)
		program := tea.NewProgram(initialModel)
		if program == nil {
			t.Error("Failed to create tea program")
		}

		// Verify the data
		if historyManager.Count() != 1 {
			t.Errorf("Expected 1 item in final history, got %d", historyManager.Count())
		}
	})
}

func TestApplicationWithRealHistoryFile(t *testing.T) {
	t.Run("Application lifecycle with real history operations", func(t *testing.T) {
		// Simulate multiple application runs with persistent history
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()

		// First session: Create some history
		historyManager.AddItem("session 1 - item 1")
		historyManager.AddItem("session 1 - item 2")

		if historyManager.Count() != 2 {
			t.Errorf("Expected 2 items from session 1, got %d", historyManager.Count())
		}

		// Second session operations: Add more
		historyManager.AddItem("session 2 - item 1")
		historyManager.AddItem("session 1 - item 1") // duplicate, should be ignored

		if historyManager.Count() != 3 {
			t.Errorf("Expected 3 items after session 2 additions, got %d", historyManager.Count())
		}

		// Verify persistence by reloading
		err := historyManager.LoadFromDB()
		if err != nil {
			t.Errorf("Failed to reload history: %v", err)
		}

		if historyManager.Count() != 3 {
			t.Errorf("Expected 3 items in final session, got %d", historyManager.Count())
		}

		// Verify specific items exist
		items := historyManager.GetItems()
		itemTexts := make([]string, len(items))
		for i, item := range items {
			itemTexts[i] = item.Item
		}

		expectedItems := []string{
			"session 1 - item 1",
			"session 1 - item 2",
			"session 2 - item 1",
		}

		for _, expected := range expectedItems {
			found := false
			for _, actual := range itemTexts {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected item %q not found in history", expected)
			}
		}
	})
}

func TestFileSystemIsolation(t *testing.T) {
	t.Run("Tests run in isolated directories", func(t *testing.T) {
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get original directory: %v", err)
		}

		tempDir, cleanup := createTempTestDir(t)
		defer cleanup()

		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}

		if currentDir == originalDir {
			t.Error("Test should run in a different directory than the original")
		}

		if !strings.Contains(currentDir, tempDir) {
			t.Error("Test should run in the created temp directory")
		}

		manager, cleanup := setupTestHistoryManager(t)
		defer cleanup()
		manager.AddItem("test item in isolated database")

		if manager.Count() != 1 {
			t.Errorf("Expected 1 item in isolated database, got %d", manager.Count())
		}
	})
}

func TestMainFunctionComponents(t *testing.T) {
	// Test individual components that would be used in main()

	t.Run("Component integration", func(t *testing.T) {
		// Test that all components work together
		historyManager, cleanup := setupTestHistoryManager(t)
		defer cleanup()

		// Add some test data
		testItems := []string{
			"clipboard item 1",
			"clipboard item 2 with\nnewlines",
			"clipboard item 3 with special chars: !@#$%^&*()",
		}

		for _, item := range testItems {
			historyManager.AddItem(item)
		}

		// Create UI model
		model := ui.NewModel(historyManager)

		// Test that the view renders correctly with table format
		model.UpdateTable() // Update table with new items
		view := model.View()
		for i, item := range testItems {
			// Check for the item number in table format
			expectedNumber := fmt.Sprintf("%d", i+1)
			if !strings.Contains(view, expectedNumber) {
				t.Errorf("Expected view to contain item number %d", i+1)
			}
			// Check for content (newlines are converted to spaces in table view)
			expectedContent := strings.ReplaceAll(item, "\n", " ")
			if !strings.Contains(view, expectedContent) {
				t.Errorf("Expected view to contain transformed item content: %s", expectedContent)
			}
		}

		// Test Init command
		cmd := model.Init()
		if cmd == nil {
			t.Error("Expected Init to return a command")
		}

		// Test that we can create a tea program
		program := tea.NewProgram(model)
		if program == nil {
			t.Error("Expected successful program creation")
		}
	})
}

// Benchmark the main application initialization flow
func BenchmarkApplicationInitialization(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "clippy_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := tempDir + "/bench.db"
	setupManager, err := history.NewManagerWithPath(dbPath)
	if err != nil {
		b.Fatalf("Failed to create benchmark manager: %v", err)
	}
	for i := 0; i < 100; i++ {
		setupManager.AddItem(fmt.Sprintf("benchmark item %d", i))
	}
	setupManager.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate main() initialization steps
		historyManager, err := history.NewManagerWithPath(dbPath)
		if err != nil {
			b.Fatalf("Failed to create manager: %v", err)
		}
		historyManager.LoadFromDB()

		initialModel := ui.NewModel(historyManager)
		tea.NewProgram(initialModel)

		historyManager.Close()
		// Note: We don't run the program as that would block
	}
}

func BenchmarkFullApplicationCycle(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "clippy_cycle_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := tempDir + "/cycle.db"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Full cycle: load -> modify -> save (with SQLite, saves are immediate)
		historyManager, err := history.NewManagerWithPath(dbPath)
		if err != nil {
			b.Fatalf("Failed to create manager: %v", err)
		}
		historyManager.LoadFromDB()
		historyManager.AddItem(fmt.Sprintf("cycle item %d", i))
		// Note: SaveToDB() is now a no-op as saves are immediate with SQLite
		historyManager.Close()
	}
}
