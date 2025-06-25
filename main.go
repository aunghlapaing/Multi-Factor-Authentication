package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aungh/login-form/database"
	"github.com/aungh/login-form/handlers"
	"github.com/aungh/login-form/middleware"
	"github.com/aungh/login-form/models"
	"github.com/aungh/login-form/utils"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// createDemoUserIfNeeded creates a demo user if no users exist
func createDemoUserIfNeeded() error {
	// Check if we already have users
	users, err := models.GetAllUsersSafe()
	if err != nil {
		return err
	}

	// If we have users, no need to create a demo user
	if len(users) > 0 {
		return nil
	}

	// Create a demo user
	hashedPassword, err := utils.HashPassword("password")
	if err != nil {
		return err
	}

	demoUser := models.User{
		Username:        "Tester",
		Nickname:        "DemoUser",
		Email:           "testing@sample.com",
		Role: 			 "admin",
		PasswordHash:    hashedPassword,
		TwoFAEnabled:    false,
		FaceAuthEnabled: false,
		CreatedAt:       time.Now(),
	}

	return models.CreateUser(&demoUser)
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using default configuration")
	}

	// Initialize SQLite database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()
	log.Println("SQLite database initialized successfully")
	
	// Run database migrations
	if err := database.MigrateDB(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	log.Println("Database migrations completed successfully")

	// Initialize face storage
	if err := utils.InitFaceStorage(); err != nil {
		log.Printf("Warning: Failed to initialize face storage: %v", err)
	}

	// Initialize session store
	utils.InitSessionStore()
	
	// Initialize OAuth with Goth
	utils.InitGothOAuth()

	// Create a demo user if none exists
	if err := createDemoUserIfNeeded(); err != nil {
		log.Printf("Warning: Failed to create demo user: %v", err)
	}

	// Initialize router
	r := mux.NewRouter()

	// Register static file directory
	staticDir := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticDir))

	// Register routes
	// Root route redirects to login page
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Clear any existing session to ensure login page is shown
		session, err := handlers.GetSession(r)
		if err == nil {
			session.Values["authenticated"] = false
			delete(session.Values, "user_id")
			delete(session.Values, "username")
			delete(session.Values, "email")
			session.Save(r, w)
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	// Test route to create a social login user
	r.HandleFunc("/test-social-user", func(w http.ResponseWriter, r *http.Request) {
		// Create a test user for social login
		hashedPassword, err := utils.HashPassword("test_password")
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		testUser := &models.User{
			Username:        "Social Test User",
			Nickname:        "SocialTester",
			Email:           "social_test@example.com",
			PasswordHash:    hashedPassword,
			GoogleID:        "test_google_id",
			GithubID:        "",
			ProfileImage:    "",
			TwoFASecret:     "",
			TwoFAEnabled:    false,
			FaceAuthEnabled: false,
			Role:            "user",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Try to create the user directly using database package
		query := `
		INSERT INTO users (
			username, nickname, email, password_hash, google_id, github_id, 
			profile_image, twofa_secret, twofa_enabled, face_auth_enabled, 
			role, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		result, err := database.DB.Exec(
			query,
			testUser.Username,
			testUser.Nickname,
			testUser.Email,
			testUser.PasswordHash,
			testUser.GoogleID,
			testUser.GithubID,
			testUser.ProfileImage,
			testUser.TwoFASecret,
			testUser.TwoFAEnabled,
			testUser.FaceAuthEnabled,
			testUser.Role,
			testUser.CreatedAt,
			testUser.UpdatedAt,
		)

		if err != nil {
			http.Error(w, "Failed to create test user: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get the user ID
		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Failed to get last insert ID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Test social login user created successfully with ID: %d", id)
	})

	// Public routes (no authentication required)
	r.HandleFunc("/login", handlers.LoginHandler)
	r.HandleFunc("/signup", handlers.SignupHandler)
	r.HandleFunc("/logout", handlers.LogoutHandler)
	r.HandleFunc("/auth/google", handlers.GoogleAuthHandler)
	r.HandleFunc("/auth/github", handlers.GithubAuthHandler)
	// Add explicit callback routes for all possible callback URLs
	r.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler)
	r.HandleFunc("/auth/github/callback", handlers.GithubCallbackHandler)

	// Routes that require basic authentication
	r.Handle("/home", middleware.RequireAuth(http.HandlerFunc(handlers.HomeHandler)))

	// 2FA routes
	r.HandleFunc("/verify-2fa", handlers.Verify2FAHandler) // No auth middleware as this is part of auth flow
	r.Handle("/setup-2fa", middleware.RequireAuth(http.HandlerFunc(handlers.Setup2FAHandler)))
	r.HandleFunc("/qrcode", handlers.QRCodeHandler) // QR code image endpoint

	// Face authentication routes
	r.HandleFunc("/verify-face", handlers.VerifyFaceHandler) // No auth middleware as this is part of auth flow
	r.Handle("/setup-face", middleware.RequireAuth(http.HandlerFunc(handlers.SetupFaceHandler)))
	r.HandleFunc("/api/verify-face", handlers.APIVerifyFaceHandler) // API endpoint for face verification

	// Keep old routes for backward compatibility
	r.HandleFunc("/verify-mfa", handlers.Verify2FAHandler)
	r.Handle("/setup-mfa", middleware.RequireAuth(http.HandlerFunc(handlers.Setup2FAHandler)))

	// Captcha route
	r.HandleFunc("/captcha-image", handlers.CaptchaImageHandler)

	// Admin routes - only user management
	r.Handle("/admin/users", middleware.RequireAuth(http.HandlerFunc(handlers.AdminUsersHandler)))

	// User settings route
	r.Handle("/user/settings", middleware.RequireAuth(http.HandlerFunc(handlers.UserSettingsHandler)))

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Start server
	fmt.Printf("Server is running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
