package main

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/ui"
)

func main() {
	// Create history manager
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
