package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// ClipboardEntry represents a clipboard entry in the database
type ClipboardEntry struct {
	Content   string
	Hash      string
	Timestamp time.Time
	Count     int
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
		db.Close()
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
		count INTEGER NOT NULL DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON clipboard_history(timestamp DESC);
	`

	_, err := c.db.Exec(schema)
	return err
}

// migrate handles schema migrations
func (c *Client) migrate() error {
	// Check if count column exists
	var hasCount bool
	row := c.db.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM pragma_table_info('clipboard_history') 
		WHERE name = 'count'
	`)
	if err := row.Scan(&hasCount); err != nil {
		// Table doesn't exist yet, no migration needed
		return nil
	}

	if hasCount {
		// Already has count column
		return nil
	}

	// Check if table exists at all
	var tableExists bool
	row = c.db.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM sqlite_master 
		WHERE type='table' AND name='clipboard_history'
	`)
	if err := row.Scan(&tableExists); err != nil {
		return err
	}

	if !tableExists {
		// Table doesn't exist yet, no migration needed
		return nil
	}

	// Add count column to existing table
	_, err := c.db.Exec(`ALTER TABLE clipboard_history ADD COLUMN count INTEGER NOT NULL DEFAULT 0`)
	return err
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
	count := entry.Count
	_, err := c.db.Exec(
		"INSERT INTO clipboard_history (hash, content, timestamp, count) VALUES (?, ?, ?, ?)",
		entry.Hash, entry.Content, entry.Timestamp, count,
	)
	return err
}

// Delete removes a clipboard entry by hash
func (c *Client) Delete(hash string) error {
	_, err := c.db.Exec("DELETE FROM clipboard_history WHERE hash = ?", hash)
	return err
}

// LoadAll retrieves all clipboard entries ordered by timestamp ascending
func (c *Client) LoadAll() ([]ClipboardEntry, error) {
	rows, err := c.db.Query("SELECT content, hash, timestamp, count FROM clipboard_history ORDER BY count DESC, timestamp DESC")
	if err != nil {
		return nil, fmt.Errorf("error querying history: %w", err)
	}
	defer rows.Close()

	entries := make([]ClipboardEntry, 0)
	for rows.Next() {
		var entry ClipboardEntry
		if err := rows.Scan(&entry.Content, &entry.Hash, &entry.Timestamp, &entry.Count); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// IncrementCount increments the copy count for a clipboard entry
func (c *Client) IncrementCount(hash string) error {
	_, err := c.db.Exec("UPDATE clipboard_history SET count = count + 1 WHERE hash = ?", hash)
	return err
}
