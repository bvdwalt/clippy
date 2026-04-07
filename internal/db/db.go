package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// ClipboardEntry represents a clipboard entry in the persistence layer
type ClipboardEntry struct {
	Content   string
	Hash      string
	Timestamp time.Time
	Pinned    bool
}

// DBClient is the interface implemented by all persistence backends.
type DBClient interface {
	Insert(entry ClipboardEntry) error
	Delete(hash string) error
	LoadAll() ([]ClipboardEntry, error)
	SetPinned(hash string, pinned bool) error
	Close() error
}

// Client handles database operations for clipboard history
type Client struct {
	db *sql.DB
}

// New creates a new database client with the given database path
func New(dbPath string) (*Client, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	client := &Client{db: db}

	if err := client.initialize(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("error initializing database: %w (also failed to close db: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("error initializing database: %w", err)
	}

	return client, nil
}

// initialize creates the necessary tables and runs migrations
func (c *Client) initialize() error {
	if err := c.migrate(); err != nil {
		return fmt.Errorf("error migrating schema: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS clipboard_history (
		hash TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		pinned INTEGER NOT NULL DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON clipboard_history(timestamp ASC);
	`

	_, err := c.db.Exec(schema)
	return err
}

// migrate handles schema migrations for existing databases
func (c *Client) migrate() error {
	// Check if table exists
	var tableExists bool
	row := c.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM sqlite_master
		WHERE type='table' AND name='clipboard_history'
	`)
	if err := row.Scan(&tableExists); err != nil || !tableExists {
		return nil
	}

	// Add pinned column if missing (migration from count-based schema)
	var hasPinned bool
	row = c.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('clipboard_history')
		WHERE name = 'pinned'
	`)
	if err := row.Scan(&hasPinned); err != nil {
		return err
	}
	if !hasPinned {
		_, err := c.db.Exec(`ALTER TABLE clipboard_history ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0`)
		return err
	}

	return nil
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Insert adds a new clipboard entry to the database
func (c *Client) Insert(entry ClipboardEntry) error {
	pinned := 0
	if entry.Pinned {
		pinned = 1
	}
	_, err := c.db.Exec(
		"INSERT INTO clipboard_history (hash, content, timestamp, pinned) VALUES (?, ?, ?, ?)",
		entry.Hash, entry.Content, entry.Timestamp, pinned,
	)
	return err
}

// Delete removes a clipboard entry by hash
func (c *Client) Delete(hash string) error {
	res, err := c.db.Exec("DELETE FROM clipboard_history WHERE hash = ?", hash)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("clip with hash %s not found", hash)
	}
	return nil
}

// LoadAll retrieves all clipboard entries ordered by timestamp ascending
func (c *Client) LoadAll() ([]ClipboardEntry, error) {
	rows, err := c.db.Query("SELECT content, hash, timestamp, pinned FROM clipboard_history ORDER BY timestamp ASC")
	if err != nil {
		return nil, fmt.Errorf("error querying history: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

	entries := make([]ClipboardEntry, 0)
	for rows.Next() {
		var entry ClipboardEntry
		var pinnedInt int
		if err := rows.Scan(&entry.Content, &entry.Hash, &entry.Timestamp, &pinnedInt); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		entry.Pinned = pinnedInt != 0
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// SetPinned updates the pinned state for a clipboard entry
func (c *Client) SetPinned(hash string, pinned bool) error {
	pinnedInt := 0
	if pinned {
		pinnedInt = 1
	}
	res, err := c.db.Exec("UPDATE clipboard_history SET pinned = ? WHERE hash = ?", pinnedInt, hash)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("clip with hash %s not found", hash)
	}
	return nil
}
