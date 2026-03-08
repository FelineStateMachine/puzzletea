package store

import (
	"database/sql"
	"fmt"
)

type gameRowMeta struct {
	ID       int64
	Name     string
	GameType string
	Mode     string
	GameID   string
	ModeID   string
	RunKind  string
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
