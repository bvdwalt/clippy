package db

import (
	"fmt"
	"os"
	"time"

	automerge "github.com/automerge/automerge-go"
)

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
		doc.Commit("Initialised clippy document")
	}

	return &AutomergeClient{doc: doc, path: path}, nil
}

func (c *AutomergeClient) Insert(entry ClipboardEntry) error {
	clippings := c.doc.Path(docPathName).List()
	length := clippings.Len()

	clip := map[string]interface{}{
		"content":   entry.Content,
		"hash":      entry.Hash,
		"timestamp": entry.Timestamp.Format(time.RFC3339Nano),
		"count":     float64(entry.Count),
	}

	if err := clippings.Insert(length, clip); err != nil {
		return fmt.Errorf("inserting clip: %w", err)
	}

	c.doc.Commit(fmt.Sprintf("Add clip %s", entry.Hash[:8]))
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
		h, err := automerge.As[string](c.doc.Path(docPathName, i, "hash").Get())
		if err != nil {
			continue
		}
		if h == hash {
			if err := clippings.Delete(i); err != nil {
				return fmt.Errorf("delete at index %d: %w", i, err)
			}
			c.doc.Commit(fmt.Sprintf("Delete clip %s", hash[:8]))
			return c.save()
		}
	}

	return fmt.Errorf("clip with hash %s not found", hash[:8])
}

func (c *AutomergeClient) LoadAll() ([]ClipboardEntry, error) {
	clippings := c.doc.Path(docPathName).List()
	length := clippings.Len()

	entries := make([]ClipboardEntry, 0, length)
	for i := 0; i < length; i++ {
		content, _ := automerge.As[string](c.doc.Path(docPathName, i, "content").Get())
		hash, _ := automerge.As[string](c.doc.Path(docPathName, i, "hash").Get())
		tsStr, _ := automerge.As[string](c.doc.Path(docPathName, i, "timestamp").Get())
		count, _ := automerge.As[float64](c.doc.Path(docPathName, i, "count").Get())

		ts, _ := time.Parse(time.RFC3339Nano, tsStr)

		entries = append(entries, ClipboardEntry{
			Content:   content,
			Hash:      hash,
			Timestamp: ts,
			Count:     int(count),
		})
	}

	return entries, nil
}

func (c *AutomergeClient) IncrementCount(hash string) error {
	clippings := c.doc.Path(docPathName).List()
	length := clippings.Len()

	for i := 0; i < length; i++ {
		h, _ := automerge.As[string](c.doc.Path(docPathName, i, "hash").Get())
		if h == hash {
			currentCount, _ := automerge.As[float64](c.doc.Path(docPathName, i, "count").Get())
			if err := c.doc.Path(docPathName, i, "count").Set(currentCount + 1); err != nil {
				return fmt.Errorf("set count: %w", err)
			}
			c.doc.Commit(fmt.Sprintf("Increment count for %s", hash[:8]))
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
