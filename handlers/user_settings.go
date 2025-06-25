package handlers

import (
	"html/template"
	"net/http"
	"strings"
	
	"github.com/aungh/login-form/models"
	"github.com/aungh/login-form/utils"
)

// UserSettingsHandler handles the user settings page
func UserSettingsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := utils.GetSession(r)

	// Check if user is authenticated
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get current user ID
	userID, ok := session.Values["user_id"].(int)
	if !ok || userID <= 0 {
		http.Error(w, "User ID not found in session", http.StatusInternalServerError)
		return
	}

	// Get current user
	currentUser, err := models.GetUserByIDSafe(userID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"CurrentUser": currentUser,
	}

	// Handle form submissions
	if r.Method == "POST" {
		action := r.FormValue("action")
		currentPassword := r.FormValue("current_password")

		// Check if this is an OAuth user (has GoogleID or GithubID)
		isOAuthUser := currentUser.GoogleID != "" || currentUser.GithubID != ""

		// Verify current password for sensitive actions
		// Skip password verification for OAuth users when changing password
		if action != "toggle_2fa" && action != "toggle_face_auth" && 
		   !(isOAuthUser && action == "change_password") && 
		   !utils.CheckPasswordHash(currentPassword, currentUser.PasswordHash) {
			data["Error"] = "Current password is incorrect"
			renderUserSettingsTemplate(w, data)
			return
		}

		switch action {
		case "change_email":
			newEmail := strings.TrimSpace(r.FormValue("new_email"))
			if newEmail == "" {
				data["Error"] = "Email cannot be empty"
				renderUserSettingsTemplate(w, data)
				return
			}

			// Check if email is already in use by another user
			exists, _ := models.EmailExists(newEmail)
			if exists && newEmail != currentUser.Email {
				data["Error"] = "Email is already in use"
				renderUserSettingsTemplate(w, data)
				return
			}

			// Update email
			currentUser.Email = newEmail
			err = models.UpdateUser(currentUser)
			if err != nil {
				data["Error"] = "Failed to update email: " + err.Error()
				renderUserSettingsTemplate(w, data)
				return
			}

			// Update session
			session.Values["email"] = newEmail
			session.Save(r, w)

			data["Success"] = "Email updated successfully"

		case "change_password":
			newPassword := r.FormValue("new_password")
			confirmPassword := r.FormValue("confirm_password")

			if newPassword == "" {
				data["Error"] = "Password cannot be empty"
				renderUserSettingsTemplate(w, data)
				return
			}

			if newPassword != confirmPassword {
				data["Error"] = "Passwords do not match"
				renderUserSettingsTemplate(w, data)
				return
			}

			// Hash new password
			hashedPassword, err := utils.HashPassword(newPassword)
			if err != nil {
				data["Error"] = "Failed to hash password: " + err.Error()
				renderUserSettingsTemplate(w, data)
				return
			}

			// Update password
			currentUser.PasswordHash = hashedPassword
			err = models.UpdateUser(currentUser)
			if err != nil {
				data["Error"] = "Failed to update password: " + err.Error()
				renderUserSettingsTemplate(w, data)
				return
			}

			// Force logout after password change
			// Clear the session authentication
			session.Values["authenticated"] = false
			delete(session.Values, "user_id")
			delete(session.Values, "username")
			delete(session.Values, "email")
			delete(session.Values, "twofa_enabled")
			delete(session.Values, "twofa_verified")
			delete(session.Values, "face_auth_enabled")
			delete(session.Values, "face_auth_verified")
			session.Save(r, w)

			// Redirect to login page with success message
			http.Redirect(w, r, "/login?msg=password_changed", http.StatusSeeOther)
			return

		case "toggle_2fa":
			// Toggle 2FA status
			if currentUser.TwoFAEnabled {
				// Disable 2FA
				currentUser.TwoFAEnabled = false
				
				// We keep the secret in case the user wants to re-enable 2FA later
				// but we don't need to clear it from the database
				
				err = models.UpdateUser(currentUser)
				if err != nil {
					data["Error"] = "Failed to disable 2FA: " + err.Error()
					renderUserSettingsTemplate(w, data)
					return
				}
				
				// Update session
				session.Values["twofa_enabled"] = false
				session.Save(r, w)
				
				data["Success"] = "Two-factor authentication has been disabled"
			} else {
				// To enable 2FA, redirect to the setup page
				http.Redirect(w, r, "/setup-2fa", http.StatusSeeOther)
				return
			}
			
		case "toggle_face_auth":
			// Toggle face authentication status
			if currentUser.FaceAuthEnabled {
				// Disable face authentication
				currentUser.FaceAuthEnabled = false
				
				err = models.UpdateUser(currentUser)
				if err != nil {
					data["Error"] = "Failed to disable face authentication: " + err.Error()
					renderUserSettingsTemplate(w, data)
					return
				}
				
				// Update session
				session.Values["face_auth_enabled"] = false
				session.Save(r, w)
				
				// Also delete face data
				if err := utils.DeleteFaceData(currentUser.ID); err != nil {
					// Just log the error, don't stop the process
					// The face auth is already disabled in the database
					data["Warning"] = "Face authentication disabled, but there was an error deleting face data: " + err.Error()
				} else {
					data["Success"] = "Face authentication has been disabled"
				}
			} else {
				// To enable face authentication, redirect to the setup page
				http.Redirect(w, r, "/setup-face", http.StatusSeeOther)
				return
			}
		}
	}

	renderUserSettingsTemplate(w, data)
}

func renderUserSettingsTemplate(w http.ResponseWriter, data map[string]interface{}) {
	tmpl, err := template.ParseFiles("templates/user-settings.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}
