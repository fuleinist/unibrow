package memory

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides SQLite-backed memory storage.
type Store struct {
	db *sql.DB
}

// MemoryEntry represents a memory item.
type MemoryEntry struct {
	ID         int64
	SessionID  string
	Agent      string
	Content    string
	CreatedAt  time.Time
	UsageCount int
}

// NewStore creates a new memory store with the given database path.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("create tables: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying database connection.
func (s *Store) DB() *sql.DB {
	return s.db
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS memory (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		agent TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		usage_count INTEGER DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_memory_session ON memory(session_id);
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_active DATETIME DEFAULT CURRENT_TIMESTAMP,
		context_summary TEXT DEFAULT ''
	);
	`
	_, err := db.Exec(schema)
	return err
}

// Add adds a new memory entry.
func (s *Store) Add(sessionID, agent, content string) (*MemoryEntry, error) {
	result, err := s.db.Exec(
		`INSERT INTO memory (session_id, agent, content) VALUES (?, ?, ?)`,
		sessionID, agent, content,
	)
	if err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}

	id, _ := result.LastInsertId()
	return &MemoryEntry{
		ID:        id,
		SessionID: sessionID,
		Agent:     agent,
		Content:   content,
		CreatedAt: time.Now(),
	}, nil
}

// List returns all memory entries for a session.
func (s *Store) List(sessionID string) ([]MemoryEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, agent, content, created_at, usage_count
		 FROM memory WHERE session_id = ? ORDER BY created_at DESC`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		if err := rows.Scan(&e.ID, &e.SessionID, &e.Agent, &e.Content, &e.CreatedAt, &e.UsageCount); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// ListAll returns all memory entries across all sessions.
func (s *Store) ListAll() ([]MemoryEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, agent, content, created_at, usage_count
		 FROM memory ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		if err := rows.Scan(&e.ID, &e.SessionID, &e.Agent, &e.Content, &e.CreatedAt, &e.UsageCount); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// RemoveByID removes a single memory entry by id, scoped to the given
// session. Returns (true, nil) if a row was deleted, (false, nil) if no
// row matched the (session, id) pair — e.g. wrong session, or the id
// has already been removed. Errors are reserved for driver failures.
func (s *Store) RemoveByID(sessionID string, id int64) (bool, error) {
	result, err := s.db.Exec(
		`DELETE FROM memory WHERE session_id = ? AND id = ?`,
		sessionID, id,
	)
	if err != nil {
		return false, fmt.Errorf("delete by id: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}
	return n > 0, nil
}

// Clear removes all memory entries for a session.
func (s *Store) Clear(sessionID string) error {
	_, err := s.db.Exec(`DELETE FROM memory WHERE session_id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

// ClearAll removes all memory entries.
func (s *Store) ClearAll() error {
	_, err := s.db.Exec(`DELETE FROM memory`)
	if err != nil {
		return fmt.Errorf("delete all: %w", err)
	}
	return nil
}

// IncrementUsage increments the usage count for a memory entry.
func (s *Store) IncrementUsage(id int64) error {
	_, err := s.db.Exec(`UPDATE memory SET usage_count = usage_count + 1 WHERE id = ?`, id)
	return err
}

// Count returns the total number of memory entries for a session.
// Used by callers that paginate or truncate (e.g. `context show`)
// to report the total count vs. the number shown.
func (s *Store) Count(sessionID string) (int, error) {
	var n int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM memory WHERE session_id = ?`,
		sessionID,
	).Scan(&n); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return n, nil
}

// GetRecent returns the most recent memory entries for a session.
func (s *Store) GetRecent(sessionID string, limit int) ([]MemoryEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, agent, content, created_at, usage_count
		 FROM memory WHERE session_id = ? ORDER BY created_at DESC LIMIT ?`,
		sessionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		if err := rows.Scan(&e.ID, &e.SessionID, &e.Agent, &e.Content, &e.CreatedAt, &e.UsageCount); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}