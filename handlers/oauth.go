package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aungh/login-form/models"
	"github.com/aungh/login-form/utils"
)

// GoogleAuthHandler initiates Google OAuth flow using Goth
func GoogleAuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Starting Google OAuth flow")

	// Use custom auth handler instead of gothic.BeginAuthHandler
	utils.CustomBeginAuthHandler(w, r, "google")
}

// GithubAuthHandler initiates GitHub OAuth flow using Goth
func GithubAuthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Starting GitHub OAuth flow")

	// Use custom auth handler instead of gothic.BeginAuthHandler
	utils.CustomBeginAuthHandler(w, r, "github")
}

// GoogleCallbackHandler handles the callback from Google OAuth
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Processing Google OAuth callback")

	// Add specific Google provider debugging
	log.Printf("Google callback URL: %s", r.URL.String())
	log.Printf("Google callback query params: %v", r.URL.Query())

	// Check for error in the callback
	if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
		log.Printf("Google OAuth error: %s - %s",
			errorMsg,
			r.URL.Query().Get("error_description"))
		http.Error(w, "Google authentication failed: "+errorMsg, http.StatusBadRequest)
		return
	}

	// Proceed with normal OAuth callback handling
	handleOAuthCallback(w, r, "google")
}

// GithubCallbackHandler handles the callback from GitHub OAuth
func GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Processing GitHub OAuth callback")

	// Add specific GitHub provider debugging
	log.Printf("GitHub callback URL: %s", r.URL.String())
	log.Printf("GitHub callback query params: %v", r.URL.Query())

	// Check for error in the callback
	if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
		log.Printf("GitHub OAuth error: %s - %s",
			errorMsg,
			r.URL.Query().Get("error_description"))
		http.Error(w, "GitHub authentication failed: "+errorMsg, http.StatusBadRequest)
		return
	}

	// Proceed with normal OAuth callback handling
	handleOAuthCallback(w, r, "github")
}

// handleOAuthCallback processes OAuth callbacks from any provider
func handleOAuthCallback(w http.ResponseWriter, r *http.Request, provider string) {
	log.Printf("Handling OAuth callback for provider: %s", provider)

	// Fix any session issues before processing the callback
	session, err := utils.FixSession(w, r)
	if err != nil {
		log.Printf("Error fixing session in handleOAuthCallback: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Ensure the provider is set in the session
	session.Values["oauth_provider"] = provider

	// Make sure we have a clean session state for the OAuth flow
	delete(session.Values, "_gothic_session")

	// Save the session with our provider information
	err = utils.SaveSession(session, w, r)
	if err != nil {
		log.Printf("Error saving session in handleOAuthCallback: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Log the session state before completing auth
	log.Printf("Session before CompleteUserAuth: %+v", session.Values)

	// Complete the auth process using our custom function
	log.Printf("About to complete %s auth with session: %+v", provider, session.Values)
	gothUser, err := utils.CustomCompleteUserAuth(w, r, provider)
	if err != nil {
		log.Printf("Error completing %s auth: %v", provider, err)
		http.Error(w, "Authentication failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully completed %s auth for user %s", provider, gothUser.Email)

	log.Printf("User authenticated: %s, Email: %s, Provider: %s", gothUser.Name, gothUser.Email, gothUser.Provider)

	// Check if user exists in database
	user, err := models.GetUserByEmailSafe(gothUser.Email)
	if err != nil {
		log.Printf("User not found with email %s, creating new user via %s OAuth", gothUser.Email, provider)

		// Create new user if not exists
		hashedPassword, err := utils.HashPassword(fmt.Sprintf("%s_%s", provider, gothUser.UserID))
		if err != nil {
			log.Printf("Failed to hash password: %s", err.Error())
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		// Create a new user
		newUser := &models.User{
			Email:           gothUser.Email,
			PasswordHash:    hashedPassword,
			TwoFAEnabled:    false,
			FaceAuthEnabled: false,
			Role:            "user",
			ProfileImage:    gothUser.AvatarURL,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Set provider-specific ID and handle username/nickname
		if provider == "google" {
			newUser.GoogleID = gothUser.UserID
			newUser.Username = gothUser.Name
		} else if provider == "github" {
			newUser.GithubID = gothUser.UserID

			// For GitHub, store the nickname and use it as username if name is empty
			log.Printf("GitHub user data - Name: %s, NickName: %s", gothUser.Name, gothUser.NickName)

			if gothUser.Name != "" {
				newUser.Username = gothUser.Name
				newUser.Nickname = gothUser.NickName
			} else {
				// If name is empty, use nickname as the username
				newUser.Username = gothUser.NickName
				newUser.Nickname = gothUser.NickName
			}
		}

		// Create the user in the database
		err = models.CreateUser(newUser)
		if err != nil {
			log.Printf("Failed to create user: %s", err.Error())
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		// Get the newly created user
		user, err = models.GetUserByEmail(gothUser.Email)
		if err != nil {
			log.Printf("Failed to get newly created user: %s", err.Error())
			http.Error(w, "Authentication failed", http.StatusInternalServerError)
			return
		}

		log.Printf("Successfully created new user %s (ID: %d) via %s OAuth", gothUser.Email, user.ID, provider)
	} else {
		log.Printf("User found with email %s (ID: %d)", gothUser.Email, user.ID)

		// Update provider ID if needed
		if provider == "google" && user.GoogleID == "" {
			user.GoogleID = gothUser.UserID
			err = models.UpdateUser(user)
			if err != nil {
				log.Printf("Failed to update Google ID: %s", err.Error())
			}
		} else if provider == "github" {
			updateNeeded := false

			// Update GitHub ID if needed
			if user.GithubID == "" {
				user.GithubID = gothUser.UserID
				updateNeeded = true
			}

			// Log GitHub user data
			log.Printf("GitHub user data for existing user - Name: %s, NickName: %s", gothUser.Name, gothUser.NickName)

			// For GitHub users, update nickname if needed
			if gothUser.NickName != "" && user.Nickname == "" {
				user.Nickname = gothUser.NickName
				updateNeeded = true
			}

			// If username is empty but we have a nickname, use nickname as username
			if user.Username == "" && gothUser.NickName != "" {
				user.Username = gothUser.NickName
				updateNeeded = true
			}

			// Update user if needed
			if updateNeeded {
				err = models.UpdateUser(user)
				if err != nil {
					log.Printf("Failed to update GitHub user data: %s", err.Error())
				}
			}
		}
	}

	// Already have the session from earlier, no need to get it again

	// Check if 2FA is required
	if user.TwoFAEnabled {
		log.Printf("2FA is enabled for user %s, redirecting to 2FA verification", user.Email)
		
		// Store pending auth info in session
		session.Values["pending_auth"] = true
		session.Values["pending_auth_user_id"] = user.ID
		session.Values["pending_auth_username"] = user.Username
		session.Values["pending_auth_nickname"] = user.Nickname
		session.Values["pending_auth_email"] = user.Email
		session.Values["pending_auth_twofa_enabled"] = true
		session.Values["pending_auth_role"] = user.Role
		
		// If user also has face auth enabled, set up the flag for face auth after 2FA
		if user.FaceAuthEnabled {
			log.Printf("User %s has both 2FA and face auth enabled, setting up face auth after 2FA", user.Email)
			session.Values["pending_face_after_2fa"] = true
			session.Values["pending_auth_face_enabled"] = true
		}

		err = utils.SaveSession(session, w, r)
		if err != nil {
			log.Printf("Error saving session before 2FA redirect: %v", err)
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// Redirect to 2FA verification
		http.Redirect(w, r, "/verify-2fa", http.StatusSeeOther)
		return
	}

	// Check if face auth is required (and 2FA is not enabled)
	if user.FaceAuthEnabled && !user.TwoFAEnabled {
		log.Printf("Only face auth is enabled for user %s, redirecting to face verification", user.Email)

		// Store pending auth info in session
		session.Values["pending_auth"] = true
		session.Values["pending_auth_user_id"] = user.ID
		session.Values["pending_auth_username"] = user.Username
		session.Values["pending_auth_nickname"] = user.Nickname
		session.Values["pending_auth_email"] = user.Email
		session.Values["pending_auth_face_enabled"] = true
		session.Values["pending_auth_role"] = user.Role

		err = utils.SaveSession(session, w, r)
		if err != nil {
			log.Printf("Error saving session before face verification redirect: %v", err)
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// Redirect to face verification
		http.Redirect(w, r, "/verify-face", http.StatusSeeOther)
		return
	}

	// If no additional auth required, create full session
	session.Values["authenticated"] = true
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Values["nickname"] = user.Nickname
	session.Values["email"] = user.Email
	session.Values["twofa_enabled"] = user.TwoFAEnabled
	session.Values["face_auth_enabled"] = user.FaceAuthEnabled
	session.Values["role"] = user.Role

	err = utils.SaveSession(session, w, r)
	if err != nil {
		log.Printf("Error saving session at end of OAuth flow: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s successfully authenticated via %s OAuth", user.Email, provider)

	// Redirect to home page
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}
