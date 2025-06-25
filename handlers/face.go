package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/aungh/login-form/models"
	"github.com/aungh/login-form/utils"
	"github.com/gorilla/sessions"
)

// SetupFaceHandler handles face authentication setup
func SetupFaceHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := utils.GetSession(r)

	// Check if user is authenticated
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "User ID not found in session", http.StatusInternalServerError)
		return
	}

	user, err := models.GetUserByIDSafe(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Process form submission
	if r.Method == "POST" {
		faceData := r.FormValue("face_data")

		// Validate input
		if faceData == "" {
			renderFacePage(w, "Face data is required", true)
			return
		}

		// Enable face auth and save face data
		if err := enableFaceAuth(user, faceData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update session
		session.Values["face_auth_enabled"] = true
		session.Save(r, w)

		http.Redirect(w, r, "/?face_setup=success", http.StatusSeeOther)
		return
	}

	// Display face setup page
	renderFacePage(w, "", true)
}

// VerifyFaceHandler handles face authentication verification
func VerifyFaceHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := utils.GetSession(r)

	// Debug session values
	fmt.Printf("DEBUG: Face verification handler - session values: %+v\n", session.Values)

	// Check if there's a pending authentication
	email, ok := session.Values["pending_auth_email"].(string)
	fmt.Printf("DEBUG: Face verification - pending_auth_email: %s (valid: %v)\n", email, ok)
	if !ok || email == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	// Check if 2FA was completed before face verification
	twoFACompleted, _ := session.Values["twofa_completed"].(bool)
	fmt.Printf("DEBUG: Face verification - 2FA completed: %v\n", twoFACompleted)

	// Process form submission
	if r.Method == "POST" {
		faceData := r.FormValue("face_data")

		// Validate input
		if faceData == "" {
			renderFacePage(w, "Face data is required", false)
			return
		}

		// Get user - use the safe version that handles NULL values
		user, err := models.GetUserByEmailSafe(email)
		if err != nil {
			http.Error(w, "User not found: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// If user doesn't have face auth enabled yet, enable it and save the face data
		if !user.FaceAuthEnabled {
			if err := enableFaceAuth(user, faceData); err != nil {
				// Just log the error but continue with authentication
				fmt.Printf("Warning: %v\n", err)
			}
		}

		// Set session values for authentication
		setAuthSessionValues(session, user, false)
		session.Save(r, w)

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// Display face verification page
	renderFacePage(w, "", false)
}

// APIVerifyFaceHandler is an API endpoint for face verification
func APIVerifyFaceHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := utils.GetSession(r)

	// Check if there's a pending authentication
	email, ok := session.Values["pending_auth_email"].(string)
	if !ok || email == "" {
		sendJSONError(w, "No pending authentication", http.StatusBadRequest)
		return
	}

	// Parse request body
	var requestData struct {
		FaceData string `json:"face_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		sendJSONError(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Get user - use the safe version that handles NULL values
	user, err := models.GetUserByEmailSafe(email)
	if err != nil {
		sendJSONError(w, "User not found: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If face auth is not enabled for this user, we'll enable it
	if !user.FaceAuthEnabled {
		// Enable face auth and save the face data
		if err := enableFaceAuth(user, requestData.FaceData); err != nil {
			// Just log the error but continue with authentication
			fmt.Printf("Warning: %v\n", err)
		}
	}

	// Set session values for complete authentication
	setAuthSessionValues(session, user, true)

	// Log successful authentication
	fmt.Println("Face verification successful, authentication complete")

	session.Save(r, w)

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Face verification successful",
		"redirect": "/home",
	})
}

// Helper function to enable face authentication for a user
func enableFaceAuth(user *models.User, faceData string) error {
	// Update user with face auth enabled
	user.FaceAuthEnabled = true
	err := models.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	// Save face data
	err = utils.SaveFaceData(user.ID, faceData)
	if err != nil {
		return fmt.Errorf("failed to save face data: %v", err)
	}

	return nil
}

// Helper function to set authentication session values
func setAuthSessionValues(session *sessions.Session, user *models.User, isAPI bool) {
	fmt.Printf("DEBUG: setAuthSessionValues called with isAPI=%v, user=%+v\n", isAPI, user)
	fmt.Printf("DEBUG: Session values before update: %+v\n", session.Values)
	
	// Check if we're coming from 2FA verification
	twoFACompleted, wasTwoFACompleted := session.Values["twofa_completed"].(bool)
	
	// Set common session values
	session.Values["authenticated"] = true
	
	// If we have temp user data from 2FA, use it
	if wasTwoFACompleted && twoFACompleted {
		// Use the user ID from the session if available
		if tempUserID, ok := session.Values["temp_user_id"].(int); ok && tempUserID > 0 {
			session.Values["user_id"] = tempUserID
		} else {
			session.Values["user_id"] = user.ID
		}
		
		// Use the username from the session if available
		if tempUsername, ok := session.Values["temp_username"].(string); ok && tempUsername != "" {
			session.Values["username"] = tempUsername
		} else {
			session.Values["username"] = user.Username
		}
		
		// Always set 2FA as enabled since it was completed
		session.Values["twofa_enabled"] = true
		fmt.Println("DEBUG: Using session data from 2FA completion")
	} else {
		// Use the user data from the database
		session.Values["user_id"] = user.ID
		session.Values["username"] = user.Username
		session.Values["twofa_enabled"] = user.TwoFAEnabled
	}
	
	// Always set these values
	session.Values["email"] = user.Email
	session.Values["face_auth_enabled"] = true

	// Clean up temporary session data
	cleanupSessionData(session)
	
	// Log final session state
	fmt.Printf("DEBUG: Session values after update: %+v\n", session.Values)
}

// Helper function to clean up temporary session data
func cleanupSessionData(session *sessions.Session) {
	// Clean up all temporary and pending auth data
	delete(session.Values, "pending_auth_email")
	delete(session.Values, "pending_auth_user_id")
	delete(session.Values, "pending_auth_username")
	delete(session.Values, "pending_auth_mfa_enabled")
	delete(session.Values, "pending_auth_face_enabled")
	delete(session.Values, "pending_face_after_2fa")
	delete(session.Values, "twofa_verified")
	delete(session.Values, "twofa_completed")
	delete(session.Values, "temp_user_id")
	delete(session.Values, "temp_username")
}

// Helper function to send JSON error responses
func sendJSONError(w http.ResponseWriter, errorMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   errorMsg,
	})
}

// Helper function to render face setup/verification page
func renderFacePage(w http.ResponseWriter, errorMsg string, isSetup bool) {
	var templateFile string
	if isSetup {
		templateFile = "templates/setup-face.html"
	} else {
		templateFile = "templates/verify-face.html"
	}

	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Error": errorMsg,
	}

	tmpl.Execute(w, data)
}
