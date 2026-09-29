package ui

import (
	"log"

	"github.com/atotto/clipboard"
	"github.com/bvdwalt/clippy/internal/pasteboard"
)

// clipboardSource is the system clipboard as seen by the tick handler.
type clipboardSource interface {
	ReadAll() (string, error)
	// ok is false when the platform has no change counter.
	ChangeCount() (count int, ok bool)
	IsConcealed() bool
}

type systemClipboard struct{}

func (systemClipboard) ReadAll() (string, error) { return clipboard.ReadAll() }

func (systemClipboard) ChangeCount() (int, bool) { return pasteboard.ChangeCount() }

// IsConcealed fails open: if the type lookup breaks, capture keeps working.
func (systemClipboard) IsConcealed() bool {
	concealed, err := pasteboard.IsConcealed()
	if err != nil {
		log.Printf("Failed to check clipboard types: %v", err)
	}
	return concealed
}
