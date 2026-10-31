package history

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bvdwalt/clippy/internal/db"
)

const (
	ConfigDir  = ".clippy"
	DBFileName = "clippy.db"

	// DefaultTTL is how long unpinned entries are kept
	DefaultTTL = 24 * time.Hour
)

// Manager handles clipboard history storage and management
type Manager struct {
	items    []ClipboardHistory
	hashes   map[string]struct{}
	lastHash string
	dbClient db.DBClient // nil for in-memory managers
	dbPath   string
	ttl      time.Duration // 0 disables expiry
}

// NewManager creates a new history manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigDir)
	if err := os.MkdirAll(configDir, 0700); err != nil {
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
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("error creating directory: %w", err)
	}

	dbClient, err := db.New(dbPath)
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

// DBPath returns the database file path, empty for in-memory managers
func (m *Manager) DBPath() string {
	return m.dbPath
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
			if err := m.dbClient.Delete(item.Hash); err != nil && !errors.Is(err, db.ErrNotFound) {
				return false
			}
		}

		m.removeAt(index)
		return true
	}
	return false
}

// removeAt drops the item at index from the list and hash set
func (m *Manager) removeAt(index int) {
	delete(m.hashes, m.items[index].Hash)
	if m.lastHash == m.items[index].Hash {
		m.lastHash = ""
	}
	m.items = append(m.items[:index], m.items[index+1:]...)
}

// Count returns the number of items in history
func (m *Manager) Count() int {
	return len(m.items)
}

// LoadFromDB loads history from the Automerge document
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
				if errors.Is(err, db.ErrNotFound) {
					// row was removed by another process; drop the stale entry
					m.removeAt(index)
					return nil
				}
				return err
			}
		}
		item.Pinned = newPinned
		sortItems(m.items)
		return nil
	}
	return fmt.Errorf("invalid index: %d", index)
}

// ParseTTL parses a TTL setting such as "24h" or "90m". An empty value yields
// DefaultTTL and "0" disables expiry.
func ParseTTL(s string) (time.Duration, error) {
	if s == "" {
		return DefaultTTL, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid TTL %q: %w", s, err)
	}
	if d < 0 {
		return 0, fmt.Errorf("invalid TTL %q: must not be negative", s)
	}
	return d, nil
}

// SetTTL sets how long unpinned entries are kept. Zero disables expiry.
func (m *Manager) SetTTL(ttl time.Duration) {
	m.ttl = ttl
}

// PruneExpired removes unpinned items older than the TTL and returns how many were removed
func (m *Manager) PruneExpired(now time.Time) int {
	if m.ttl <= 0 {
		return 0
	}
	cutoff := now.Add(-m.ttl)

	if m.dbClient != nil {
		if _, err := m.dbClient.DeleteExpired(cutoff); err != nil {
			log.Printf("Warning: could not delete expired entries: %v", err)
			return 0
		}
	}

	kept := m.items[:0]
	removed := 0
	for _, item := range m.items {
		if !item.Pinned && item.TimeStamp.Before(cutoff) {
			delete(m.hashes, item.Hash)
			if m.lastHash == item.Hash {
				m.lastHash = ""
			}
			removed++
			continue
		}
		kept = append(kept, item)
	}
	m.items = kept
	return removed
}
