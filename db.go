package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// db is the package-level database connection pool.
var db *sql.DB

// initDB opens a PostgreSQL connection using DATABASE_URL and runs migrations.
func initDB() {
	var err error
	db, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// FIX M-4: set connection pool limits to prevent exhausting DB connections.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	migrateDB()
	log.Println("database connection established and schema is up to date")
}

// migrateDB creates the registrations table if it does not already exist.
func migrateDB() {
	const query = `
CREATE TABLE IF NOT EXISTS registrations (
    email         TEXT PRIMARY KEY,
    first_name    TEXT NOT NULL,
    last_name     TEXT NOT NULL,
    area          TEXT NOT NULL,
    grp           TEXT NOT NULL,
    function      TEXT NOT NULL,
    gender        TEXT NOT NULL,
    local         TEXT NOT NULL,
    district      TEXT NOT NULL,
    status        TEXT NOT NULL,
    preferred_day TEXT NOT NULL
);`
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("failed to run database migration: %v", err)
	}
}

// record upserts an Entry into the registrations table.
// Returns true on success, false on failure.
func record(details Entry) bool {
	const query = `
INSERT INTO registrations
    (email, first_name, last_name, area, grp, function, gender, local, district, status, preferred_day)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (email) DO UPDATE SET
    first_name    = EXCLUDED.first_name,
    last_name     = EXCLUDED.last_name,
    area          = EXCLUDED.area,
    grp           = EXCLUDED.grp,
    function      = EXCLUDED.function,
    gender        = EXCLUDED.gender,
    local         = EXCLUDED.local,
    district      = EXCLUDED.district,
    status        = EXCLUDED.status,
    preferred_day = EXCLUDED.preferred_day;`

	_, err := db.Exec(query,
		details.Email,
		details.FirstName,
		details.LastName,
		details.Area,
		details.Group,
		details.Function,
		details.Gender,
		details.Local,
		details.District,
		details.Status,
		details.PreferredDay,
	)
	if err != nil {
		log.Printf("failed to upsert registration for %q: %v", details.Email, err)
		return false
	}
	return true
}

// checkSize returns the number of registrations for a given preferred day.
func checkSize(day string) int {
	const query = `SELECT COUNT(*) FROM registrations WHERE preferred_day = $1`
	var count int
	if err := db.QueryRow(query, day).Scan(&count); err != nil {
		log.Printf("failed to count registrations for day %q: %v", day, err)
		return 0
	}
	return count
}

// retrieveRecords returns all registrations ordered by last name then first name.
func retrieveRecords() []Entry {
	const query = `
SELECT email, first_name, last_name, area, grp, function, gender, local, district, status, preferred_day
FROM registrations
ORDER BY last_name, first_name`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("failed to retrieve registrations: %v", err)
		return nil
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(
			&e.Email,
			&e.FirstName,
			&e.LastName,
			&e.Area,
			&e.Group,
			&e.Function,
			&e.Gender,
			&e.Local,
			&e.District,
			&e.Status,
			&e.PreferredDay,
		); err != nil {
			log.Printf("failed to scan registration row: %v", err)
			continue
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		log.Printf("error iterating registration rows: %v", err)
	}
	return entries
}
