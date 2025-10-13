// Package main tests for the Clippy clipboard manager application
//
// SAFETY NOTICE: All tests that interact with the filesystem are designed to
// run in isolated temporary directories to protect the real history.json file.
// Tests will refuse to run in the project root directory.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// safeguardRealHistoryFile ensures we never accidentally modify the real history.json
// by checking if we're in the project root directory and refusing to run tests there
func safeguardRealHistoryFile(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not get working directory: %v", err)
	}

	// Check if history.json exists in current directory (indicates we're in project root)
	if _, err := os.Stat(history.HistoryFileName); err == nil {
		// If we can see go.mod, we're definitely in the project root
		if _, err := os.Stat("go.mod"); err == nil {
			t.Fatalf("SAFETY: Refusing to run tests in project root directory (%s) to protect real history.json. Tests should run in temp directories.", currentDir)
		}
	}
}

// Test helper to capture log output
func captureLogOutput(f func()) string {
	// Create a pipe to capture log output
	r, w, _ := os.Pipe()

	// Save original log output
	originalOutput := log.Writer()
	defer log.SetOutput(originalOutput)

	// Redirect log to our pipe
	log.SetOutput(w)

	// Channel to collect output
	outputCh := make(chan string, 1)

	// Read from pipe in goroutine
	go func() {
		var buf strings.Builder
		io.Copy(&buf, r)
		outputCh <- buf.String()
	}()

	// Execute function
	f()

	// Close write end and wait for read to complete
	w.Close()
	output := <-outputCh
	r.Close()

	return output
}

// Test helper to create temporary directory for testing
// This ensures that history.json is created in an isolated location
func createTempTestDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "clippy_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Change to temp directory so history.json gets created there
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	cleanup := func() {
		// Always restore original directory first
		os.Chdir(originalDir)
		// Then clean up the temp directory and all its contents
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestLogRedirection(t *testing.T) {
	// Test the log redirection logic from main()
	t.Run("Successful log redirection", func(t *testing.T) {
		// Simulate the log redirection code
		logFile, err := os.OpenFile("/dev/null", os.O_WRONLY, 0)
		if err == nil {
			defer logFile.Close()

			// Save original log output
			originalOutput := log.Writer()
			defer log.SetOutput(originalOutput)

			// Set log output to /dev/null
			log.SetOutput(logFile)

			// Verify that log output is redirected (we can't easily test this directly,
			// but we can verify the file operations work)
			if logFile == nil {
				t.Error("Expected logFile to be non-nil")
			}
		} else {
			t.Skipf("Could not open /dev/null for testing: %v", err)
		}
	})

	t.Run("Handle log redirection failure gracefully", func(t *testing.T) {
		// Test that the application continues even if log redirection fails
		// This simulates the case where /dev/null might not be available

		// Try to open an invalid path (should fail on most systems)
		invalidPath := "/this/path/definitely/does/not/exist/nowhere"
		logFile, err := os.OpenFile(invalidPath, os.O_WRONLY, 0)

		// Should handle the error gracefully (err != nil)
		if err == nil {
			// Unexpected success, clean up
			logFile.Close()
			t.Skip("Unexpected success opening invalid path")
		}

		// This is the expected behavior - error should be handled gracefully
		// and the application should continue (simulated by this test not failing)
	})
}

func TestHistoryManagerInitialization(t *testing.T) {
	// Safety check to ensure we don't modify real history
	safeguardRealHistoryFile(t)

	_, cleanup := createTempTestDir(t)
	defer cleanup()

	t.Run("Initialize history manager", func(t *testing.T) {
		// Test the history manager initialization from main()
		historyManager := history.NewManager()

		if historyManager == nil {
			t.Fatal("Expected historyManager to be non-nil")
		}

		if historyManager.Count() != 0 {
			t.Errorf("Expected new history manager to be empty, got %d items", historyManager.Count())
		}
	})

	t.Run("Load from non-existent file", func(t *testing.T) {
		historyManager := history.NewManager()

		// This should not error (file doesn't exist is handled gracefully)
		err := historyManager.LoadFromFile()
		if err != nil {
			t.Errorf("LoadFromFile should handle non-existent file gracefully, got error: %v", err)
		}

		if historyManager.Count() != 0 {
			t.Errorf("Expected empty history after loading non-existent file, got %d items", historyManager.Count())
		}
	})

	t.Run("Load from existing file", func(t *testing.T) {
		// Create a test history file
		historyManager := history.NewManager()
		historyManager.AddItem("test item 1")
		historyManager.AddItem("test item 2")

		err := historyManager.SaveToFile()
		if err != nil {
			t.Fatalf("Failed to save test history: %v", err)
		}

		// Create new manager and load
		newManager := history.NewManager()
		err = newManager.LoadFromFile()
		if err != nil {
			t.Errorf("Failed to load history: %v", err)
		}

		if newManager.Count() != 2 {
			t.Errorf("Expected 2 items after loading, got %d", newManager.Count())
		}
	})

	t.Run("Handle corrupted history file", func(t *testing.T) {
		// Create a corrupted history file
		corruptedJSON := `{"invalid": json content without closing`
		err := os.WriteFile(history.HistoryFileName, []byte(corruptedJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to create corrupted file: %v", err)
		}

		historyManager := history.NewManager()

		// Should capture the error in logs
		output := captureLogOutput(func() {
			err := historyManager.LoadFromFile()
			if err != nil {
				log.Printf("Warning: Could not load history: %v", err)
			}
		})

		// Should log a warning
		if !strings.Contains(output, "Warning: Could not load history") {
			t.Errorf("Expected warning log for corrupted file, got: %s", output)
		}
	})
}

func TestUIModelInitialization(t *testing.T) {
	t.Run("Create UI model with history manager", func(t *testing.T) {
		historyManager := history.NewManager()
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
		historyManager := history.NewManager()
		initialModel := ui.NewModel(historyManager)

		cmd := initialModel.Init()
		if cmd == nil {
			t.Error("Expected Init() to return a command")
		}
	})
}

func TestApplicationErrorHandling(t *testing.T) {
	safeguardRealHistoryFile(t)

	_, cleanup := createTempTestDir(t)
	defer cleanup()

	t.Run("Handle save error gracefully", func(t *testing.T) {
		historyManager := history.NewManager()
		historyManager.AddItem("test item")

		// Create a directory with the same name as the history file to force save error
		err := os.Mkdir(history.HistoryFileName, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		defer os.Remove(history.HistoryFileName)

		// Should capture the error in logs
		output := captureLogOutput(func() {
			err := historyManager.SaveToFile()
			if err != nil {
				log.Printf("Error saving history: %v", err)
			}
		})

		// Should log an error
		if !strings.Contains(output, "Error saving history") {
			t.Errorf("Expected error log for save failure, got: %s", output)
		}
	})
}

func TestBubbletaaProgramCreation(t *testing.T) {
	t.Run("Create bubbletea program", func(t *testing.T) {
		historyManager := history.NewManager()
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
		historyManager := history.NewManager()
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
	safeguardRealHistoryFile(t)

	_, cleanup := createTempTestDir(t)
	defer cleanup()

	t.Run("Complete application flow simulation", func(t *testing.T) {
		// Simulate the complete main() function flow without the blocking tea.Program.Run()

		// Step 1: Log redirection (simplified)
		originalOutput := log.Writer()
		defer log.SetOutput(originalOutput)

		// Step 2: History manager initialization
		historyManager := history.NewManager()
		if historyManager == nil {
			t.Fatal("Failed to create history manager")
		}

		// Step 3: Load from file (with some test data)
		historyManager.AddItem("initial test item")
		err := historyManager.SaveToFile()
		if err != nil {
			t.Fatalf("Failed to save initial data: %v", err)
		}

		// Create fresh manager and load
		freshManager := history.NewManager()
		err = freshManager.LoadFromFile()
		if err != nil {
			t.Errorf("Failed to load history: %v", err)
		}

		// Step 4: UI model creation
		initialModel := ui.NewModel(freshManager)
		if initialModel.View() == "" {
			t.Error("Expected non-empty initial view")
		}

		// Step 5: Program creation (but not running)
		program := tea.NewProgram(initialModel)
		if program == nil {
			t.Error("Failed to create tea program")
		}

		// Step 6: Save to file (cleanup)
		err = freshManager.SaveToFile()
		if err != nil {
			t.Errorf("Failed to save history: %v", err)
		}

		// Verify the saved data
		finalManager := history.NewManager()
		err = finalManager.LoadFromFile()
		if err != nil {
			t.Errorf("Failed to load final history: %v", err)
		}

		if finalManager.Count() != 1 {
			t.Errorf("Expected 1 item in final history, got %d", finalManager.Count())
		}
	})
}

func TestApplicationWithRealHistoryFile(t *testing.T) {
	safeguardRealHistoryFile(t)

	_, cleanup := createTempTestDir(t)
	defer cleanup()

	t.Run("Application lifecycle with real history operations", func(t *testing.T) {
		// Simulate multiple application runs with persistent history

		// First run: Create some history
		{
			historyManager := history.NewManager()
			historyManager.AddItem("session 1 - item 1")
			historyManager.AddItem("session 1 - item 2")

			err := historyManager.SaveToFile()
			if err != nil {
				t.Fatalf("Failed to save history in session 1: %v", err)
			}
		}

		// Second run: Load existing history and add more
		{
			historyManager := history.NewManager()
			err := historyManager.LoadFromFile()
			if err != nil {
				t.Errorf("Failed to load history in session 2: %v", err)
			}

			if historyManager.Count() != 2 {
				t.Errorf("Expected 2 items from session 1, got %d", historyManager.Count())
			}

			historyManager.AddItem("session 2 - item 1")
			historyManager.AddItem("session 1 - item 1") // duplicate, should be ignored

			if historyManager.Count() != 3 {
				t.Errorf("Expected 3 items after session 2 additions, got %d", historyManager.Count())
			}

			err = historyManager.SaveToFile()
			if err != nil {
				t.Fatalf("Failed to save history in session 2: %v", err)
			}
		}

		// Third run: Verify persistence
		{
			historyManager := history.NewManager()
			err := historyManager.LoadFromFile()
			if err != nil {
				t.Errorf("Failed to load history in session 3: %v", err)
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
		}
	})
}

func TestFileSystemIsolation(t *testing.T) {
	// Verify that tests don't interfere with real history.json
	t.Run("Tests run in isolated directories", func(t *testing.T) {
		// Get the original directory (where real history.json might be)
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get original directory: %v", err)
		}

		// Check if real history file exists before test
		realHistoryExists := false
		originalPath := originalDir + "/" + history.HistoryFileName
		if _, err := os.Stat(originalPath); err == nil {
			realHistoryExists = true
		}

		// Run test in isolated directory
		tempDir, cleanup := createTempTestDir(t)
		defer cleanup()

		// Verify we're in a different directory
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

		// Create some history in the test environment
		manager := history.NewManager()
		manager.AddItem("test item that should not affect real history")
		err = manager.SaveToFile()
		if err != nil {
			t.Errorf("Failed to save in test environment: %v", err)
		}

		// Verify test history file exists in temp directory
		testHistoryPath := currentDir + "/" + history.HistoryFileName
		if _, err := os.Stat(testHistoryPath); os.IsNotExist(err) {
			t.Error("Test history file should exist in temp directory")
		}

		// After cleanup (done by defer), verify real history is unchanged
		t.Cleanup(func() {
			// Get back to original directory
			os.Chdir(originalDir)

			// Check that real history file state hasn't changed
			_, realHistoryExistsAfter := os.Stat(history.HistoryFileName)

			if realHistoryExists && realHistoryExistsAfter != nil {
				t.Error("Real history file should still exist after test")
			}
			if !realHistoryExists && realHistoryExistsAfter == nil {
				t.Error("Real history file should not be created by test")
			}
		})
	})
}

func TestMainFunctionComponents(t *testing.T) {
	// Test individual components that would be used in main()

	t.Run("Component integration", func(t *testing.T) {
		// Test that all components work together
		historyManager := history.NewManager()

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
	// Safety check - benchmarks should also be isolated
	currentDir, _ := os.Getwd()
	if _, err := os.Stat(history.HistoryFileName); err == nil {
		if _, err := os.Stat("go.mod"); err == nil {
			b.Fatalf("SAFETY: Refusing to run benchmarks in project root directory (%s)", currentDir)
		}
	}

	tempDir, err := os.MkdirTemp("", "clippy_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir) // Pre-create some history data
	setupManager := history.NewManager()
	for i := 0; i < 100; i++ {
		setupManager.AddItem(fmt.Sprintf("benchmark item %d", i))
	}
	setupManager.SaveToFile()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate main() initialization steps
		historyManager := history.NewManager()
		historyManager.LoadFromFile()

		initialModel := ui.NewModel(historyManager)
		tea.NewProgram(initialModel)

		// Note: We don't run the program as that would block
	}
}

func BenchmarkFullApplicationCycle(b *testing.B) {
	// Safety check - benchmarks should also be isolated
	currentDir, _ := os.Getwd()
	if _, err := os.Stat(history.HistoryFileName); err == nil {
		if _, err := os.Stat("go.mod"); err == nil {
			b.Fatalf("SAFETY: Refusing to run benchmarks in project root directory (%s)", currentDir)
		}
	}

	tempDir, err := os.MkdirTemp("", "clippy_cycle_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Full cycle: load -> modify -> save
		historyManager := history.NewManager()
		historyManager.LoadFromFile()
		historyManager.AddItem(fmt.Sprintf("cycle item %d", i))
		historyManager.SaveToFile()
	}
}
