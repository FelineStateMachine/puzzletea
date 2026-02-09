package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS games (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL UNIQUE,
    game_type     TEXT    NOT NULL,
    mode          TEXT    NOT NULL,
    initial_state TEXT    NOT NULL,
    save_state    TEXT,
    status        TEXT    NOT NULL DEFAULT 'new'
                  CHECK(status IN ('new','in_progress','completed','abandoned')),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at  DATETIME
);`

type Store struct {
	db *sql.DB
}

// DefaultDBPath returns ~/.puzzletea/history.db.
func DefaultDBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".puzzletea", "history.db")
}

// Open creates the directory if needed, opens the database, creates the table,
// and enables WAL mode.
func Open(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating table: %w", err)
	}

	return &Store{db: db}, nil
}

// CreateGame inserts a new game record and sets rec.ID.
func (s *Store) CreateGame(rec *GameRecord) error {
	result, err := s.db.Exec(
		`INSERT INTO games (name, game_type, mode, initial_state, save_state, status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		rec.Name, rec.GameType, rec.Mode, rec.InitialState, rec.SaveState, rec.Status,
	)
	if err != nil {
		return fmt.Errorf("inserting game: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	rec.ID = id
	return nil
}

// UpdateSaveState updates save_state and updated_at for the given game.
func (s *Store) UpdateSaveState(id int64, saveState string) error {
	_, err := s.db.Exec(
		`UPDATE games SET save_state = ?, updated_at = ? WHERE id = ?`,
		saveState, time.Now().UTC(), id,
	)
	return err
}

// UpdateStatus updates the status (and completed_at if completed) for the given game.
func (s *Store) UpdateStatus(id int64, status GameStatus) error {
	if status == StatusCompleted {
		now := time.Now().UTC()
		_, err := s.db.Exec(
			`UPDATE games SET status = ?, completed_at = ?, updated_at = ? WHERE id = ?`,
			status, now, now, id,
		)
		return err
	}
	_, err := s.db.Exec(
		`UPDATE games SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now().UTC(), id,
	)
	return err
}

// NameExists checks whether a game with the given name already exists.
func (s *Store) NameExists(name string) (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM games WHERE name = ?`, name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
