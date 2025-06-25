package utils

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

// Global session store
var sessionStore *sessions.CookieStore

// InitSessionStore initializes the global session store
func InitSessionStore() {
	key := os.Getenv("SESSION_KEY")
	if key == "" {
		key = "login-form-session-key-please-change-in-production"
		log.Printf("Warning: SESSION_KEY not set, using default key")
	}
	
	sessionStore = sessions.NewCookieStore([]byte(key))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
	
	log.Printf("Session store initialized")
}

// GetSession returns a session for the given request
func GetSession(r *http.Request) (*sessions.Session, error) {
	if sessionStore == nil {
		InitSessionStore()
	}
	session, err := sessionStore.Get(r, "auth-session")
	if err != nil {
		log.Printf("Error getting session: %v", err)
		// Try to recover by creating a new session
		session = sessions.NewSession(sessionStore, "auth-session")
		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7, // 7 days
			HttpOnly: true,
		}
		return session, nil
	}
	return session, nil
}

// SaveSession saves the session
func SaveSession(session *sessions.Session, w http.ResponseWriter, r *http.Request) error {
	err := session.Save(r, w)
	if err != nil {
		log.Printf("Error saving session: %v", err)
	}
	return err
}

// GetSessionStore returns the global session store
func GetSessionStore() sessions.Store {
	if sessionStore == nil {
		InitSessionStore()
	}
	return sessionStore
}

// GetSessionKey returns the session key from environment or a default one
func GetSessionKey() string {
	key := os.Getenv("SESSION_KEY")
	if key == "" {
		key = "login-form-session-key-please-change-in-production"
		log.Printf("Warning: SESSION_KEY not set, using default key")
	}
	return key
}

// FixSession attempts to repair a broken session
func FixSession(w http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	// First try to get the existing session
	session, err := GetSession(r)
	if err != nil {
		log.Printf("Error getting session in FixSession: %v, creating new session", err)
		// Create a new session
		session = sessions.NewSession(sessionStore, "auth-session")
		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7, // 7 days
			HttpOnly: true,
		}
	}
	
	// Save the session to ensure it's valid
	err = SaveSession(session, w, r)
	if err != nil {
		log.Printf("Error saving session in FixSession: %v", err)
		return session, err
	}
	
	log.Printf("Session fixed successfully, values: %+v", session.Values)
	return session, nil
}
