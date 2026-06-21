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
// row with the given id in the given session, returns (true, nil)
// on success and (false, nil) when the id is unknown or belongs to a
// different session. Used by `context remove <id>`.
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

	// Three entries in session1, one in session2. IDs are 1, 2, 3, 4.
	for _, content := range []string{"a", "b", "c"} {
		if _, err := store.Add("session1", "context", content); err != nil {
			t.Fatal(err)
		}
	}
	other, err := store.Add("session2", "context", "x")
	if err != nil {
		t.Fatal(err)
	}

	// Removing an existing id returns (true, nil) and drops one row.
	removed, err := store.RemoveByID("session1", 2)
	if err != nil {
		t.Fatalf("RemoveByID(#2) error = %v", err)
	}
	if !removed {
		t.Error("RemoveByID(#2) returned removed=false, want true")
	}
	if n, _ := store.Count("session1"); n != 2 {
		t.Errorf("session1 count after remove = %d, want 2", n)
	}

	// Removing the same id again returns (false, nil) — not an error,
	// because the caller wants to distinguish "id not found" from
	// "store failure".
	removed, err = store.RemoveByID("session1", 2)
	if err != nil {
		t.Fatalf("RemoveByID(#2) second call error = %v", err)
	}
	if removed {
		t.Error("RemoveByID(#2) second call returned removed=true, want false")
	}

	// Removing an id that belongs to a different session returns
	// (false, nil) and leaves session2 untouched. This is the safe
	// minimum scoping rule for `context remove`: don't accidentally
	// delete an entry from another session just because IDs can
	// collide in theory.
	removed, err = store.RemoveByID("session1", other.ID)
	if err != nil {
		t.Fatalf("RemoveByID(cross-session) error = %v", err)
	}
	if removed {
		t.Error("RemoveByID(cross-session) returned removed=true, want false")
	}
	otherEntries, _ := store.List("session2")
	if len(otherEntries) != 1 {
		t.Errorf("session2 entries = %d, want 1 (untouched)", len(otherEntries))
	}
}