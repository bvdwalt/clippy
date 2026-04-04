package history

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bvdwalt/clippy/internal/db"
)

const (
	ConfigDir  = ".clippy"
	DBFileName = "clippy.automerge"
)

// Manager handles clipboard history storage and management
type Manager struct {
	items    []ClipboardHistory
	hashes   map[string]struct{}
	lastHash string
	dbClient db.DBClient
	dbPath   string
}

// NewManager creates a new history manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating config directory: %w", err)
	}

	dbPath := filepath.Join(configDir, DBFileName)

	return NewManagerWithPath(dbPath)
}

// NewInMemoryManager creates a history manager with no database backing.
// Items are stored in memory only and are not persisted between runs.
func NewInMemoryManager() *Manager {
	return &Manager{
		items:  make([]ClipboardHistory, 0),
		hashes: make(map[string]struct{}),
	}
}

// NewManagerWithPath creates a new history manager with a custom database path
// This is useful for testing with isolated databases
func NewManagerWithPath(dbPath string) (*Manager, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error creating directory: %w", err)
	}

	dbClient, err := db.NewAutomergeClient(dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	manager := &Manager{
		items:    make([]ClipboardHistory, 0),
		hashes:   make(map[string]struct{}),
		dbClient: dbClient,
		dbPath:   dbPath,
	}

	return manager, nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	if m.dbClient == nil {
		return nil
	}
	return m.dbClient.Close()
}

// AddItem adds a new clipboard item if it doesn't already exist
func (m *Manager) AddItem(content string) bool {
	item := newClipboardItem(content)
	if !m.containsHash(item.Hash) {
		if m.dbClient != nil {
			entry := db.ClipboardEntry{
				Content:   item.Item,
				Hash:      item.Hash,
				Timestamp: item.TimeStamp,
				Pinned:    item.Pinned,
			}
			if err := m.dbClient.Insert(entry); err != nil {
				return false
			}
		}

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

		if m.dbClient != nil {
			if err := m.dbClient.Delete(item.Hash); err != nil {
				return false
			}
		}

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

// LoadFromDB loads history from the SQLite database
func (m *Manager) LoadFromDB() error {
	if m.dbClient == nil {
		return nil
	}
	entries, err := m.dbClient.LoadAll()
	if err != nil {
		return err
	}

	m.items = make([]ClipboardHistory, 0, len(entries))
	m.hashes = make(map[string]struct{})

	for _, entry := range entries {
		item := ClipboardHistory{
			Item:      entry.Content,
			Hash:      entry.Hash,
			TimeStamp: entry.Timestamp,
			Pinned:    entry.Pinned,
		}
		m.items = append(m.items, item)
		m.hashes[item.Hash] = struct{}{}
		m.lastHash = item.Hash
	}

	sortItems(m.items)
	return nil
}

// sortItems sorts in-place: pinned first, then by timestamp ascending.
func sortItems(items []ClipboardHistory) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Pinned != items[j].Pinned {
			return items[i].Pinned
		}
		return items[i].TimeStamp.Before(items[j].TimeStamp)
	})
}

// newClipboardItem creates a new clipboard history item
func newClipboardItem(content string) ClipboardHistory {
	return ClipboardHistory{
		Item:      content,
		Hash:      fmt.Sprintf("%x", sha256.Sum256([]byte(content))),
		TimeStamp: time.Now(),
	}
}

// TogglePin toggles the pinned state for an item by index
func (m *Manager) TogglePin(index int) error {
	if index >= 0 && index < len(m.items) {
		item := &m.items[index]
		newPinned := !item.Pinned
		if m.dbClient != nil {
			if err := m.dbClient.SetPinned(item.Hash, newPinned); err != nil {
				return err
			}
		}
		item.Pinned = newPinned
		sortItems(m.items)
		return nil
	}
	return fmt.Errorf("invalid index: %d", index)
}
