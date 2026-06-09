package session

import (
	"os"
	"testing"
	"time"
)

func TestManager_CreateAndGet(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-session-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	mgr, err := NewManager(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	// Create a session
	s, err := mgr.Create("test-id-1", "my-session")
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != "test-id-1" {
		t.Errorf("s.ID = %q, want test-id-1", s.ID)
	}
	if s.Name != "my-session" {
		t.Errorf("s.Name = %q, want my-session", s.Name)
	}

	// Get the session
	s2, err := mgr.Get("test-id-1")
	if err != nil {
		t.Fatal(err)
	}
	if s2 == nil {
		t.Fatal("s2 is nil")
	}
	if s2.Name != "my-session" {
		t.Errorf("s2.Name = %q, want my-session", s2.Name)
	}
}

func TestManager_GetByName(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-session-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	mgr, err := NewManager(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	// Create sessions
	mgr.Create("id-1", "session-one")
	mgr.Create("id-2", "session-two")

	// Get by name
	s, err := mgr.GetByName("session-one")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("session is nil")
	}
	if s.ID != "id-1" {
		t.Errorf("s.ID = %q, want id-1", s.ID)
	}

	// Get non-existent name
	s3, err := mgr.GetByName("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if s3 != nil {
		t.Error("expected nil for nonexistent session")
	}
}

func TestManager_List(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-session-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	mgr, err := NewManager(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	// List empty
	sessions, err := mgr.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Errorf("len(sessions) = %d, want 0", len(sessions))
	}

	// Create sessions
	mgr.Create("id-1", "first")
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	mgr.Create("id-2", "second")

	// List all
	sessions, err = mgr.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Errorf("len(sessions) = %d, want 2", len(sessions))
	}
}

func TestManager_UpdateLastActive(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-session-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	mgr, err := NewManager(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	s, _ := mgr.Create("id-1", "test")
	oldLastActive := s.LastActive

	time.Sleep(10 * time.Millisecond)

	err = mgr.UpdateLastActive("id-1")
	if err != nil {
		t.Fatal(err)
	}

	s2, _ := mgr.Get("id-1")
	if !s2.LastActive.After(oldLastActive) {
		t.Error("LastActive should be updated")
	}
}

func TestManager_UpdateContextSummary(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-session-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	mgr, err := NewManager(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mgr.Create("id-1", "test")

	err = mgr.UpdateContextSummary("id-1", "Working on API endpoints")
	if err != nil {
		t.Fatal(err)
	}

	s, _ := mgr.Get("id-1")
	if s.ContextSummary != "Working on API endpoints" {
		t.Errorf("ContextSummary = %q, want 'Working on API endpoints'", s.ContextSummary)
	}
}

func TestManager_Delete(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-session-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	mgr, err := NewManager(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	mgr.Create("id-1", "to-delete")

	s, _ := mgr.Get("id-1")
	if s == nil {
		t.Fatal("session should exist")
	}

	err = mgr.Delete("id-1")
	if err != nil {
		t.Fatal(err)
	}

	s2, _ := mgr.Get("id-1")
	if s2 != nil {
		t.Error("session should be deleted")
	}
}

func TestManager_Close(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "unibrow-session-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	mgr, err := NewManager(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	err = mgr.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}