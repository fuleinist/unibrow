package cmd

import (
	"path/filepath"
	"testing"

	"github.com/fuleinist/unibrow/internal/memory"
	"github.com/fuleinist/unibrow/internal/session"
)

func TestMemoryIntegration(t *testing.T) {
	// Use a temp dir for database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_memory.db")

	store, err := memory.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// Test Add
	entry, err := store.Add("test-session", "user", "hello world")
	if err != nil {
		t.Fatalf("store.Add failed: %v", err)
	}
	if entry.ID == 0 {
		t.Error("expected non-zero entry ID")
	}
	if entry.Content != "hello world" {
		t.Errorf("content = %q, want %q", entry.Content, "hello world")
	}

	// Test List
	entries, err := store.List("test-session")
	if err != nil {
		t.Fatalf("store.List failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("len(entries) = %d, want 1", len(entries))
	}

	// Test Clear
	err = store.Clear("test-session")
	if err != nil {
		t.Fatalf("store.Clear failed: %v", err)
	}
	entries, err = store.List("test-session")
	if err != nil {
		t.Fatalf("store.List after clear failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) after clear = %d, want 0", len(entries))
	}
}

func TestSessionManagerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_sessions.db")

	mgr, err := session.NewManager(dbPath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Close()

	// Test Create
	sess, err := mgr.Create("my-session", "My Session")
	if err != nil {
		t.Fatalf("mgr.Create failed: %v", err)
	}
	if sess.Name != "My Session" {
		t.Errorf("session name = %q, want %q", sess.Name, "My Session")
	}

	// Test List
	sessions, err := mgr.List()
	if err != nil {
		t.Fatalf("mgr.List failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("len(sessions) = %d, want 1", len(sessions))
	}

	// Test GetByName
	retrieved, err := mgr.GetByName("My Session")
	if err != nil {
		t.Fatalf("mgr.GetByName failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetByName returned nil")
	}
	if retrieved.Name != "My Session" {
		t.Errorf("retrieved name = %q, want %q", retrieved.Name, "My Session")
	}
}

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Error("GetVersion returned empty string")
	}
	if v != "0.1.0" {
		t.Errorf("GetVersion = %q, want %q", v, "0.1.0")
	}
}


