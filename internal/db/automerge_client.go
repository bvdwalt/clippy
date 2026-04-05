package db

import (
	"fmt"
	"os"
	"time"

	automerge "github.com/automerge/automerge-go"
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

const docPathName string = "clippings"

type AutomergeClient struct {
	doc  *automerge.Doc
	path string
}

func NewAutomergeClient(path string) (*AutomergeClient, error) {
	var doc *automerge.Doc

	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		doc, err = automerge.Load(data)
		if err != nil {
			return nil, fmt.Errorf("corrupted automerge file %s: %w", path, err)
		}
	} else {
		doc = automerge.New()
		if err := doc.Path(docPathName).Set([]interface{}{}); err != nil {
			return nil, fmt.Errorf("init document: %w", err)
		}
		if _, err := doc.Commit("Initialised clippy document"); err != nil {
			return nil, fmt.Errorf("init commit: %w", err)
		}
	}

	return &AutomergeClient{doc: doc, path: path}, nil
}

// pathAs reads a typed value from the document at the given path segments.
// Returns the zero value of T on error.
func pathAs[T any](doc *automerge.Doc, path ...any) T {
	v, _ := automerge.As[T](doc.Path(path...).Get())
	return v
}

// decodeEntry reads a ClipboardEntry from the document at list index i.
func (c *AutomergeClient) decodeEntry(i int) ClipboardEntry {
	tsStr := pathAs[string](c.doc, docPathName, i, "timestamp")
	ts, _ := time.Parse(time.RFC3339Nano, tsStr)
	return ClipboardEntry{
		Content:   pathAs[string](c.doc, docPathName, i, "content"),
		Hash:      pathAs[string](c.doc, docPathName, i, "hash"),
		Timestamp: ts,
		Pinned:    pathAs[bool](c.doc, docPathName, i, "pinned"),
	}
}

func (c *AutomergeClient) Insert(entry ClipboardEntry) error {
	clippings := c.doc.Path(docPathName).List()
	length := clippings.Len()

	clip := map[string]interface{}{
		"content":   entry.Content,
		"hash":      entry.Hash,
		"timestamp": entry.Timestamp.Format(time.RFC3339Nano),
		"pinned":    entry.Pinned,
	}

	if err := clippings.Insert(length, clip); err != nil {
		return fmt.Errorf("inserting clip: %w", err)
	}

	if _, err := c.doc.Commit(fmt.Sprintf("Add clip %s", entry.Hash[:8])); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return c.save()
}

func (c *AutomergeClient) Delete(hash string) error {
	val, err := c.doc.Path(docPathName).Get()
	if err != nil {
		return fmt.Errorf("get clippings: %w", err)
	}
	clippings := val.List()
	length := clippings.Len()

	for i := 0; i < length; i++ {
		if pathAs[string](c.doc, docPathName, i, "hash") == hash {
			if err := clippings.Delete(i); err != nil {
				return fmt.Errorf("delete at index %d: %w", i, err)
			}
			if _, err := c.doc.Commit(fmt.Sprintf("Delete clip %s", hash[:8])); err != nil {
				return fmt.Errorf("commit: %w", err)
			}
			return c.save()
		}
	}

	return fmt.Errorf("clip with hash %s not found", hash[:8])
}

func (c *AutomergeClient) LoadAll() ([]ClipboardEntry, error) {
	length := c.doc.Path(docPathName).List().Len()

	entries := make([]ClipboardEntry, 0, length)
	for i := 0; i < length; i++ {
		entries = append(entries, c.decodeEntry(i))
	}

	return entries, nil
}

func (c *AutomergeClient) SetPinned(hash string, pinned bool) error {
	length := c.doc.Path(docPathName).List().Len()

	for i := 0; i < length; i++ {
		if pathAs[string](c.doc, docPathName, i, "hash") == hash {
			if err := c.doc.Path(docPathName, i, "pinned").Set(pinned); err != nil {
				return fmt.Errorf("set pinned: %w", err)
			}
			if _, err := c.doc.Commit(fmt.Sprintf("Set pinned=%v for %s", pinned, hash[:8])); err != nil {
				return fmt.Errorf("commit: %w", err)
			}
			return c.save()
		}
	}

	return fmt.Errorf("clip with hash %s not found", hash[:8])
}

func (c *AutomergeClient) Close() error {
	return c.save()
}

func (c *AutomergeClient) Doc() *automerge.Doc {
	return c.doc
}

func (c *AutomergeClient) save() error {
	data := c.doc.Save()

	tmpPath := c.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmpPath, c.path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
