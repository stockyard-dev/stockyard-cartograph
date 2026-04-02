package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Sitemap struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Domain       string   `json:"domain"`
	PageCount    int      `json:"page_count"`
	Status       string   `json:"status"`
	CreatedAt    string   `json:"created_at"`
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dsn := filepath.Join(dataDir, "cartograph.db") + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sitemaps (
			id TEXT PRIMARY KEY,\n\t\t\tname TEXT DEFAULT '',\n\t\t\tdomain TEXT DEFAULT '',\n\t\t\tpage_count INTEGER DEFAULT 0,\n\t\t\tstatus TEXT DEFAULT 'pending',
			created_at TEXT DEFAULT (datetime('now'))
		)`)
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }

func (d *DB) Create(e *Sitemap) error {
	e.ID = genID()
	e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := d.db.Exec(`INSERT INTO sitemaps (id, name, domain, page_count, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		e.ID, e.Name, e.Domain, e.PageCount, e.Status, e.CreatedAt)
	return err
}

func (d *DB) Get(id string) *Sitemap {
	row := d.db.QueryRow(`SELECT id, name, domain, page_count, status, created_at FROM sitemaps WHERE id=?`, id)
	var e Sitemap
	if err := row.Scan(&e.ID, &e.Name, &e.Domain, &e.PageCount, &e.Status, &e.CreatedAt); err != nil {
		return nil
	}
	return &e
}

func (d *DB) List() []Sitemap {
	rows, err := d.db.Query(`SELECT id, name, domain, page_count, status, created_at FROM sitemaps ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []Sitemap
	for rows.Next() {
		var e Sitemap
		if err := rows.Scan(&e.ID, &e.Name, &e.Domain, &e.PageCount, &e.Status, &e.CreatedAt); err != nil {
			continue
		}
		result = append(result, e)
	}
	return result
}

func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM sitemaps WHERE id=?`, id)
	return err
}

func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM sitemaps`).Scan(&n)
	return n
}
