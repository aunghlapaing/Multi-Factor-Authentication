package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aungh/login-form/database"
)

// User represents a user in the system
type User struct {
	ID                int
	Username          string
	Nickname          string
	Email             string
	PasswordHash      string
	GoogleID          string
	GithubID          string
	ProfileImage      string
	TwoFASecret       string
	TwoFAEnabled      bool
	FaceAuthEnabled   bool
	Role              string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CreateUser creates a new user
func CreateUser(user *User) error {
	fmt.Printf("CreateUser called with: %+v\n", user)

	if user.Email == "" {
		fmt.Println("CreateUser error: email is required")
		return errors.New("email is required")
	}

	// Check if email already exists
	exists, err := EmailExists(user.Email)
	if err != nil {
		fmt.Printf("Error checking if email exists: %v\n", err)
	}
	if exists {
		fmt.Printf("CreateUser error: email %s already exists\n", user.Email)
		return errors.New("email already exists")
	}

	// Set timestamps
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
		fmt.Println("Set CreatedAt timestamp")
	}
	user.UpdatedAt = time.Now()
	fmt.Println("Set UpdatedAt timestamp")

	// Set default role if not specified
	if user.Role == "" {
		user.Role = "user"
		fmt.Println("Set default role to 'user'")
	}

	// Store user in database
	query := `
	INSERT INTO users (
		username, nickname, email, password_hash, google_id, github_id, 
		profile_image, twofa_secret, twofa_enabled, face_auth_enabled, 
		role, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	fmt.Println("Executing SQL query to insert user:")
	fmt.Println(query)

	result, err := database.DB.Exec(
		query,
		user.Username,
		user.Nickname,
		user.Email,
		user.PasswordHash,
		user.GoogleID,
		user.GithubID,
		user.ProfileImage,
		user.TwoFASecret,
		user.TwoFAEnabled,
		user.FaceAuthEnabled,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		fmt.Printf("Error executing SQL query: %v\n", err)
		return err
	}
	fmt.Println("SQL query executed successfully")

	// Get the last inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Update the user ID
	user.ID = int(id)

	return nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(id int) (*User, error) {
	query := `
	SELECT id, username, nickname, email, password_hash, google_id, github_id, 
	profile_image, twofa_secret, twofa_enabled, face_auth_enabled, 
	role, created_at, updated_at 
	FROM users WHERE id = ?
	`

	row := database.DB.QueryRow(query, id)

	user := &User{}
	var createdAt, updatedAt string

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Nickname,
		&user.Email,
		&user.PasswordHash,
		&user.GoogleID,
		&user.GithubID,
		&user.ProfileImage,
		&user.TwoFASecret,
		&user.TwoFAEnabled,
		&user.FaceAuthEnabled,
		&user.Role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Parse timestamps
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*User, error) {
	query := `
	SELECT id, username, nickname, email, password_hash, google_id, github_id, 
	profile_image, twofa_secret, twofa_enabled, face_auth_enabled, 
	role, created_at, updated_at 
	FROM users WHERE email = ?
	`

	row := database.DB.QueryRow(query, email)

	user := &User{}
	var createdAt, updatedAt string

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Nickname,
		&user.Email,
		&user.PasswordHash,
		&user.GoogleID,
		&user.GithubID,
		&user.ProfileImage,
		&user.TwoFASecret,
		&user.TwoFAEnabled,
		&user.FaceAuthEnabled,
		&user.Role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Parse timestamps
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return user, nil
}

// GetAllUsers returns all users in the system
func GetAllUsers() ([]*User, error) {
	query := `
	SELECT id, username, nickname, email, password_hash, google_id, github_id, 
	profile_image, twofa_secret, twofa_enabled, face_auth_enabled, 
	role, created_at, updated_at 
	FROM users ORDER BY id
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userList := make([]*User, 0)
	for rows.Next() {
		user := &User{}
		var createdAt, updatedAt string

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Nickname,
			&user.Email,
			&user.PasswordHash,
			&user.GoogleID,
			&user.GithubID,
			&user.ProfileImage,
			&user.TwoFASecret,
			&user.TwoFAEnabled,
			&user.FaceAuthEnabled,
			&user.Role,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Parse timestamps
		user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		userList = append(userList, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userList, nil
}

// UpdateUser updates an existing user
func UpdateUser(user *User) error {
	if user.ID == 0 {
		return errors.New("user ID is required")
	}

	// Check if user exists
	existingUser, err := GetUserByIDSafe(user.ID)
	if err != nil {
		return err
	}

	// Check if email is already taken by another user
	if existingUser.Email != user.Email {
		exists, _ := EmailExists(user.Email)
		if exists {
			return errors.New("email already exists")
		}
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Store updated user in database
	query := `
	UPDATE users SET 
		username = ?, 
		nickname = ?, 
		email = ?, 
		password_hash = ?, 
		google_id = ?, 
		github_id = ?, 
		profile_image = ?, 
		twofa_secret = ?, 
		twofa_enabled = ?, 
		face_auth_enabled = ?, 
		role = ?,
		updated_at = ? 
	WHERE id = ?
	`

	_, err = database.DB.Exec(
		query,
		user.Username,
		user.Nickname,
		user.Email,
		user.PasswordHash,
		user.GoogleID,
		user.GithubID,
		user.ProfileImage,
		user.TwoFASecret,
		user.TwoFAEnabled,
		user.FaceAuthEnabled,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)

	return err
}

// DeleteUser deletes a user
func DeleteUser(id int) error {
	// Check if user exists and get user data
	user, err := GetUserByIDSafe(id)
	if err != nil {
		return err
	}
	
	// Delete user from database
	query := "DELETE FROM users WHERE id = ?"
	_, err = database.DB.Exec(query, id)
	if err != nil {
		return err
	}
	
	// Return a special flag if face auth was enabled
	// This allows the caller to handle face data deletion
	if user.FaceAuthEnabled {
		return errors.New("face_auth_data_needs_deletion")
	}
	
	return nil
}

// EmailExists checks if an email already exists
func EmailExists(email string) (bool, error) {
	query := "SELECT COUNT(*) FROM users WHERE email = ?"
	var count int
	err := database.DB.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
