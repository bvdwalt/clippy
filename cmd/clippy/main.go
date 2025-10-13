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
		defer logFile.Close()
	}

	historyManager := history.NewManager()
	if err := historyManager.LoadFromFile(); err != nil {
		log.Printf("Warning: Could not load history: %v", err)
	}

	initialModel := ui.NewModel(historyManager)
	program := tea.NewProgram(initialModel)

	_, err = program.Run()
	if err != nil {
		log.Fatal(err)
	}

	if err := historyManager.SaveToFile(); err != nil {
		log.Printf("Error saving history: %v", err)
	}
}
