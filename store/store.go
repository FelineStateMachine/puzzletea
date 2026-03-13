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
    game_id       TEXT    NOT NULL DEFAULT '',
    game_type     TEXT    NOT NULL,
    mode_id       TEXT    NOT NULL DEFAULT '',
    mode          TEXT    NOT NULL,
    initial_state TEXT    NOT NULL,
    save_state    TEXT,
    status        TEXT    NOT NULL DEFAULT 'new'
                  CHECK(status IN ('new','in_progress','completed','abandoned')),
    run_kind      TEXT    NOT NULL DEFAULT 'normal'
                  CHECK(run_kind IN ('normal','daily','weekly','seeded')),
    run_date      DATE,
    week_year     INTEGER,
    week_number   INTEGER,
    week_index    INTEGER,
    seed_text     TEXT,
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
    COUNT(*) FILTER (WHERE run_kind = 'daily')            AS times_daily,
    COUNT(*) FILTER (WHERE status = 'completed'
                       AND run_kind = 'daily')            AS daily_victories,
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
                       AND run_kind = 'daily')            AS daily_victories
FROM games
GROUP BY game_type, mode;`

type Store struct {
	db *sql.DB
}

const gameSelectColumns = `id, name, game_id, game_type, mode_id, mode, initial_state, save_state,
        status, run_kind, run_date, week_year, week_number, week_index, seed_text,
        created_at, updated_at, completed_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanGameRecord(scanner rowScanner) (GameRecord, error) {
	var g GameRecord
	var runDate sql.NullTime
	var weekYear sql.NullInt64
	var weekNumber sql.NullInt64
	var weekIndex sql.NullInt64
	var seedText sql.NullString
	var completedAt sql.NullTime
	if err := scanner.Scan(
		&g.ID, &g.Name, &g.GameID, &g.GameType, &g.ModeID, &g.Mode,
		&g.InitialState, &g.SaveState, &g.Status,
		&g.RunKind, &runDate, &weekYear, &weekNumber, &weekIndex, &seedText,
		&g.CreatedAt, &g.UpdatedAt, &completedAt,
	); err != nil {
		return GameRecord{}, err
	}
	if runDate.Valid {
		t := runDate.Time
		g.RunDate = &t
	}
	if weekYear.Valid {
		g.WeekYear = int(weekYear.Int64)
	}
	if weekNumber.Valid {
		g.WeekNumber = int(weekNumber.Int64)
	}
	if weekIndex.Valid {
		g.WeekIndex = int(weekIndex.Int64)
	}
	if seedText.Valid {
		g.SeedText = seedText.String
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

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// CreateGame inserts a new game record and sets rec.ID.
func (s *Store) CreateGame(rec *GameRecord) error {
	if rec.GameID == "" {
		rec.GameID = CanonicalGameID(rec.GameType)
	}
	if rec.ModeID == "" {
		rec.ModeID = CanonicalModeID(rec.Mode)
	}
	if rec.RunKind == "" {
		rec.RunKind = RunKindForName(rec.Name)
	}
	if rec.RunDate == nil {
		rec.RunDate = RunDateForName(rec.Name)
	}
	if rec.SeedText == "" {
		rec.SeedText = SeedTextForName(rec.Name)
	}
	if rec.WeekYear == 0 || rec.WeekNumber == 0 || rec.WeekIndex == 0 {
		if year, week, index, ok := WeeklyIdentityForName(rec.Name); ok {
			rec.WeekYear = year
			rec.WeekNumber = week
			rec.WeekIndex = index
		}
	}

	result, err := s.db.Exec(
		`INSERT INTO games
		    (name, game_id, game_type, mode_id, mode, initial_state, save_state, status,
		     run_kind, run_date, week_year, week_number, week_index, seed_text)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.Name, rec.GameID, rec.GameType, rec.ModeID, rec.Mode,
		rec.InitialState, rec.SaveState, rec.Status, string(rec.RunKind), rec.RunDate,
		nullableInt(rec.WeekYear), nullableInt(rec.WeekNumber), nullableInt(rec.WeekIndex), nullableString(rec.SeedText),
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
	return s.GetDeterministicGame(name)
}

// GetDeterministicGame looks up an exact deterministic game name, including
// abandoned rows.
func (s *Store) GetDeterministicGame(name string) (*GameRecord, error) {
	g, err := s.getGameByQuery(
		fmt.Sprintf(`SELECT %s
		 FROM games
		 WHERE name = ?`, gameSelectColumns),
		name,
	)
	if err != nil {
		return nil, fmt.Errorf("querying deterministic game: %w", err)
	}
	return g, nil
}

// GetWeeklyGame looks up a weekly game by canonical week and index.
func (s *Store) GetWeeklyGame(year, weekNumber, index int) (*GameRecord, error) {
	g, err := s.getGameByQuery(
		fmt.Sprintf(`SELECT %s
		 FROM games
		 WHERE run_kind = ?
		   AND week_year = ?
		   AND week_number = ?
		   AND week_index = ?`, gameSelectColumns),
		string(RunKindWeekly), year, weekNumber, index,
	)
	if err != nil {
		return nil, fmt.Errorf("querying weekly game: %w", err)
	}
	return g, nil
}

// ListWeeklyGames returns all records matching the canonical weekly prefix for
// the provided ISO week-year, including abandoned rows.
func (s *Store) ListWeeklyGames(year, weekNumber int) ([]GameRecord, error) {
	games, err := s.listGamesByQuery(
		fmt.Sprintf(`SELECT %s
		 FROM games
		 WHERE run_kind = ?
		   AND week_year = ?
		   AND week_number = ?
		 ORDER BY name DESC`, gameSelectColumns),
		string(RunKindWeekly), year, weekNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("listing weekly games: %w", err)
	}
	return games, nil
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
