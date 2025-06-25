package database

import (
	"log"
)

// MigrateDB handles database schema migrations
func MigrateDB() error {
	log.Println("Running database migrations...")

	// Check if nickname column exists in users table
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='nickname'`).Scan(&count)
	if err != nil {
		return err
	}

	// If nickname column doesn't exist, add it
	if count == 0 {
		log.Println("Adding nickname column to users table...")
		_, err := DB.Exec(`ALTER TABLE users ADD COLUMN nickname TEXT;`)
		if err != nil {
			log.Printf("Error adding nickname column: %v", err)
			return err
		}
		log.Println("Successfully added nickname column to users table")
	} else {
		log.Println("Nickname column already exists in users table")
	}

	// Check if username column is NOT NULL
	var notNull int
	err = DB.QueryRow(`SELECT "notnull" FROM pragma_table_info('users') WHERE name='username'`).Scan(&notNull)
	if err != nil {
		return err
	}

	// If username is NOT NULL, we need to modify it to allow NULL values
	if notNull == 1 {
		log.Println("Modifying username column to allow NULL values...")
		
		// SQLite doesn't support ALTER COLUMN, so we need to recreate the table
		// First, create a backup of the current table
		_, err := DB.Exec(`
			CREATE TABLE users_backup AS SELECT * FROM users;
		`)
		if err != nil {
			log.Printf("Error creating backup table: %v", err)
			return err
		}

		// Drop the original table
		_, err = DB.Exec(`DROP TABLE users;`)
		if err != nil {
			log.Printf("Error dropping original table: %v", err)
			return err
		}

		// Recreate the table with the updated schema
		_, err = DB.Exec(`
		CREATE TABLE users (
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
		`)
		if err != nil {
			log.Printf("Error recreating table: %v", err)
			return err
		}

		// Copy data from backup table
		_, err = DB.Exec(`
			INSERT INTO users 
			SELECT 
				id, username, NULL as nickname, email, password_hash, 
				google_id, github_id, profile_image, twofa_secret, 
				twofa_enabled, face_auth_enabled, role, created_at, updated_at 
			FROM users_backup;
		`)
		if err != nil {
			log.Printf("Error copying data from backup: %v", err)
			return err
		}

		// Drop backup table
		_, err = DB.Exec(`DROP TABLE users_backup;`)
		if err != nil {
			log.Printf("Error dropping backup table: %v", err)
			return err
		}

		// Recreate email index
		_, err = DB.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`)
		if err != nil {
			log.Printf("Error recreating email index: %v", err)
			return err
		}

		log.Println("Successfully modified username column to allow NULL values")
	}

	log.Println("Database migrations completed successfully")
	return nil
}
