package storage

import (
	"errors"
	"path/filepath"
	"testing"

	"navigation/internal/domain"
)

func newSQLiteNoteTestStore(t *testing.T) *SQLiteSiteStore {
	t.Helper()
	dir := t.TempDir()
	store, err := NewSQLiteSiteStore(filepath.Join(dir, "sites.db"), filepath.Join(dir, "sites.json"))
	if err != nil {
		t.Fatalf("NewSQLiteSiteStore() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})
	return store
}

func TestSQLiteNoteStoreLifecycle(t *testing.T) {
	store := newSQLiteNoteTestStore(t)
	note := domain.Note{
		ID:         "note_1",
		Title:      "标题",
		FilePath:   "notes/2026/05/note_1.md",
		Summary:    "摘要",
		Tags:       []string{"idea", "work"},
		Status:     domain.NoteStatusActive,
		Visibility: domain.NoteVisibilityPublic,
		Pinned:     true,
		CreatedAt:  "2026-05-28T00:00:00Z",
		UpdatedAt:  "2026-05-28T00:00:00Z",
	}
	if err := store.CreateNote(note); err != nil {
		t.Fatalf("CreateNote() error = %v", err)
	}

	notes, err := store.ListNotes("", "", true)
	if err != nil {
		t.Fatalf("ListNotes() error = %v", err)
	}
	if len(notes) != 1 || len(notes[0].Tags) != 2 || !notes[0].Pinned || notes[0].Visibility != domain.NoteVisibilityPublic {
		t.Fatalf("notes = %#v, want created note with tags and pinned", notes)
	}

	note.Title = "新标题"
	note.Pinned = false
	note.Tags = []string{"updated"}
	if err := store.UpdateNote(note); err != nil {
		t.Fatalf("UpdateNote() error = %v", err)
	}
	got, err := store.GetNote("note_1")
	if err != nil {
		t.Fatalf("GetNote() error = %v", err)
	}
	if got.Title != "新标题" || got.Pinned || len(got.Tags) != 1 || got.Tags[0] != "updated" {
		t.Fatalf("got = %#v, want updated note", got)
	}

	if err := store.SoftDeleteNote("note_1", "2026-05-29T00:00:00Z"); err != nil {
		t.Fatalf("SoftDeleteNote() error = %v", err)
	}
	notes, err = store.ListNotes("", "", true)
	if err != nil {
		t.Fatalf("ListNotes() after delete error = %v", err)
	}
	if len(notes) != 0 {
		t.Fatalf("active notes = %d, want 0", len(notes))
	}
}

func TestSQLiteNoteStoreMissingRows(t *testing.T) {
	store := newSQLiteNoteTestStore(t)
	if _, err := store.GetNote("missing"); err == nil {
		t.Fatal("GetNote() error = nil, want error")
	}
	if err := store.UpdateNote(domain.Note{ID: "missing"}); err == nil {
		t.Fatal("UpdateNote() error = nil, want error")
	}
	if err := store.SoftDeleteNote("missing", "2026-05-29T00:00:00Z"); err == nil || errors.Is(err, nil) {
		t.Fatalf("SoftDeleteNote() error = %v, want error", err)
	}
}

func TestSQLiteNoteStoreRebuildNoteIndex(t *testing.T) {
	store := newSQLiteNoteTestStore(t)
	if err := store.CreateNote(domain.Note{
		ID:         "note_1",
		Title:      "旧标题",
		FilePath:   "notes/2026/05/note_1.md",
		Status:     domain.NoteStatusActive,
		Pinned:     true,
		Visibility: domain.NoteVisibilityPublic,
		CreatedAt:  "2026-05-01T00:00:00Z",
		UpdatedAt:  "2026-05-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("CreateNote() error = %v", err)
	}
	if err := store.CreateNote(domain.Note{
		ID:        "missing",
		Title:     "缺失文件",
		FilePath:  "notes/2026/05/missing.md",
		Status:    domain.NoteStatusActive,
		CreatedAt: "2026-05-01T00:00:00Z",
		UpdatedAt: "2026-05-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("CreateNote(missing) error = %v", err)
	}

	if err := store.RebuildNoteIndex([]domain.Note{{
		ID:        "note_1",
		Title:     "新标题",
		FilePath:  "notes/2026/05/note_1.md",
		Summary:   "新摘要",
		Status:    domain.NoteStatusActive,
		CreatedAt: "2026-05-28T00:00:00Z",
		UpdatedAt: "2026-05-28T00:00:00Z",
	}}); err != nil {
		t.Fatalf("RebuildNoteIndex() error = %v", err)
	}

	notes, err := store.ListNotes("", "", true)
	if err != nil {
		t.Fatalf("ListNotes() error = %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("notes = %d, want 1", len(notes))
	}
	if notes[0].Title != "新标题" || !notes[0].Pinned || notes[0].Visibility != domain.NoteVisibilityPublic || notes[0].CreatedAt != "2026-05-01T00:00:00Z" {
		t.Fatalf("note = %#v, want rebuilt note preserving pin and created_at", notes[0])
	}
}
