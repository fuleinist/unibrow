package session

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Manager handles session persistence.
type Manager struct {
	db *sql.DB
}

// Session represents a user session.
type Session struct {
	ID             string
	Name           string
	CreatedAt      time.Time
	LastActive     time.Time
	ContextSummary string
}

// NewManager creates a new session manager.
func NewManager(dbPath string) (*Manager, error) {
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

	return &Manager{db: db}, nil
}

// Close closes the database.
func (m *Manager) Close() error {
	return m.db.Close()
}

func createTables(db *sql.DB) error {
	schema := `
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

// Create creates a new session.
func (m *Manager) Create(id, name string) (*Session, error) {
	now := time.Now()
	_, err := m.db.Exec(
		`INSERT OR REPLACE INTO sessions (id, name, created_at, last_active) VALUES (?, ?, ?, ?)`,
		id, name, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &Session{
		ID:         id,
		Name:       name,
		CreatedAt:  now,
		LastActive: now,
	}, nil
}

// Get retrieves a session by ID.
func (m *Manager) Get(id string) (*Session, error) {
	var s Session
	err := m.db.QueryRow(
		`SELECT id, name, created_at, last_active, context_summary FROM sessions WHERE id = ?`,
		id,
	).Scan(&s.ID, &s.Name, &s.CreatedAt, &s.LastActive, &s.ContextSummary)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &s, nil
}

// GetByName retrieves a session by name.
func (m *Manager) GetByName(name string) (*Session, error) {
	var s Session
	err := m.db.QueryRow(
		`SELECT id, name, created_at, last_active, context_summary FROM sessions WHERE name = ?`,
		name,
	).Scan(&s.ID, &s.Name, &s.CreatedAt, &s.LastActive, &s.ContextSummary)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session by name: %w", err)
	}
	return &s, nil
}

// List returns all sessions.
func (m *Manager) List() ([]Session, error) {
	rows, err := m.db.Query(
		`SELECT id, name, created_at, last_active, context_summary FROM sessions ORDER BY last_active DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt, &s.LastActive, &s.ContextSummary); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// UpdateLastActive updates the last active timestamp.
func (m *Manager) UpdateLastActive(id string) error {
	_, err := m.db.Exec(
		`UPDATE sessions SET last_active = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

// UpdateContextSummary updates the context summary.
func (m *Manager) UpdateContextSummary(id, summary string) error {
	_, err := m.db.Exec(
		`UPDATE sessions SET context_summary = ?, last_active = ? WHERE id = ?`,
		summary, time.Now(), id,
	)
	return err
}

// Delete removes a session.
func (m *Manager) Delete(id string) error {
	_, err := m.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}