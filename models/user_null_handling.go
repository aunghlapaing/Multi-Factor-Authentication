package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/aungh/login-form/database"
)

// GetUserByIDSafe retrieves a user by ID with NULL handling
func GetUserByIDSafe(id int) (*User, error) {
	query := `
	SELECT id, username, nickname, email, password_hash, google_id, github_id, 
	profile_image, twofa_secret, twofa_enabled, face_auth_enabled, 
	role, created_at, updated_at 
	FROM users WHERE id = ?
	`

	row := database.DB.QueryRow(query, id)

	user := &User{}
	var createdAt, updatedAt string
	var username, nickname, googleID, githubID, profileImage, twoFASecret, role sql.NullString

	err := row.Scan(
		&user.ID,
		&username,
		&nickname,
		&user.Email,
		&user.PasswordHash,
		&googleID,
		&githubID,
		&profileImage,
		&twoFASecret,
		&user.TwoFAEnabled,
		&user.FaceAuthEnabled,
		&role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Convert NullString to string, empty string if NULL
	if username.Valid {
		user.Username = username.String
	}
	if nickname.Valid {
		user.Nickname = nickname.String
	}
	if googleID.Valid {
		user.GoogleID = googleID.String
	}
	if githubID.Valid {
		user.GithubID = githubID.String
	}
	if profileImage.Valid {
		user.ProfileImage = profileImage.String
	}
	if twoFASecret.Valid {
		user.TwoFASecret = twoFASecret.String
	}
	if role.Valid {
		user.Role = role.String
	} else {
		user.Role = "user" // Default role
	}

	// Parse timestamps
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return user, nil
}

// GetUserByEmailSafe retrieves a user by email with NULL handling
func GetUserByEmailSafe(email string) (*User, error) {
	query := `
	SELECT id, username, nickname, email, password_hash, google_id, github_id, 
	profile_image, twofa_secret, twofa_enabled, face_auth_enabled, 
	role, created_at, updated_at 
	FROM users WHERE email = ?
	`

	row := database.DB.QueryRow(query, email)

	user := &User{}
	var createdAt, updatedAt string
	var username, nickname, googleID, githubID, profileImage, twoFASecret, role sql.NullString

	err := row.Scan(
		&user.ID,
		&username,
		&nickname,
		&user.Email,
		&user.PasswordHash,
		&googleID,
		&githubID,
		&profileImage,
		&twoFASecret,
		&user.TwoFAEnabled,
		&user.FaceAuthEnabled,
		&role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Convert NullString to string, empty string if NULL
	if username.Valid {
		user.Username = username.String
	}
	if nickname.Valid {
		user.Nickname = nickname.String
	}
	if googleID.Valid {
		user.GoogleID = googleID.String
	}
	if githubID.Valid {
		user.GithubID = githubID.String
	}
	if profileImage.Valid {
		user.ProfileImage = profileImage.String
	}
	if twoFASecret.Valid {
		user.TwoFASecret = twoFASecret.String
	}
	if role.Valid {
		user.Role = role.String
	} else {
		user.Role = "user" // Default role
	}

	// Parse timestamps
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return user, nil
}

// GetAllUsersSafe returns all users in the system with NULL handling
func GetAllUsersSafe() ([]*User, error) {
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
		var username, nickname, googleID, githubID, profileImage, twoFASecret, role sql.NullString

		err := rows.Scan(
			&user.ID,
			&username,
			&nickname,
			&user.Email,
			&user.PasswordHash,
			&googleID,
			&githubID,
			&profileImage,
			&twoFASecret,
			&user.TwoFAEnabled,
			&user.FaceAuthEnabled,
			&role,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Convert NullString to string, empty string if NULL
		if username.Valid {
			user.Username = username.String
		}
		if nickname.Valid {
			user.Nickname = nickname.String
		}
		if googleID.Valid {
			user.GoogleID = googleID.String
		}
		if githubID.Valid {
			user.GithubID = githubID.String
		}
		if profileImage.Valid {
			user.ProfileImage = profileImage.String
		}
		if twoFASecret.Valid {
			user.TwoFASecret = twoFASecret.String
		}
		if role.Valid {
			user.Role = role.String
		} else {
			user.Role = "user" // Default role
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
