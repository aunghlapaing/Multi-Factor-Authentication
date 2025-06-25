# Login Form with Go

A modern web application with comprehensive authentication features built using Go.

## Features

- Modern UI for Sign In and Sign Up Pages
- Client-side and Server-side Form Validation
- Strong Password Checking with Visual Strength Indicator
- CAPTCHA Protection for Registration and Login
- Social Login Integration (Google and GitHub)
- Two-Factor Authentication (2FA) using TOTP
- Face Authentication
- Support for Multiple Authentication Methods
- Persistent Data Storage with SQLite
- Admin Account Management

## Setup and Installation

1. Make sure you have Go installed (version 1.16+)
2. Clone this repository
3. Install dependencies:
   ```
   go mod download
   ```
4. Set up environment variables (create a `.env` file based on `.env.example`)
   - Configure OAuth credentials for Google and GitHub
   - Set session secret key
5. Run the application:
   ```
   go run main.go
   ```
6. Open your browser and navigate to `http://localhost:8081`

## Admin Account Login For testing

1. gmail - testing@sample.com
2. password - password

## Usage

### Authentication Methods

#### Email/Password Login
1. Register a new account using the Sign Up page
2. Log in with your email and password

#### Social Login
1. Click on the Google or GitHub button on the login page
2. Authorize the application to access your account

#### Two-Factor Authentication (2FA)
1. Enable 2FA from your account settings
2. Scan the QR code with an authenticator app (like Google Authenticator)
3. Enter the verification code to complete setup
4. For future logins, you'll need to enter the code from your authenticator app

#### Face Authentication
1. Enable Face Authentication from your account settings
2. Allow camera access and follow the prompts to register your face
3. For future logins, you'll need to verify your identity using your camera
esting@sample.com
- `main.go`: Entry point of the application
- `handlers/`: HTTP request handlers
  - `simple.go`: Basic authentication handlers
  - `oauth.go`: Social login handlers
  - `face.go`: Face authentication handlers
  - `2fa.go`: Two-factor authentication handlers
- `models/`: Data models
  - `user.go`: User model and database operations
- `middleware/`: Middleware functions
- `database/`: Database configuration and migrations
- `static/`: Static assets (CSS, JS)
- `templates/`: HTML templates
- `utils/`: Utility functions
  - `session.go`: Session management utilities

## Technology Stack

### Backend
- **Language**: Go (Golang)
- **Web Framework**: Standard Go HTTP library with gorilla/mux for routing
- **Session Management**: gorilla/sessions
- **Authentication**: bcrypt for password hashing, TOTP for 2FA
- **Database**: SQLite 

### Frontend
- **Languages**: HTML, CSS, JavaScript
- **Libraries**: jQuery for DOM manipulation
- **CSS Framework**: Bootstrap for responsive design
- **Face Recognition**: JavaScript face-api.js library

### Architecture
- **Pattern**: Layered architecture with clear separation of concerns
- **Handlers Layer**: Processes HTTP requests and responses
- **Models Layer**: Manages data and business logic
- **Template Layer**: Handles view rendering
- **API Style**: RESTful endpoints for authentication operations
- **Templating**: Go's html/template package for server-side rendering
- **State Management**: Server-side sessions with client-side cookies

## Database

The application uses SQLite for data persistence. The database file is located at `data/users.db`. All user data, including authentication settings, are stored in this database.

## Security Features

- Password hashing using bcrypt
- Session-based authentication
- CSRF protection
- Multiple authentication factors (2FA and Face Authentication)
- Secure session management
- CAPTCHA protection to prevent automated attacks
- Strong password enforcement:
  - Minimum length requirements
  - Requires combination of uppercase, lowercase, numbers, and special characters
  - Visual password strength indicator
  - Prevents common passwords and patterns
- Comprehensive input validation:
  - Email format validation
  - Username format validation
  - Protection against SQL injection
  - XSS prevention

## Known Issues and Limitations

- Face authentication requires camera access and may not work on all browsers
- Social login requires valid OAuth credentials to be configured

## License

This project is licensed under the MIT License