package main

import (
	"log"
	"os"
	"path/filepath"
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

	// stderr output would be drawn over the TUI
	if logFile, err := os.OpenFile(filepath.Join(filepath.Dir(historyManager.DBPath()), "clippy.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); err == nil {
		defer func() { _ = logFile.Close() }()
		log.SetOutput(logFile)
	}

	initialModel := ui.NewModel(historyManager, version)
	program := tea.NewProgram(initialModel)

	_, err = program.Run()
	if err != nil {
		log.Fatal(err)
	}
}
