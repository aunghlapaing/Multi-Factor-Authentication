package middleware

import (
	"log"
	"net/http"

	"github.com/aungh/login-form/utils"
)

// RequireAuth middleware checks if a user is authenticated
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := utils.GetSession(r)
		if err != nil {
			log.Printf("Error getting session in RequireAuth middleware: %v", err)
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// Check if user is authenticated
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// User is authenticated, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// Require2FA middleware checks if a user has completed 2FA verification
func Require2FA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := utils.GetSession(r)
		if err != nil {
			log.Printf("Error getting session in Require2FA middleware: %v", err)
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// First check if user is authenticated
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check if 2FA is enabled for this user
		twoFAEnabled, ok := session.Values["twofa_enabled"].(bool)
		if ok && twoFAEnabled {
			// Check if 2FA is verified in this session
			twoFAVerified, ok := session.Values["twofa_verified"].(bool)
			if !ok || !twoFAVerified {
				http.Redirect(w, r, "/verify-2fa", http.StatusSeeOther)
				return
			}
		}

		// 2FA check passed, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// RequireFaceAuth middleware checks if a user has completed face authentication
func RequireFaceAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := utils.GetSession(r)
		if err != nil {
			log.Printf("Error getting session in RequireFaceAuth middleware: %v", err)
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// First check if user is authenticated
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Check if face auth is enabled for this user
		faceAuthEnabled, ok := session.Values["face_auth_enabled"].(bool)
		if ok && faceAuthEnabled {
			// Check if face auth is verified in this session
			faceAuthVerified, ok := session.Values["face_auth_verified"].(bool)
			if !ok || !faceAuthVerified {
				http.Redirect(w, r, "/verify-face", http.StatusSeeOther)
				return
			}
		}

		// Face auth check passed, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// RequireFullAuth middleware checks if a user has completed all required authentication steps
func RequireFullAuth(next http.Handler) http.Handler {
	return RequireAuth(Require2FA(RequireFaceAuth(next)))
}
