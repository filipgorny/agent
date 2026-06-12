package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SQLite is a lexical memory backed by SQLite FTS5 (full-text + BM25 ranking)
// with metadata stored as JSON (queryable via SQLite's native JSON1). Uses the
// pure-Go modernc.org/sqlite driver (no cgo; FTS5 and JSON1 are built in).
type SQLite struct {
	db *sql.DB
}

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS memory (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    content    TEXT NOT NULL,
    meta       TEXT,
    source     TEXT,
    created_at INTEGER NOT NULL
);
CREATE VIRTUAL TABLE IF NOT EXISTS memory_fts USING fts5(
    content,
    content='memory',
    content_rowid='id'
);
CREATE TRIGGER IF NOT EXISTS memory_ai AFTER INSERT ON memory BEGIN
    INSERT INTO memory_fts(rowid, content) VALUES (new.id, new.content);
END;
CREATE TRIGGER IF NOT EXISTS memory_ad AFTER DELETE ON memory BEGIN
    INSERT INTO memory_fts(memory_fts, rowid, content) VALUES('delete', old.id, old.content);
END;
CREATE TRIGGER IF NOT EXISTS memory_au AFTER UPDATE ON memory BEGIN
    INSERT INTO memory_fts(memory_fts, rowid, content) VALUES('delete', old.id, old.content);
    INSERT INTO memory_fts(rowid, content) VALUES (new.id, new.content);
END;
`

// NewSQLite opens (or creates) a SQLite-backed memory at path. An empty path
// uses an in-memory database (no file).
func NewSQLite(path string) (*SQLite, error) {
	if path == "" {
		path = ":memory:"
	}

	db, err := sql.Open("sqlite", path)

	if err != nil {
		return nil, fmt.Errorf("memory: open sqlite: %w", err)
	}

	// Keep a single connection so an in-memory database persists across queries.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(sqliteSchema); err != nil {
		db.Close()

		return nil, fmt.Errorf("memory: init sqlite schema: %w", err)
	}

	return &SQLite{db: db}, nil
}

// Close releases the underlying database.
func (m *SQLite) Close() error {
	return m.db.Close()
}

func (m *SQLite) Remember(ctx context.Context, content string, meta map[string]any) error {
	var metaJSON string

	if meta != nil {
		b, err := json.Marshal(meta)

		if err != nil {
			return fmt.Errorf("memory: marshal meta: %w", err)
		}

		metaJSON = string(b)
	}

	source, _ := meta["source"].(string)

	_, err := m.db.ExecContext(ctx,
		`INSERT INTO memory(content, meta, source, created_at) VALUES(?, ?, ?, ?)`,
		content, metaJSON, source, time.Now().Unix())

	if err != nil {
		return fmt.Errorf("memory: remember: %w", err)
	}

	return nil
}

func (m *SQLite) Read(ctx context.Context, query string, topK int) ([]Record, error) {
	match := ftsMatchQuery(query)

	if match == "" {
		return nil, nil
	}

	if topK <= 0 {
		topK = DefaultTopK
	}

	rows, err := m.db.QueryContext(ctx,
		`SELECT m.content, m.meta, bm25(memory_fts) AS score
		 FROM memory_fts JOIN memory m ON m.id = memory_fts.rowid
		 WHERE memory_fts MATCH ?
		 ORDER BY score
		 LIMIT ?`,
		match, topK)

	if err != nil {
		return nil, fmt.Errorf("memory: read: %w", err)
	}

	defer rows.Close()

	var out []Record

	for rows.Next() {
		var (
			content  string
			metaJSON sql.NullString
			score    float64
		)

		if err := rows.Scan(&content, &metaJSON, &score); err != nil {
			return nil, fmt.Errorf("memory: scan row: %w", err)
		}

		rec := Record{Content: content, Score: score}

		if metaJSON.Valid && metaJSON.String != "" {
			_ = json.Unmarshal([]byte(metaJSON.String), &rec.Meta)
		}

		out = append(out, rec)
	}

	return out, rows.Err()
}

// ftsMatchQuery turns a free-form query into a safe FTS5 MATCH expression by
// quoting each token as a phrase (implicit AND), avoiding operator parse errors.
func ftsMatchQuery(query string) string {
	fields := strings.Fields(query)

	if len(fields) == 0 {
		return ""
	}

	quoted := make([]string, 0, len(fields))

	for _, f := range fields {
		f = strings.ReplaceAll(f, `"`, "")

		if f != "" {
			quoted = append(quoted, `"`+f+`"`)
		}
	}

	return strings.Join(quoted, " ")
}
