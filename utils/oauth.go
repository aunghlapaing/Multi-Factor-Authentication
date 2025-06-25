package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

// CustomBeginAuthHandler is a replacement for gothic.BeginAuthHandler that fixes session issues
func CustomBeginAuthHandler(w http.ResponseWriter, r *http.Request, provider string) {
	log.Printf("Starting custom OAuth flow for provider: %s", provider)
	
	// Create a new clean session
	session, err := FixSession(w, r)
	if err != nil {
		log.Printf("Error fixing session in CustomBeginAuthHandler: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	
	// Store the provider in our session
	session.Values["oauth_provider"] = provider
	
	// Save the session
	err = SaveSession(session, w, r)
	if err != nil {
		log.Printf("Error saving session in CustomBeginAuthHandler: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	
	// Get the provider from Goth
	gothProvider, err := goth.GetProvider(provider)
	if err != nil {
		log.Printf("Error getting provider: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Get the authentication URL
	authURL, err := gothProvider.BeginAuth(gothic.SetState(r))
	if err != nil {
		log.Printf("Error beginning auth: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Get the authorization URL
	url, err := authURL.GetAuthURL()
	if err != nil {
		log.Printf("Error getting auth URL: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Store the state in our session
	session.Values["oauth_state"] = gothic.GetState(r)
	err = SaveSession(session, w, r)
	if err != nil {
		log.Printf("Error saving state in session: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	
	// Redirect to the authorization URL
	log.Printf("Redirecting to auth URL: %s", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// CustomCompleteUserAuth is a replacement for gothic.CompleteUserAuth that fixes session issues
func CustomCompleteUserAuth(w http.ResponseWriter, r *http.Request, provider string) (goth.User, error) {
	log.Printf("Completing custom OAuth flow for provider: %s", provider)
	
	// Get our session
	session, err := FixSession(w, r)
	if err != nil {
		log.Printf("Error fixing session in CustomCompleteUserAuth: %v", err)
		return goth.User{}, err
	}
	
	// Check if we have a state in our session
	stateFromSession, ok := session.Values["oauth_state"].(string)
	if !ok || stateFromSession == "" {
		log.Printf("No state found in session")
		// Try to proceed anyway
	}
	
	// Get the state from the request
	stateFromRequest := r.URL.Query().Get("state")
	if stateFromRequest == "" {
		log.Printf("No state found in request")
	}
	
	// Log the states for debugging
	log.Printf("State from session: %s", stateFromSession)
		log.Printf("State from request: %s", stateFromRequest)
	
	// Get the provider
	gothProvider, err := goth.GetProvider(provider)
	if err != nil {
		log.Printf("Error getting provider: %v", err)
		return goth.User{}, err
	}
	
	// Get the code from the request
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("No code found in request")
		return goth.User{}, fmt.Errorf("no code provided")
	}
	
	// Exchange the code for a token
	// If we don't have a state from the session, try to create a new session
	var value goth.Session
	if stateFromSession == "" {
		log.Printf("No state in session, creating a new session")
		// Create a new session with the provider
		var beginErr error
		value, beginErr = gothProvider.BeginAuth(stateFromRequest)
		if beginErr != nil {
			log.Printf("Error beginning auth: %v", beginErr)
			return goth.User{}, beginErr
		}
	} else {
		// Use the state from the session
		var err error
		value, err = gothProvider.UnmarshalSession(stateFromSession)
		if err != nil {
			log.Printf("Error unmarshalling session: %v, trying with empty state", err)
			// Try with an empty state as a fallback
			var beginErr error
			value, beginErr = gothProvider.BeginAuth(stateFromRequest)
			if beginErr != nil {
				log.Printf("Error beginning auth: %v", beginErr)
				return goth.User{}, beginErr
			}
		}
	}
	
	// Complete the auth process
	// Log the query parameters for debugging
	log.Printf("Query parameters for authorization: %+v", r.URL.Query())
	
	// Make sure we have the code parameter
	if code == "" {
		log.Printf("No code parameter found in request")
		return goth.User{}, fmt.Errorf("no code parameter found in request")
	}
	
	// Create a new query with just the code parameter to avoid issues with state
	q := url.Values{}
	q.Set("code", code)
	
	// Try to authorize with just the code parameter
	_, err = value.Authorize(gothProvider, q)
	if err != nil {
		log.Printf("Error authorizing with code only: %v, trying with full query", err)
		// Try with the full query as a fallback
		_, err = value.Authorize(gothProvider, r.URL.Query())
		if err != nil {
			log.Printf("Error authorizing with full query: %v", err)
			return goth.User{}, err
		}
	}
	
	// Get the user
	user, err := gothProvider.FetchUser(value)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return goth.User{}, err
	}
	
	// Set the provider in the user
	user.Provider = provider
	
	// Clear the state from the session
	delete(session.Values, "oauth_state")
	err = SaveSession(session, w, r)
	if err != nil {
		log.Printf("Error saving session in CustomCompleteUserAuth: %v", err)
		// Continue anyway
	}
	
	log.Printf("Successfully completed OAuth flow for user: %s", user.Email)
	return user, nil
}

// InitGothOAuth initializes OAuth configurations using Goth
func InitGothOAuth() {
	// Make sure the session store is initialized first
	InitSessionStore()
	// Print environment variables for debugging
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	log.Printf("Google OAuth Configuration - Client ID length: %d, Secret length: %d",
		len(googleClientID),
		len(googleClientSecret))
	log.Printf("GitHub OAuth Configuration - Client ID length: %d, Secret length: %d",
		len(githubClientID),
		len(githubClientSecret))
	log.Printf("Using redirect URI for Google: http://localhost:8081/auth/google/callback")
	log.Printf("Using redirect URI for GitHub: http://localhost:8081/auth/github/callback")
	// Set up Goth session store with consistent configuration
	log.Printf("Setting up gothic.Store with our session store")
	gothic.Store = GetSessionStore()
	
	// Note: gothic uses "_gothic_session" as the session name by default
	// We'll handle this with our custom auth handlers

	// Important: Configure gothic to use our custom provider name function
	gothic.GetProviderName = func(req *http.Request) (string, error) {
		// First check the URL path
		path := req.URL.Path
		log.Printf("GetProviderName called for path: %s", path)
		
		if strings.Contains(path, "/auth/google") {
			log.Printf("Provider detected from path: google")
			return "google", nil
		} else if strings.Contains(path, "/auth/github") {
			log.Printf("Provider detected from path: github")
			return "github", nil
		}

		// If not found in path, check the session
		session, err := gothic.Store.Get(req, "auth-session")
		if err != nil {
			log.Printf("Error getting session in GetProviderName: %v", err)
		}
		
		log.Printf("Session values in GetProviderName: %+v", session.Values)
		
		if provider, ok := session.Values["oauth_provider"].(string); ok && provider != "" {
			log.Printf("Provider detected from session: %s", provider)
			return provider, nil
		}

		// If still not found, return an error
		log.Printf("No provider detected, returning error")
		return "", fmt.Errorf("you must select a provider")
	}

	// Set up OAuth providers
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			"http://localhost:8081/auth/google/callback",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		),
		github.New(
			os.Getenv("GITHUB_CLIENT_ID"),
			os.Getenv("GITHUB_CLIENT_SECRET"),
			"http://localhost:8081/auth/github/callback",
			"user",
			"user:email",
		),
	)

	// Print OAuth configurations for debugging
	providers := goth.GetProviders()
	for provider := range providers {
		fmt.Printf("OAuth provider configured: %s\n", provider)
	}
}
