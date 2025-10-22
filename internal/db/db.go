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
	schema := `
	CREATE TABLE IF NOT EXISTS clipboard_history (
		hash TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		timestamp DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON clipboard_history(timestamp DESC);
	`

	_, err := c.db.Exec(schema)
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
	_, err := c.db.Exec(
		"INSERT INTO clipboard_history (hash, content, timestamp) VALUES (?, ?, ?)",
		entry.Hash, entry.Content, entry.Timestamp,
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
	rows, err := c.db.Query("SELECT content, hash, timestamp FROM clipboard_history ORDER BY timestamp ASC")
	if err != nil {
		return nil, fmt.Errorf("error querying history: %w", err)
	}
	defer rows.Close()

	entries := make([]ClipboardEntry, 0)
	for rows.Next() {
		var entry ClipboardEntry
		if err := rows.Scan(&entry.Content, &entry.Hash, &entry.Timestamp); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}
