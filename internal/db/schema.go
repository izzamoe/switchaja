package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Console represents a game console (e.g., PS1, PS2) with rental status.
//
// Fields:
//
//	ID: primary key
//	Name: human readable name (PS1, PS2, etc.)
//	Status: IDLE or RUNNING
//	EndTime: when the current rental ends (valid if RUNNING)
//	PricePerHour: pricing in local currency per hour
//
// The zero value of EndTime is treated as no active session.
type Console struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	EndTime      time.Time `json:"end_time"`
	PricePerHour int       `json:"price_per_hour"`
}

// Transaction records a rental usage window for a console.
// TotalPrice is precalculated for quick reporting.
type Transaction struct {
	ID          int64     `json:"id"`
	ConsoleID   int64     `json:"console_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	DurationMin int       `json:"duration_minutes"`
	TotalPrice  int       `json:"total_price"`
	// PricePerHourSnapshot is the hourly price used for this (current) transaction calculation.
	PricePerHourSnapshot int `json:"price_per_hour"`
}

// Init creates tables if they do not exist and seeds initial consoles.
// Init prepares the database schema and seeds a number of consoles.
// consoleCount determines how many PS entries (PS1..PSn) are created
// when the table is empty.
func Init(db *sql.DB, consoleCount int, pricePerHour int) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS consoles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		status TEXT NOT NULL DEFAULT 'IDLE',
		end_time DATETIME,
		price_per_hour INTEGER NOT NULL
	);`)
	if err != nil {
		return err
	}
	// users (auth). role: 'admin' or 'user'.
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`)
	if err != nil {
		return err
	}
	// generic settings key-value
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		console_id INTEGER NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		duration_minutes INTEGER NOT NULL,
		total_price INTEGER NOT NULL,
		price_per_hour_snapshot INTEGER,
		FOREIGN KEY(console_id) REFERENCES consoles(id)
	);`)
	if err != nil {
		return err
	}

	// seed consoles if empty
	var c int
	err = db.QueryRow(`SELECT COUNT(1) FROM consoles`).Scan(&c)
	if err != nil {
		return err
	}
	if c == 0 {
		for i := 1; i <= consoleCount; i++ {
			name := fmt.Sprintf("PS%d", i)
			if _, err := db.Exec(`INSERT INTO consoles(name,status,price_per_hour) VALUES(?, 'IDLE', ?)`, name, pricePerHour); err != nil {
				return err
			}
		}
	}
	// price change history table
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS price_changes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		console_id INTEGER NOT NULL,
		old_price INTEGER NOT NULL,
		new_price INTEGER NOT NULL,
		changed_at DATETIME NOT NULL,
		FOREIGN KEY(console_id) REFERENCES consoles(id)
	);`)

	// Migration: ensure column price_per_hour_snapshot exists (older DBs)
	// SQLite ADD COLUMN is safe idempotent if we check first.
	var colCount int
	// pragma table_info returns a row per column; count where name matches
	_ = db.QueryRow(`SELECT COUNT(1) FROM pragma_table_info('transactions') WHERE name='price_per_hour_snapshot'`).Scan(&colCount)
	if colCount == 0 {
		_, _ = db.Exec(`ALTER TABLE transactions ADD COLUMN price_per_hour_snapshot INTEGER`)
	}
	return nil
}

// ----- Users & Auth -----

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUser inserts a new user (hashed password already) returns id.
func CreateUser(dbx *sql.DB, username, passwordHash, role string) (int64, error) {
	if role != "admin" && role != "user" {
		return 0, errors.New("invalid role")
	}
	res, err := dbx.Exec(`INSERT INTO users(username,password_hash,role,created_at) VALUES(?,?,?,?)`, username, passwordHash, role, time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetUserByUsername fetches user and hash.
func GetUserByUsername(dbx *sql.DB, username string) (User, string, bool, error) {
	row := dbx.QueryRow(`SELECT id, username, password_hash, role, created_at FROM users WHERE username=?`, username)
	var u User
	var hash string
	if err := row.Scan(&u.ID, &u.Username, &hash, &u.Role, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, "", false, nil
		}
		return User{}, "", false, err
	}
	return u, hash, true, nil
}

// ListUsers returns all users (without password hash).
func ListUsers(dbx *sql.DB) ([]User, error) {
	rows, err := dbx.Query(`SELECT id, username, role, created_at FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

// DeleteUser removes user by id.
func DeleteUser(dbx *sql.DB, id int64) error {
	_, err := dbx.Exec(`DELETE FROM users WHERE id=?`, id)
	return err
}

// CountUsers returns count.
func CountUsers(dbx *sql.DB) (int, error) {
	var c int
	err := dbx.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&c)
	return c, err
}

// GetConsoles returns all consoles.
func GetConsoles(db *sql.DB) ([]Console, error) {
	rows, err := db.Query(`SELECT id,name,status,end_time,price_per_hour FROM consoles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Console
	for rows.Next() {
		var c Console
		var end sql.NullTime
		if err := rows.Scan(&c.ID, &c.Name, &c.Status, &end, &c.PricePerHour); err != nil {
			return nil, err
		}
		if end.Valid {
			c.EndTime = end.Time
		}
		res = append(res, c)
	}
	return res, rows.Err()
}

// StartRental sets a console to RUNNING and inserts a transaction skeleton.
func StartRental(db *sql.DB, consoleID int64, durationMin int) error {
	if durationMin <= 0 {
		return errors.New("duration must be > 0")
	}
	return withTx(db, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRow(`SELECT status FROM consoles WHERE id=?`, consoleID).Scan(&status); err != nil {
			return err
		}
		if status == "RUNNING" {
			return errors.New("console already running")
		}
		end := time.Now().Add(time.Duration(durationMin) * time.Minute)
		if _, err := tx.Exec(`UPDATE consoles SET status='RUNNING', end_time=? WHERE id=?`, end, consoleID); err != nil {
			return err
		}
		// Insert transaction
		pricePerHour := 0
		if err := tx.QueryRow(`SELECT price_per_hour FROM consoles WHERE id=?`, consoleID).Scan(&pricePerHour); err != nil {
			return err
		}
		price := calcPrice(pricePerHour, durationMin)
		if _, err := tx.Exec(`INSERT INTO transactions(console_id,start_time,end_time,duration_minutes,total_price,price_per_hour_snapshot) VALUES(?,?,?,?,?,?)`, consoleID, time.Now(), end, durationMin, price, pricePerHour); err != nil {
			return err
		}
		return nil
	})
}

// ExtendRental extends the end_time and updates the latest transaction.
func ExtendRental(db *sql.DB, consoleID int64, addMinutes int) error {
	if addMinutes <= 0 {
		return errors.New("addMinutes must be > 0")
	}
	return withTx(db, func(tx *sql.Tx) error {
		var status string
		var end time.Time
		if err := tx.QueryRow(`SELECT status,end_time FROM consoles WHERE id=?`, consoleID).Scan(&status, &end); err != nil {
			return err
		}
		if status != "RUNNING" {
			return errors.New("console not running")
		}
		newEnd := end.Add(time.Duration(addMinutes) * time.Minute)
		if _, err := tx.Exec(`UPDATE consoles SET end_time=? WHERE id=?`, newEnd, consoleID); err != nil {
			return err
		}
		// Update last transaction for this console
		row := tx.QueryRow(`SELECT id, duration_minutes, total_price FROM transactions WHERE console_id=? ORDER BY id DESC LIMIT 1`, consoleID)
		var tid int64
		var duration int
		var total int
		if err := row.Scan(&tid, &duration, &total); err != nil {
			return err
		}
		pricePerHour := 0
		if err := tx.QueryRow(`SELECT price_per_hour FROM consoles WHERE id=?`, consoleID).Scan(&pricePerHour); err != nil {
			return err
		}
		newDuration := duration + addMinutes
		newPrice := calcPrice(pricePerHour, newDuration)
		if _, err := tx.Exec(`UPDATE transactions SET end_time=?, duration_minutes=?, total_price=?, price_per_hour_snapshot=? WHERE id=?`, newEnd, newDuration, newPrice, pricePerHour, tid); err != nil {
			return err
		}
		return nil
	})
}

// StopRental stops an active rental.
func StopRental(db *sql.DB, consoleID int64) error {
	return withTx(db, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRow(`SELECT status FROM consoles WHERE id=?`, consoleID).Scan(&status); err != nil {
			return err
		}
		if status != "RUNNING" {
			return errors.New("console not running")
		}
		if _, err := tx.Exec(`UPDATE consoles SET status='IDLE', end_time=NULL WHERE id=?`, consoleID); err != nil {
			return err
		}
		return nil
	})
}

// LastTransaction returns the most recent transaction for a console.
func LastTransaction(db *sql.DB, consoleID int64) (Transaction, bool, error) {
	row := db.QueryRow(`SELECT id, console_id, start_time, end_time, duration_minutes, total_price, COALESCE(price_per_hour_snapshot,0) FROM transactions WHERE console_id=? ORDER BY id DESC LIMIT 1`, consoleID)
	var t Transaction
	err := row.Scan(&t.ID, &t.ConsoleID, &t.StartTime, &t.EndTime, &t.DurationMin, &t.TotalPrice, &t.PricePerHourSnapshot)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Transaction{}, false, nil
		}
		return Transaction{}, false, err
	}
	return t, true, nil
}

// DueSoon returns consoles whose rentals will end within threshold.
func DueSoon(db *sql.DB, threshold time.Duration) ([]Console, error) {
	rows, err := db.Query(`SELECT id,name,status,end_time,price_per_hour FROM consoles WHERE status='RUNNING' AND end_time <= ? ORDER BY end_time`, time.Now().Add(threshold))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []Console
	for rows.Next() {
		var c Console
		var end sql.NullTime
		if err := rows.Scan(&c.ID, &c.Name, &c.Status, &end, &c.PricePerHour); err != nil {
			return nil, err
		}
		if end.Valid {
			c.EndTime = end.Time
		}
		res = append(res, c)
	}
	return res, rows.Err()
}

func calcPrice(pricePerHour int, durationMin int) int {
	// price is proportional to minutes (ceil to next minute already done by integer minutes)
	return int(float64(pricePerHour) * (float64(durationMin) / 60.0))
}

func withTx(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	err = fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// UpdatePrice sets a new price_per_hour for a console.
func UpdatePrice(dbx *sql.DB, consoleID int64, newPrice int) error {
	if newPrice <= 0 {
		return errors.New("price must be > 0")
	}
	var oldPrice int
	if err := dbx.QueryRow(`SELECT price_per_hour FROM consoles WHERE id=?`, consoleID).Scan(&oldPrice); err != nil {
		return err
	}
	if _, err := dbx.Exec(`UPDATE consoles SET price_per_hour=? WHERE id=?`, newPrice, consoleID); err != nil {
		return err
	}
	// record history
	_, _ = dbx.Exec(`INSERT INTO price_changes(console_id, old_price, new_price, changed_at) VALUES(?,?,?,?)`, consoleID, oldPrice, newPrice, time.Now())
	return nil
}

// GetLastPriceChange returns the latest price change for a console.
func GetLastPriceChange(dbx *sql.DB, consoleID int64) (oldPrice, newPrice int, changedAt time.Time, ok bool, err error) {
	row := dbx.QueryRow(`SELECT old_price, new_price, changed_at FROM price_changes WHERE console_id=? ORDER BY id DESC LIMIT 1`, consoleID)
	if err = row.Scan(&oldPrice, &newPrice, &changedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, time.Time{}, false, nil
		}
		return 0, 0, time.Time{}, false, err
	}
	return oldPrice, newPrice, changedAt, true, nil
}

// ----- Settings & MQTT Config -----

// SetSetting upserts a key/value pair.
func SetSetting(dbx *sql.DB, key, value string) error {
	_, err := dbx.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

// GetSetting fetches a setting; bool false if not present.
func GetSetting(dbx *sql.DB, key string) (string, bool, error) {
	var v string
	err := dbx.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&v)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return v, true, nil
}

// MQTTConfig persisted config.
type MQTTConfig struct {
	Broker   string `json:"broker"`
	Prefix   string `json:"prefix"`
	Username string `json:"username"`
	Password string `json:"password"`
}

const mqttConfigKey = "mqtt_config"

func SaveMQTTConfig(dbx *sql.DB, cfg MQTTConfig) error {
	if cfg.Broker == "" { // delete config
		_, err := dbx.Exec(`DELETE FROM settings WHERE key=?`, mqttConfigKey)
		return err
	}
	b, _ := json.Marshal(cfg)
	return SetSetting(dbx, mqttConfigKey, string(b))
}

func LoadMQTTConfig(dbx *sql.DB) (MQTTConfig, bool, error) {
	v, ok, err := GetSetting(dbx, mqttConfigKey)
	if err != nil || !ok {
		return MQTTConfig{}, false, err
	}
	var cfg MQTTConfig
	if err := json.Unmarshal([]byte(v), &cfg); err != nil {
		return MQTTConfig{}, false, err
	}
	return cfg, true, nil
}
