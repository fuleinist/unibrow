package cmd

import (
	"path/filepath"
	"strings"
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

// TestMemoryAddMultiWordArgs verifies that `memory add` stores the
// full prompt (joined with spaces) when multiple word args are
// passed, rather than silently truncating to the first word.
func TestMemoryAddMultiWordArgs(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	// getMemoryDBPath uses os.UserHomeDir which on Linux respects $HOME;
	// also clear XDG_CONFIG_HOME and other vars that may redirect.
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	words := []string{"hello", "world", "from", "test"}
	if err := memoryAdd(nil, words); err != nil {
		t.Fatalf("memoryAdd failed: %v", err)
	}

	// The store will have written to <tmpDir>/.unibrow/memory.db.
	store, err := memory.NewStore(filepath.Join(tmpDir, ".unibrow", "memory.db"))
	if err != nil {
		t.Fatalf("open store for verification: %v", err)
	}
	defer store.Close()

	entries, err := store.List("default")
	if err != nil {
		t.Fatalf("list entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	want := strings.Join(words, " ")
	if entries[0].Content != want {
		t.Errorf("entry content = %q, want %q", entries[0].Content, want)
	}
}

// TestContextAddMultiWordArgs verifies that `context add` stores the
// full prompt (joined with spaces) when multiple word args are
// passed, rather than silently truncating to the first word.
func TestContextAddMultiWordArgs(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	words := []string{"fix", "the", "bug"}
	if err := contextAdd(nil, words); err != nil {
		t.Fatalf("contextAdd failed: %v", err)
	}

	store, err := memory.NewStore(filepath.Join(tmpDir, ".unibrow", "memory.db"))
	if err != nil {
		t.Fatalf("open store for verification: %v", err)
	}
	defer store.Close()

	entries, err := store.List("default")
	if err != nil {
		t.Fatalf("list entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	want := strings.Join(words, " ")
	if entries[0].Content != want {
		t.Errorf("entry content = %q, want %q", entries[0].Content, want)
	}
	if entries[0].Agent != "context" {
		t.Errorf("entry agent = %q, want %q", entries[0].Agent, "context")
	}
}

// TestContextRemoveNotFound verifies that removing an id that does not
// exist in the current session is a non-fatal no-op (with a friendly
// message), not a hard error.
func TestContextRemoveNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	prevSession := contextSessionID
	contextSessionID = "remove-test"
	defer func() { contextSessionID = prevSession }()

	if err := contextRemove(nil, []string{"9999"}); err != nil {
		t.Fatalf("contextRemove on missing id should be non-fatal, got: %v", err)
	}
}

// TestContextRemoveInvalidID verifies that a non-numeric id returns a
// friendly error mentioning the bad input, instead of panicking or
// silently no-op'ing.
func TestContextRemoveInvalidID(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	err := contextRemove(nil, []string{"not-a-number"})
	if err == nil {
		t.Fatal("contextRemove with non-numeric id should return an error")
	}
	if !strings.Contains(err.Error(), "not-a-number") {
		t.Errorf("error = %q, want it to mention the bad input", err.Error())
	}
}


