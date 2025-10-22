package history

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const (
	ConfigDir  = ".clippy"
	DBFileName = "clippy.db"
)

// Manager handles clipboard history storage and management
type Manager struct {
	items    []ClipboardHistory
	hashes   map[string]struct{}
	lastHash string
	db       *sql.DB
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

// NewManagerWithPath creates a new history manager with a custom database path
// This is useful for testing with isolated databases
func NewManagerWithPath(dbPath string) (*Manager, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error creating directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	manager := &Manager{
		items:  make([]ClipboardHistory, 0),
		hashes: make(map[string]struct{}),
		db:     db,
		dbPath: dbPath,
	}

	if err := manager.initDB(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error initializing database: %w", err)
	}

	return manager, nil
}

// initDB creates the necessary tables if they don't exist
func (m *Manager) initDB() error {
	schema := `
	CREATE TABLE IF NOT EXISTS clipboard_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		hash TEXT NOT NULL UNIQUE,
		timestamp DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON clipboard_history(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_hash ON clipboard_history(hash);
	`

	_, err := m.db.Exec(schema)
	return err
}

// Close closes the database connection
func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// AddItem adds a new clipboard item if it doesn't already exist
func (m *Manager) AddItem(content string) bool {
	item := newClipboardItem(content)
	if !m.containsHash(item.Hash) {
		_, err := m.db.Exec(
			"INSERT INTO clipboard_history (content, hash, timestamp) VALUES (?, ?, ?)",
			item.Item, item.Hash, item.TimeStamp,
		)
		if err != nil {
			return false
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

		_, err := m.db.Exec("DELETE FROM clipboard_history WHERE hash = ?", item.Hash)
		if err != nil {
			return false
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
	rows, err := m.db.Query("SELECT content, hash, timestamp FROM clipboard_history ORDER BY timestamp ASC")
	if err != nil {
		return fmt.Errorf("error querying history: %w", err)
	}
	defer rows.Close()

	m.items = make([]ClipboardHistory, 0)
	m.hashes = make(map[string]struct{})

	for rows.Next() {
		var item ClipboardHistory
		if err := rows.Scan(&item.Item, &item.Hash, &item.TimeStamp); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}
		m.items = append(m.items, item)
		m.hashes[item.Hash] = struct{}{}
		m.lastHash = item.Hash
	}

	return rows.Err()
}

// newClipboardItem creates a new clipboard history item
func newClipboardItem(content string) ClipboardHistory {
	return ClipboardHistory{
		Item:      content,
		Hash:      fmt.Sprintf("%x", sha256.Sum256([]byte(content))),
		TimeStamp: time.Now(),
	}
}
