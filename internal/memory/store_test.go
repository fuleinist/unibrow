package memory

import (
	"os"
	"testing"
)

func TestStore_AddAndList(t *testing.T) {
	// Create temp database
	tmpFile, err := os.CreateTemp("", "unibrow-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := NewStore(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Add entries
	entry1, err := store.Add("session1", "claude", "Remember: use context")
	if err != nil {
		t.Fatal(err)
	}
	if entry1.ID == 0 {
		t.Error("entry1.ID should not be 0")
	}
	if entry1.SessionID != "session1" {
		t.Errorf("entry1.SessionID = %q, want session1", entry1.SessionID)
	}
	if entry1.Agent != "claude" {
		t.Errorf("entry1.Agent = %q, want claude", entry1.Agent)
	}

	entry2, err := store.Add("session1", "user", "Test prompt")
	if err != nil {
		t.Fatal(err)
	}

	// List entries for session1
	entries, err := store.List("session1")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("len(entries) = %d, want 2", len(entries))
	}

	// List entries for session2 (should be empty)
	entries2, err := store.List("session2")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries2) != 0 {
		t.Errorf("len(entries2) = %d, want 0", len(entries2))
	}

	// Add entry for session2
	_, err = store.Add("session2", "gemini", "Another session")
	if err != nil {
		t.Fatal(err)
	}

	// ListAll should return 3 entries
	all, err := store.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Errorf("len(all) = %d, want 3", len(all))
	}

	// Suppress unused variable warning
	_ = entry2
}

func TestStore_Clear(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := NewStore(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Add entries
	store.Add("session1", "claude", "entry1")
	store.Add("session1", "claude", "entry2")
	store.Add("session2", "gemini", "entry3")

	// Clear session1
	err = store.Clear("session1")
	if err != nil {
		t.Fatal(err)
	}

	// session1 should be empty
	entries, _ := store.List("session1")
	if len(entries) != 0 {
		t.Errorf("session1 entries after clear = %d, want 0", len(entries))
	}

	// session2 should still have entry
	entries2, _ := store.List("session2")
	if len(entries2) != 1 {
		t.Errorf("session2 entries = %d, want 1", len(entries2))
	}
}

func TestStore_ClearAll(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := NewStore(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	store.Add("session1", "claude", "entry1")
	store.Add("session2", "gemini", "entry2")

	err = store.ClearAll()
	if err != nil {
		t.Fatal(err)
	}

	all, _ := store.ListAll()
	if len(all) != 0 {
		t.Errorf("all entries after ClearAll = %d, want 0", len(all))
	}
}

func TestStore_IncrementUsage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := NewStore(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	entry, _ := store.Add("session1", "claude", "test")
	if entry.UsageCount != 0 {
		t.Errorf("initial UsageCount = %d, want 0", entry.UsageCount)
	}

	err = store.IncrementUsage(entry.ID)
	if err != nil {
		t.Fatal(err)
	}

	entries, _ := store.List("session1")
	if len(entries) != 1 {
		t.Fatal("expected 1 entry")
	}
	if entries[0].UsageCount != 1 {
		t.Errorf("UsageCount after increment = %d, want 1", entries[0].UsageCount)
	}
}

func TestStore_GetRecent(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := NewStore(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Add 5 entries
	for i := 0; i < 5; i++ {
		store.Add("session1", "claude", "entry")
	}

	recent, err := store.GetRecent("session1", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(recent) != 3 {
		t.Errorf("len(recent) = %d, want 3", len(recent))
	}
}

// TestStore_Count verifies that Count returns the total number of
// memory entries for a session, not the number returned by a
// caller-side limit. `context show` relies on this to print
// "(Showing N of M entries)" with M being the real total — not N.
func TestStore_Count(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := NewStore(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Empty session: count must be 0, not an error.
	if n, err := store.Count("session1"); err != nil || n != 0 {
		t.Errorf("empty Count = (%d, %v), want (0, nil)", n, err)
	}

	// Add 3 entries for session1, 2 for session2.
	for i := 0; i < 3; i++ {
		if _, err := store.Add("session1", "claude", "x"); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if _, err := store.Add("session2", "user", "y"); err != nil {
			t.Fatal(err)
		}
	}

	if n, err := store.Count("session1"); err != nil || n != 3 {
		t.Errorf("session1 Count = (%d, %v), want (3, nil)", n, err)
	}
	if n, err := store.Count("session2"); err != nil || n != 2 {
		t.Errorf("session2 Count = (%d, %v), want (2, nil)", n, err)
	}

	// Clearing session1 drops its count to 0; session2 unaffected.
	if err := store.Clear("session1"); err != nil {
		t.Fatal(err)
	}
	if n, err := store.Count("session1"); err != nil || n != 0 {
		t.Errorf("session1 Count after Clear = (%d, %v), want (0, nil)", n, err)
	}
	if n, err := store.Count("session2"); err != nil || n != 2 {
		t.Errorf("session2 Count after session1 Clear = (%d, %v), want (2, nil)", n, err)
	}
}

func TestStore_Close(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := NewStore(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	err = store.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// TestStore_RemoveByID verifies that RemoveByID deletes exactly the
// row identified by (sessionID, id) and leaves siblings alone. It is
// the underlying primitive that `unibrow context remove <id>` calls
// — the CLI smoke tests cover parse + not-found paths; this one
// covers the actual delete semantics.
func TestStore_RemoveByID(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := NewStore(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	first, err := store.Add("session1", "context", "foo")
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Add("session1", "context", "bar")
	if err != nil {
		t.Fatal(err)
	}

	// Remove the first entry; should return (true, nil).
	removed, err := store.RemoveByID("session1", first.ID)
	if err != nil {
		t.Fatalf("RemoveByID first: %v", err)
	}
	if !removed {
		t.Error("RemoveByID(first) = false, want true")
	}

	entries, err := store.List("session1")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) after remove = %d, want 1", len(entries))
	}
	if entries[0].ID != second.ID {
		t.Errorf("remaining entry ID = %d, want %d", entries[0].ID, second.ID)
	}
	if entries[0].Content != "bar" {
		t.Errorf("remaining entry content = %q, want %q", entries[0].Content, "bar")
	}

	// Removing again should return (false, nil) — no error, just no match.
	removed, err = store.RemoveByID("session1", first.ID)
	if err != nil {
		t.Fatalf("RemoveByID (re-remove) error = %v, want nil", err)
	}
	if removed {
		t.Error("RemoveByID on already-removed id = true, want false")
	}

	// Cross-session isolation: an id belonging to session2 must NOT be
	// removable via session1, even if a row with that id exists in session2.
	if _, err := store.Add("session2", "context", "x"); err != nil {
		t.Fatal(err)
	}
	all, err := store.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("len(all) after cross-session add = %d, want 2", len(all))
	}
	// Pick the id from session2 by listing it.
	sess2, err := store.List("session2")
	if err != nil {
		t.Fatal(err)
	}
	crossID := sess2[0].ID
	removed, err = store.RemoveByID("session1", crossID)
	if err != nil {
		t.Fatalf("RemoveByID cross-session error = %v, want nil", err)
	}
	if removed {
		t.Error("RemoveByID cross-session = true, want false (wrong session)")
	}
	// The session2 row must still be there.
	sess2After, err := store.List("session2")
	if err != nil {
		t.Fatal(err)
	}
	if len(sess2After) != 1 {
		t.Errorf("len(session2) after cross-session remove = %d, want 1", len(sess2After))
	}
}

