package main

import (
	"log"

	"github.com/bvdwalt/clippy/internal/history"
	"github.com/bvdwalt/clippy/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create a temporary history manager
	historyManager, err := history.NewManager()
	if err != nil {
		log.Fatal("Failed to create history manager:", err)
	}
	defer historyManager.Close()

	// Add some sample data to demonstrate the enhanced UI
	sampleData := []string{
		"Hello, World!",
		"This is a longer clipboard entry that will be truncated in the table view to show how long content is handled",
		"https://github.com/charmbracelet/bubbletea",
		"func main() {\n\tfmt.Println(\"Hello\")\n}",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit",
		"cat /etc/passwd | grep root",
		"SELECT * FROM users WHERE active = 1;",
		"🎉 Emojis work too! 🚀✨",
		"Multi\nline\ntext\nentry",
		"Short",
	}

	for _, data := range sampleData {
		historyManager.AddItem(data)
	}

	// Create the enhanced model
	model := ui.NewModel(historyManager)

	// Create the bubbletea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program:", err)
	}
}
