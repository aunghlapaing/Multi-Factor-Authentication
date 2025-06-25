package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/aungh/login-form/models"
	"github.com/aungh/login-form/utils"
)

// AdminUsersHandler displays all users and provides management options
func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
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
	
	// Get current user to check if they're the first user (admin)
	currentUser, err := models.GetUserByIDSafe(userID)
	if err != nil {
		http.Error(w, "Failed to get current user", http.StatusInternalServerError)
		return
	}
	
	// Check if user has admin role
	if currentUser.Role != "admin" {
		log.Printf("User %d attempted to access admin area but has role %s", currentUser.ID, currentUser.Role)
		http.Error(w, "Unauthorized - Admin access required", http.StatusForbidden)
		return
	}
	
	log.Printf("Admin access granted to user %d with role %s", currentUser.ID, currentUser.Role)
	
	// Handle user deletion
	if r.Method == "POST" {
		action := r.FormValue("action")
		
		if action == "delete" {
			userIDToDelete, err := strconv.Atoi(r.FormValue("user_id"))
			if err != nil {
				http.Error(w, "Invalid user ID", http.StatusBadRequest)
				return
			}
			
			// Don't allow deleting yourself
			if userIDToDelete == userID {
				http.Error(w, "Cannot delete your own account", http.StatusBadRequest)
				return
			}
			
			// Check if user exists before deletion
			_, err = models.GetUserByIDSafe(userIDToDelete)
			if err != nil {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			
			// Delete the user
			err = models.DeleteUser(userIDToDelete)
			if err != nil {
				// Check if we need to delete face auth data
				if err.Error() == "face_auth_data_needs_deletion" {
					// Delete face authentication data
					if err := utils.DeleteFaceData(userIDToDelete); err != nil {
						http.Error(w, fmt.Sprintf("Failed to delete face data: %v", err), http.StatusInternalServerError)
						return
					}
				} else {
					http.Error(w, fmt.Sprintf("Failed to delete user: %v", err), http.StatusInternalServerError)
					return
				}
			}
			
			// Redirect to refresh the page
			http.Redirect(w, r, "/admin/users?deleted=true", http.StatusSeeOther)
			return
		}
	}
	
	// Get all users - use the safe version that handles NULL values
	users, err := models.GetAllUsersSafe()
	if err != nil {
		log.Printf("Failed to get users: %v", err)
		http.Error(w, "Failed to get users: " + err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Render the admin users page
	tmpl, err := template.ParseFiles("templates/admin-users.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := map[string]interface{}{
		"Users":       users,
		"CurrentUser": currentUser,
		"Deleted":     r.URL.Query().Get("deleted") == "true",
	}
	
	tmpl.Execute(w, data)
}
