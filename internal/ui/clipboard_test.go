package ui

import (
	"testing"
	"time"
)

type fakeClipboard struct {
	content   string
	count     int
	countOK   bool
	concealed bool
	reads     int
}

func (f *fakeClipboard) ReadAll() (string, error) {
	f.reads++
	return f.content, nil
}

func (f *fakeClipboard) ChangeCount() (int, bool) { return f.count, f.countOK }

func (f *fakeClipboard) IsConcealed() bool { return f.concealed }

func tick(t *testing.T, m Model) Model {
	t.Helper()
	updated, _ := m.Update(TickMsg(time.Now()))
	return updated.(Model)
}

func TestTickSkipsConcealedContent(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)
	fake := &fakeClipboard{content: "hunter2", concealed: true}
	model.source = fake

	model = tick(t, model)

	if historyManager.Count() != 0 {
		t.Errorf("expected concealed content not to be stored, got %d items", historyManager.Count())
	}

	// A later plain copy is still captured
	fake.content = "plain text"
	fake.concealed = false
	tick(t, model)

	if historyManager.Count() != 1 {
		t.Errorf("expected plain content to be stored, got %d items", historyManager.Count())
	}
}

func TestTickSkipsReadWhenChangeCountUnchanged(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)
	fake := &fakeClipboard{content: "first", count: 5, countOK: true}
	model.source = fake

	model = tick(t, model)
	model = tick(t, model)

	if fake.reads != 1 {
		t.Errorf("expected 1 read while change count is unchanged, got %d", fake.reads)
	}

	fake.content = "second"
	fake.count = 6
	tick(t, model)

	if fake.reads != 2 {
		t.Errorf("expected a read after change count changed, got %d reads", fake.reads)
	}
	if historyManager.Count() != 2 {
		t.Errorf("expected 2 items, got %d", historyManager.Count())
	}
}

func TestTickReadsEveryTickWithoutChangeCount(t *testing.T) {
	historyManager, cleanup := setupTestHistoryManager(t)
	defer cleanup()
	model := NewModel(historyManager)
	fake := &fakeClipboard{content: "same"}
	model.source = fake

	model = tick(t, model)
	tick(t, model)

	if fake.reads != 2 {
		t.Errorf("expected a read on every tick, got %d", fake.reads)
	}
}
