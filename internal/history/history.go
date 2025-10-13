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
	items    []ClipboardHistory
	hashes   map[string]struct{}
	lastHash string
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
	if !m.containsHash(item.Hash) {
		m.items = append(m.items, item)
		m.lastHash = item.Hash
		m.hashes[item.Hash] = struct{}{}
		return true
	}
	return false
}

func (m *Manager) containsHash(s string) bool {
	_, contains := m.hashes[s]
	return contains || m.lastHash == s
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

// DeleteItem attempts to delete an item by index and returns the removal status
func (m *Manager) DeleteItem(index int) bool {
	if index >= 0 && index < len(m.items) {
		item := m.items[index]
		delete(m.hashes, item.Hash)
		m.items = append(m.items[:index], m.items[index+1:]...)
		return true
	}
	return false
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
		m.lastHash = item.Hash
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
