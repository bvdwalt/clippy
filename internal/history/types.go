package history

import "time"

// ClipboardHistory represents a single clipboard entry with metadata
type ClipboardHistory struct {
	Item      string    `json:"item"`
	Hash      string    `json:"hash"`
	TimeStamp time.Time `json:"timeStamp"`
	Count     int       `json:"count"`
}
