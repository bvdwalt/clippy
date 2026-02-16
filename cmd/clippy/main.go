package main

import (
	"log"
	"os"

	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Redirect stderr to suppress wl-clipboard messages
	logFile, err := os.OpenFile("/dev/null", os.O_WRONLY, 0)
	if err == nil {
		log.SetOutput(logFile)
		defer func() {
			if err := logFile.Close(); err != nil {
				log.Printf("Failed to close log file: %v", err)
			}
		}()
	}

	historyManager, err := history.NewManager()
	if err != nil {
		log.Fatalf("Failed to create history manager: %v", err)
	}
	defer func() {
		if err := historyManager.Close(); err != nil {
			log.Printf("Failed to close history manager: %v", err)
		}
	}()

	if err := historyManager.LoadFromDB(); err != nil {
		log.Printf("Warning: Could not load history: %v", err)
	}

	initialModel := ui.NewModel(historyManager)
	program := tea.NewProgram(initialModel)

	_, err = program.Run()
	if err != nil {
		log.Fatal(err)
	}
}
