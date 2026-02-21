// Package store provides persistent storage for saved games.
package store

import (
	"database/sql"
	"errors"
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

const createCategoryStatsViewSQL = `
CREATE VIEW IF NOT EXISTS category_stats AS
SELECT
    game_type,
    COUNT(*)                                              AS total_attempts,
    COUNT(*) FILTER (WHERE status = 'completed')          AS total_victories,
    COUNT(*) FILTER (WHERE status = 'abandoned')          AS total_abandoned,
    COUNT(*) FILTER (WHERE name LIKE 'Daily %')           AS times_daily,
    COUNT(*) FILTER (WHERE status = 'completed'
                       AND name LIKE 'Daily %')           AS daily_victories,
    (SELECT mode FROM games g2
     WHERE g2.game_type = games.game_type
       AND g2.status = 'completed'
     GROUP BY mode
     ORDER BY COUNT(*) DESC, mode ASC
     LIMIT 1)                                             AS preferred_mode
FROM games
GROUP BY game_type;`

const createModeStatsViewSQL = `
CREATE VIEW IF NOT EXISTS mode_stats AS
SELECT
    game_type,
    mode,
    COUNT(*) FILTER (WHERE status = 'completed')          AS victories,
    COUNT(*) FILTER (WHERE status = 'completed'
                       AND name LIKE 'Daily %')           AS daily_victories
FROM games
GROUP BY game_type, mode;`

type Store struct {
	db *sql.DB
}

const gameSelectColumns = `id, name, game_type, mode, initial_state, save_state, status,
        created_at, updated_at, completed_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanGameRecord(scanner rowScanner) (GameRecord, error) {
	var g GameRecord
	var completedAt sql.NullTime
	if err := scanner.Scan(
		&g.ID, &g.Name, &g.GameType, &g.Mode,
		&g.InitialState, &g.SaveState, &g.Status,
		&g.CreatedAt, &g.UpdatedAt, &completedAt,
	); err != nil {
		return GameRecord{}, err
	}
	if completedAt.Valid {
		g.CompletedAt = &completedAt.Time
	}
	return g, nil
}

func (s *Store) listGamesByQuery(query string, args ...any) ([]GameRecord, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying games: %w", err)
	}
	defer rows.Close()

	var games []GameRecord
	for rows.Next() {
		g, err := scanGameRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning game row: %w", err)
		}
		games = append(games, g)
	}
	return games, rows.Err()
}

func (s *Store) getGameByQuery(query string, args ...any) (*GameRecord, error) {
	g, err := scanGameRecord(s.db.QueryRow(query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
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

	if _, err := db.Exec(createCategoryStatsViewSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating category_stats view: %w", err)
	}

	if _, err := db.Exec(createModeStatsViewSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating mode_stats view: %w", err)
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

// ListGames returns all non-abandoned games ordered by most recently updated.
func (s *Store) ListGames() ([]GameRecord, error) {
	return s.listGamesByQuery(
		fmt.Sprintf(`SELECT %s
		 FROM games
		 WHERE status != ?
		 ORDER BY updated_at DESC`, gameSelectColumns),
		StatusAbandoned,
	)
}

// GetDailyGame looks up a daily game by name, including abandoned ones.
// Daily puzzles should always be resumable regardless of status.
// Returns nil, nil if no matching game is found.
func (s *Store) GetDailyGame(name string) (*GameRecord, error) {
	g, err := s.getGameByQuery(
		fmt.Sprintf(`SELECT %s
		 FROM games
		 WHERE name = ?`, gameSelectColumns),
		name,
	)
	if err != nil {
		return nil, fmt.Errorf("querying daily game: %w", err)
	}
	return g, nil
}

// GetGameByName looks up a single non-abandoned game by its unique name.
// Returns nil, nil if no matching game is found.
func (s *Store) GetGameByName(name string) (*GameRecord, error) {
	g, err := s.getGameByQuery(
		fmt.Sprintf(`SELECT %s
		 FROM games
		 WHERE name = ? AND status != ?`, gameSelectColumns),
		name, StatusAbandoned,
	)
	if err != nil {
		return nil, fmt.Errorf("querying game by name: %w", err)
	}
	return g, nil
}

// ListAllGames returns all games (including abandoned) ordered by most recently updated.
func (s *Store) ListAllGames() ([]GameRecord, error) {
	return s.listGamesByQuery(
		fmt.Sprintf(`SELECT %s
		 FROM games
		 ORDER BY updated_at DESC`, gameSelectColumns),
	)
}

func (s *Store) Close() error {
	return s.db.Close()
}
