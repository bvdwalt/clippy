package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupClient(t *testing.T) (*AutomergeClient, string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "clippy_db_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	path := filepath.Join(dir, "test.automerge")
	client, err := NewAutomergeClient(path)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("NewAutomergeClient: %v", err)
	}
	return client, path, func() {
		client.Close()
		os.RemoveAll(dir)
	}
}

func makeEntry(content string) ClipboardEntry {
	return ClipboardEntry{
		Content:   content,
		Hash:      content + "-hash", // stable fake hash for tests
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestNewAutomergeClient_CreatesEmptyDoc(t *testing.T) {
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

func TestNewAutomergeClient_CorruptFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "clippy_db_test")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "corrupt.automerge")
	if err := os.WriteFile(path, []byte("not valid automerge data"), 0644); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	_, err = NewAutomergeClient(path)
	if err == nil {
		t.Error("expected error loading corrupt file, got nil")
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
		if !loaded[i].Timestamp.Equal(e.Timestamp) {
			t.Errorf("entry %d: timestamp = %v, want %v", i, loaded[i].Timestamp, e.Timestamp)
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

	// Reopen from disk
	client2, err := NewAutomergeClient(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer client2.Close()

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
	if !e.Timestamp.Equal(entry.Timestamp) {
		t.Errorf("timestamp = %v, want %v", e.Timestamp, entry.Timestamp)
	}
	if e.Pinned != entry.Pinned {
		t.Errorf("pinned = %v, want %v", e.Pinned, entry.Pinned)
	}
}

func TestDelete(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	a, b := makeEntry("alpha"), makeEntry("beta")
	client.Insert(a)
	client.Insert(b)

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
	client.Insert(entry)

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

// TestDecodeEntry_MissingPinnedField verifies that documents written before the
// "pinned" field existed (where the field is absent) decode gracefully to
// Pinned=false rather than panicking or erroring.
func TestDecodeEntry_MissingPinnedField(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	// Insert a raw entry without the "pinned" field to simulate an old document.
	clippings := client.doc.Path(docPathName).List()
	err := clippings.Insert(clippings.Len(), map[string]interface{}{
		"content":   "old entry",
		"hash":      "old-hash",
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z07:00"),
	})
	if err != nil {
		t.Fatalf("insert raw entry: %v", err)
	}
	client.doc.Commit("legacy entry without pinned field")

	entries, err := client.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Pinned {
		t.Error("expected Pinned=false for entry missing the pinned field")
	}
	if entries[0].Content != "old entry" {
		t.Errorf("content = %q, want %q", entries[0].Content, "old entry")
	}
}
