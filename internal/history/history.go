package history

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const HistoryFileName = "history.json"

// Manager handles clipboard history storage and management
type Manager struct {
	items  []ClipboardHistory
	hashes map[string]struct{}
}

// NewManager creates a new history manager
func NewManager() *Manager {
	return &Manager{
		items:  make([]ClipboardHistory, 0),
		hashes: make(map[string]struct{}),
	}
}

// AddItem adds a new clipboard item if it doesn't already exist
func (m *Manager) AddItem(content string) bool {
	item := newClipboardItem(content)
	if _, exists := m.hashes[item.Hash]; !exists {
		m.items = append(m.items, item)
		m.hashes[item.Hash] = struct{}{}
		return true
	}
	return false
}

// GetItems returns all clipboard history items
func (m *Manager) GetItems() []ClipboardHistory {
	return m.items
}

// GetItem returns a specific item by index
func (m *Manager) GetItem(index int) (ClipboardHistory, bool) {
	if index >= 0 && index < len(m.items) {
		return m.items[index], true
	}
	return ClipboardHistory{}, false
}

// Count returns the number of items in history
func (m *Manager) Count() int {
	return len(m.items)
}

// LoadFromFile loads history from the JSON file
func (m *Manager) LoadFromFile() error {
	file, err := os.Open(HistoryFileName)
	if err != nil {
		// File doesn't exist, start with empty history
		return nil
	}
	defer file.Close()

	var items []ClipboardHistory
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&items)
	if err != nil {
		return fmt.Errorf("error loading history: %w", err)
	}

	m.items = items
	m.hashes = make(map[string]struct{})
	for _, item := range items {
		m.hashes[item.Hash] = struct{}{}
	}

	return nil
}

// SaveToFile saves history to the JSON file
func (m *Manager) SaveToFile() error {
	file, err := os.Create(HistoryFileName)
	if err != nil {
		return fmt.Errorf("error creating history file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(m.items)
	if err != nil {
		return fmt.Errorf("error saving history: %w", err)
	}

	return nil
}

// newClipboardItem creates a new clipboard history item
func newClipboardItem(content string) ClipboardHistory {
	return ClipboardHistory{
		Item:      content,
		Hash:      fmt.Sprintf("%x", sha256.Sum256([]byte(content))),
		TimeStamp: time.Now(),
	}
}
