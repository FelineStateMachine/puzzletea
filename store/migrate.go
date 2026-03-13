package store

import (
	"database/sql"
	"fmt"
)

const currentSchemaVersion = 3

type migration struct {
	version int
	name    string
	apply   func(*sql.DB) error
}

var schemaMigrations = []migration{
	{version: 1, name: "create games table", apply: createGamesSchema},
	{version: 2, name: "add game metadata columns", apply: migrateGameMetadata},
	{version: 3, name: "refresh stats views", apply: refreshStatsViews},
}

type gameRowMeta struct {
	ID       int64
	Name     string
	GameType string
	Mode     string
	GameID   string
	ModeID   string
	RunKind  string
}

func runMigrations(db *sql.DB) error {
	if err := ensureSchemaMigrationsTable(db); err != nil {
		return err
	}

	version, err := schemaVersion(db)
	if err != nil {
		return err
	}
	if version == 0 {
		version, err = detectSchemaVersion(db)
		if err != nil {
			return err
		}
		if err := setSchemaVersion(db, version); err != nil {
			return err
		}
	}

	for _, m := range schemaMigrations {
		if m.version <= version {
			continue
		}
		if err := m.apply(db); err != nil {
			return fmt.Errorf("applying migration %d (%s): %w", m.version, m.name, err)
		}
		if err := setSchemaVersion(db, m.version); err != nil {
			return err
		}
	}

	return nil
}

func ensureSchemaMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
    id         INTEGER PRIMARY KEY CHECK(id = 1),
    version    INTEGER NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}
	return nil
}

func schemaVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow(`SELECT version FROM schema_migrations WHERE id = 1`).Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("reading schema version: %w", err)
	}
	return version, nil
}

func setSchemaVersion(db *sql.DB, version int) error {
	_, err := db.Exec(`
INSERT INTO schema_migrations (id, version)
VALUES (1, ?)
ON CONFLICT(id) DO UPDATE
SET version = excluded.version,
    updated_at = CURRENT_TIMESTAMP
`, version)
	if err != nil {
		return fmt.Errorf("writing schema version %d: %w", version, err)
	}
	return nil
}

func detectSchemaVersion(db *sql.DB) (int, error) {
	exists, err := tableExists(db, "games")
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	columns, err := tableColumns(db, "games")
	if err != nil {
		return 0, err
	}

	required := []string{
		"game_id",
		"mode_id",
		"run_kind",
		"run_date",
		"week_year",
		"week_number",
		"week_index",
		"seed_text",
	}
	for _, column := range required {
		if !columns[column] {
			return 1, nil
		}
	}

	return 2, nil
}

func createGamesSchema(db *sql.DB) error {
	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("creating games table: %w", err)
	}
	return nil
}

func migrateGameMetadata(db *sql.DB) error {
	if err := ensureGameColumns(db); err != nil {
		return err
	}
	if err := backfillGameMetadata(db); err != nil {
		return err
	}
	return nil
}

func refreshStatsViews(db *sql.DB) error {
	if _, err := db.Exec(`DROP VIEW IF EXISTS category_stats`); err != nil {
		return fmt.Errorf("dropping category_stats view: %w", err)
	}
	if _, err := db.Exec(`DROP VIEW IF EXISTS mode_stats`); err != nil {
		return fmt.Errorf("dropping mode_stats view: %w", err)
	}
	if _, err := db.Exec(createCategoryStatsViewSQL); err != nil {
		return fmt.Errorf("creating category_stats view: %w", err)
	}
	if _, err := db.Exec(createModeStatsViewSQL); err != nil {
		return fmt.Errorf("creating mode_stats view: %w", err)
	}
	return nil
}

func ensureGameColumns(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{name: "game_id", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "mode_id", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "run_kind", definition: "TEXT NOT NULL DEFAULT 'normal'"},
		{name: "run_date", definition: "DATE"},
		{name: "week_year", definition: "INTEGER"},
		{name: "week_number", definition: "INTEGER"},
		{name: "week_index", definition: "INTEGER"},
		{name: "seed_text", definition: "TEXT"},
	}

	existing, err := tableColumns(db, "games")
	if err != nil {
		return err
	}
	for _, column := range columns {
		if existing[column.name] {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE games ADD COLUMN %s %s", column.name, column.definition)); err != nil {
			return fmt.Errorf("adding games.%s: %w", column.name, err)
		}
	}
	return nil
}

func tableColumns(db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, fmt.Errorf("querying %s columns: %w", table, err)
	}
	defer rows.Close()

	columns := map[string]bool{}
	for rows.Next() {
		var (
			cid        int
			name       string
			typ        string
			notNull    int
			defaultV   any
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultV, &primaryKey); err != nil {
			return nil, fmt.Errorf("scanning %s column: %w", table, err)
		}
		columns[name] = true
	}
	return columns, rows.Err()
}

func tableExists(db *sql.DB, table string) (bool, error) {
	var name string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking %s existence: %w", table, err)
	}
	return true, nil
}

func backfillGameMetadata(db *sql.DB) error {
	rows, err := db.Query(`SELECT id, name, game_type, mode, game_id, mode_id, run_kind FROM games`)
	if err != nil {
		return fmt.Errorf("querying game metadata for backfill: %w", err)
	}
	defer rows.Close()

	var metas []gameRowMeta
	for rows.Next() {
		var meta gameRowMeta
		if err := rows.Scan(&meta.ID, &meta.Name, &meta.GameType, &meta.Mode, &meta.GameID, &meta.ModeID, &meta.RunKind); err != nil {
			return fmt.Errorf("scanning game metadata for backfill: %w", err)
		}
		metas = append(metas, meta)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading game metadata for backfill: %w", err)
	}

	for _, meta := range metas {
		gameID := meta.GameID
		if gameID == "" {
			gameID = CanonicalGameID(meta.GameType)
		}
		modeID := meta.ModeID
		if modeID == "" {
			modeID = CanonicalModeID(meta.Mode)
		}
		runKind := RunKind(meta.RunKind)
		if runKind == "" || runKind == RunKindNormal {
			runKind = RunKindForName(meta.Name)
		}
		runDate := RunDateForName(meta.Name)
		seedText := SeedTextForName(meta.Name)
		year, week, index, ok := WeeklyIdentityForName(meta.Name)
		if _, err := db.Exec(
			`UPDATE games
			 SET game_id = ?, mode_id = ?, run_kind = ?, run_date = ?,
			     week_year = ?, week_number = ?, week_index = ?, seed_text = ?
			 WHERE id = ?`,
			gameID, modeID, string(runKind), runDate,
			nullableIntIf(ok, year), nullableIntIf(ok, week), nullableIntIf(ok, index),
			nullableString(seedText), meta.ID,
		); err != nil {
			return fmt.Errorf("backfilling game metadata for %d: %w", meta.ID, err)
		}
	}
	return nil
}

func nullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullableIntIf(ok bool, value int) any {
	if !ok || value == 0 {
		return nil
	}
	return value
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
