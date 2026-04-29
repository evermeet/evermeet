package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps a SQLite connection pool.
// All writes go through the write channel to serialize them — SQLite WAL mode
// supports multiple concurrent readers but only one writer at a time.
type DB struct {
	db      *sql.DB
	writeCh chan writeReq
}

type writeReq struct {
	fn     func(*sql.Tx) error
	result chan error
}

// Open opens (or creates) the SQLite database at path and runs pending migrations.
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// One writer connection to prevent SQLITE_BUSY under concurrent gossipsub writes.
	db.SetMaxOpenConns(1)

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}

	d := &DB{db: db, writeCh: make(chan writeReq, 256)}
	go d.writeLoop()
	return d, nil
}

// Close shuts down the write loop and closes the database.
func (d *DB) Close() error {
	close(d.writeCh)
	return d.db.Close()
}

// ReadDB returns the underlying *sql.DB for read-only queries.
// Writes must go through Write() to preserve serialization.
func (d *DB) ReadDB() *sql.DB {
	return d.db
}

// Write executes fn inside a transaction on the serialized write goroutine.
func (d *DB) Write(ctx context.Context, fn func(*sql.Tx) error) error {
	req := writeReq{fn: fn, result: make(chan error, 1)}
	select {
	case d.writeCh <- req:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-req.result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *DB) writeLoop() {
	for req := range d.writeCh {
		tx, err := d.db.Begin()
		if err != nil {
			req.result <- fmt.Errorf("begin tx: %w", err)
			continue
		}
		if err := req.fn(tx); err != nil {
			tx.Rollback()
			req.result <- err
			continue
		}
		req.result <- tx.Commit()
	}
}

func runMigrations(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS _migrations (version INTEGER PRIMARY KEY, name TEXT NOT NULL)`); err != nil {
		return err
	}

	files, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	// Sort by filename so migrations apply in order.
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}
		version := migrationVersion(f.Name())

		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM _migrations WHERE version = ?`, version).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue // already applied
		}

		data, err := migrationsFS.ReadFile("migrations/" + f.Name())
		if err != nil {
			return err
		}

		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("migration %s: %w", f.Name(), err)
		}
		if _, err := db.Exec(`INSERT INTO _migrations (version, name) VALUES (?, ?)`, version, f.Name()); err != nil {
			return err
		}
		log.Printf("store: applied migration %s", f.Name())
	}
	return nil
}

// migrationVersion extracts the leading integer from a filename like "0001_initial.sql".
func migrationVersion(name string) int {
	var v int
	fmt.Sscanf(name, "%d", &v)
	return v
}
