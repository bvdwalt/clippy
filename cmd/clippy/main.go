package main

import (
	"log"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/ui"
)

var version = "dev"

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

	ttl, err := history.ParseTTL(os.Getenv("CLIPPY_TTL"))
	if err != nil {
		log.Printf("Warning: %v, using default %s", err, history.DefaultTTL)
		ttl = history.DefaultTTL
	}
	historyManager.SetTTL(ttl)
	historyManager.PruneExpired(time.Now())

	initialModel := ui.NewModel(historyManager, version)
	program := tea.NewProgram(initialModel)

	_, err = program.Run()
	if err != nil {
		log.Fatal(err)
	}
}
