package app

import (
	"database/sql"
	"fmt"
)

// tuneSQLite applies PRAGMA settings for performance/safety trade-offs.
// Modes:
//
//	aggressive: maximum speed, risk of data loss on crash (synchronous=OFF, exclusive locking)
//	balanced: good speed with reasonable safety (synchronous=NORMAL)
//	safe: highest durability (synchronous=FULL)
func tuneSQLite(db *sql.DB, mode string) error {
	type pragmaSet struct {
		stmts            []string
		maxOpen, maxIdle int
	}
	makeBase := func(sync string, cacheMB int, extra []string) pragmaSet {
		// cache_size expects negative value = size in KB * -1
		cacheKB := cacheMB * 1024
		stmts := []string{
			"PRAGMA journal_mode=WAL",
			fmt.Sprintf("PRAGMA synchronous=%s", sync),
			"PRAGMA temp_store=MEMORY",
			fmt.Sprintf("PRAGMA cache_size=%d", -cacheKB),
			"PRAGMA wal_autocheckpoint=1000",
			"PRAGMA busy_timeout=5000",
		}
		stmts = append(stmts, extra...)
		return pragmaSet{stmts: stmts, maxOpen: 4, maxIdle: 4}
	}
	var cfg pragmaSet
	switch mode {
	case "aggressive":
		cfg = makeBase("OFF", 256, []string{"PRAGMA locking_mode=EXCLUSIVE", "PRAGMA mmap_size=30000000000", "PRAGMA journal_size_limit=67108864"})
		cfg.maxOpen, cfg.maxIdle = 1, 1 // exclusive mode works best single connection
	case "safe":
		cfg = makeBase("FULL", 16, []string{"PRAGMA mmap_size=100000000"})
	default: // balanced
		cfg = makeBase("NORMAL", 64, []string{"PRAGMA mmap_size=100000000", "PRAGMA journal_size_limit=67108864"})
	}
	for _, s := range cfg.stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("exec %s: %w", s, err)
		}
	}
	db.SetMaxOpenConns(cfg.maxOpen)
	db.SetMaxIdleConns(cfg.maxIdle)
	db.SetConnMaxLifetime(0)
	return nil
}