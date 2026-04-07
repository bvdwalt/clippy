package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupClient(t *testing.T) (*Client, string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "clippy_db_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	path := filepath.Join(dir, "test.db")
	client, err := New(path)
	if err != nil {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			t.Logf("remove temp dir: %v", removeErr)
		}
		t.Fatalf("New: %v", err)
	}
	return client, path, func() {
		if err := client.Close(); err != nil {
			t.Logf("close client: %v", err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("remove temp dir: %v", err)
		}
	}
}

func makeEntry(content string) ClipboardEntry {
	return ClipboardEntry{
		Content:   content,
		Hash:      content + "-hash",
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestNew_CreatesEmptyDB(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	entries, err := client.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestNew_IdempotentOpen(t *testing.T) {
	_, path, cleanup := setupClient(t)
	defer cleanup()

	// Reopen same path — should not error
	client2, err := New(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if err := client2.Close(); err != nil {
		t.Logf("close client2: %v", err)
	}
}

func TestInsertAndLoadAll(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	entries := []ClipboardEntry{
		makeEntry("first"),
		makeEntry("second"),
		makeEntry("third"),
	}
	for _, e := range entries {
		if err := client.Insert(e); err != nil {
			t.Fatalf("Insert(%q): %v", e.Content, err)
		}
	}

	loaded, err := client.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(loaded) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(loaded))
	}
	for i, e := range entries {
		if loaded[i].Content != e.Content {
			t.Errorf("entry %d: content = %q, want %q", i, loaded[i].Content, e.Content)
		}
		if loaded[i].Hash != e.Hash {
			t.Errorf("entry %d: hash = %q, want %q", i, loaded[i].Hash, e.Hash)
		}
		if loaded[i].Pinned != e.Pinned {
			t.Errorf("entry %d: pinned = %v, want %v", i, loaded[i].Pinned, e.Pinned)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	client, path, cleanup := setupClient(t)
	defer cleanup()

	entry := makeEntry("hello")
	entry.Pinned = true
	if err := client.Insert(entry); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	client2, err := New(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() {
		if err := client2.Close(); err != nil {
			t.Logf("close client2: %v", err)
		}
	}()

	entries, err := client2.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll after reopen: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after reopen, got %d", len(entries))
	}
	e := entries[0]
	if e.Content != entry.Content {
		t.Errorf("content = %q, want %q", e.Content, entry.Content)
	}
	if e.Hash != entry.Hash {
		t.Errorf("hash = %q, want %q", e.Hash, entry.Hash)
	}
	if e.Pinned != entry.Pinned {
		t.Errorf("pinned = %v, want %v", e.Pinned, entry.Pinned)
	}
}

func TestDelete(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	a, b := makeEntry("alpha"), makeEntry("beta")
	if err := client.Insert(a); err != nil {
		t.Fatalf("Insert a: %v", err)
	}
	if err := client.Insert(b); err != nil {
		t.Fatalf("Insert b: %v", err)
	}

	if err := client.Delete(a.Hash); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	entries, _ := client.LoadAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after delete, got %d", len(entries))
	}
	if entries[0].Hash != b.Hash {
		t.Errorf("remaining entry hash = %q, want %q", entries[0].Hash, b.Hash)
	}
}

func TestDelete_NotFound(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	err := client.Delete("nonexistent-hash")
	if err == nil {
		t.Error("expected error deleting nonexistent hash, got nil")
	}
}

func TestSetPinned(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	entry := makeEntry("pinme")
	if err := client.Insert(entry); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	if err := client.SetPinned(entry.Hash, true); err != nil {
		t.Fatalf("SetPinned true: %v", err)
	}
	entries, _ := client.LoadAll()
	if !entries[0].Pinned {
		t.Error("expected pinned=true after SetPinned")
	}

	if err := client.SetPinned(entry.Hash, false); err != nil {
		t.Fatalf("SetPinned false: %v", err)
	}
	entries, _ = client.LoadAll()
	if entries[0].Pinned {
		t.Error("expected pinned=false after SetPinned")
	}
}

func TestSetPinned_NotFound(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	err := client.SetPinned("nonexistent-hash", true)
	if err == nil {
		t.Error("expected error setting pinned on nonexistent hash, got nil")
	}
}

func TestMigrate_AddsPinnedColumn(t *testing.T) {
	dir, err := os.MkdirTemp("", "clippy_db_migrate_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("remove temp dir: %v", err)
		}
	}()

	// Create a legacy schema (count column, no pinned)
	path := filepath.Join(dir, "legacy.db")
	legacyClient, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	_, err = legacyClient.Exec(`
		CREATE TABLE clipboard_history (
			hash TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			count INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("create legacy schema: %v", err)
	}
	_, err = legacyClient.Exec(`INSERT INTO clipboard_history (hash, content, timestamp, count) VALUES ('h1', 'old entry', '2024-01-01T00:00:00Z', 0)`)
	if err != nil {
		t.Fatalf("insert legacy entry: %v", err)
	}
	if err := legacyClient.Close(); err != nil {
		t.Logf("close legacy db: %v", err)
	}

	// Open with New — should migrate
	client, err := New(path)
	if err != nil {
		t.Fatalf("New (migration): %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("close client: %v", err)
		}
	}()

	entries, err := client.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll after migration: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after migration, got %d", len(entries))
	}
	if entries[0].Content != "old entry" {
		t.Errorf("content = %q, want %q", entries[0].Content, "old entry")
	}
	if entries[0].Pinned {
		t.Error("expected Pinned=false for migrated entry")
	}
}
