package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	// Create data directory if it doesn't exist
	dataDir := "./data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.Mkdir(dataDir, 0755); err != nil {
			return err
		}
	}

	// Open database connection
	dbPath := filepath.Join(dataDir, "users.db")
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Set connection parameters
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(time.Hour)

	// Create tables if they don't exist
	err = createTables()
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// createTables creates the necessary tables if they don't exist
func createTables() error {
	// Create users table
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		nickname TEXT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		google_id TEXT,
		github_id TEXT,
		profile_image TEXT,
		twofa_secret TEXT,
		twofa_enabled BOOLEAN DEFAULT 0,
		face_auth_enabled BOOLEAN DEFAULT 0,
		role TEXT DEFAULT 'user',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := DB.Exec(usersTable)
	if err != nil {
		return err
	}

	// Create email index
	emailIndex := `CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`
	_, err = DB.Exec(emailIndex)
	if err != nil {
		return err
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
