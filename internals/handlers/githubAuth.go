package handlers

import (
	"context"
	"encoding/json"
	"forum/internals/database"
	"forum/internals/utils"
	"io"
	"log"
	"net/http"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var githubOauthConfig = &oauth2.Config{
	ClientID:     "Ov23liuXrs3tVsAW33pj",
	ClientSecret: "705fd4bf329d7e3c58d8fc1fea4b88283db935d6",
	RedirectURL:  "http://localhost:8080/auth/github/callback",
	Scopes:       []string{"user:email"},
	Endpoint:     github.Endpoint,
}

func GitHubLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		url := githubOauthConfig.AuthCodeURL("state-token")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GitHubCallback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGitHubCallback(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	// Exchange the code for access token
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		body, _ := io.ReadAll(r.Body)
		log.Println("GitHub token exchange error:", err)
		log.Println("Body:", string(body))
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	// Get user data from GitHub
	client := githubOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		log.Println("Failed to get GitHub user:", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var githubUser struct {
		Login string `json:"login"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		log.Println("Failed to parse GitHub user:", err)
		http.Error(w, "Failed to parse user", http.StatusInternalServerError)
		return
	}

	if githubUser.Email == "" {
		email := fetchGitHubUserEmail(client)
		if email != "" {
			githubUser.Email = email
		}
	}

	// Still require email for our forum
	if githubUser.Email == "" {
		log.Println("GitHub email not found for user:", githubUser.Login)
		http.Error(w, "GitHub email not found or not public", http.StatusBadRequest)
		return
	}

	// Database operations with single-session enforcement
	userID, isNewUser, err := createOrGetUser(githubUser.Login, githubUser.Email)
	if err != nil {
		log.Println("Database error creating/getting user:", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	if isNewUser {
		title := "Welcome to Plant Talk! 🌱"
		message := fmt.Sprintf("Welcome to our plant-loving community, %s! Start by creating your first post or exploring different plant categories. Happy growing!", githubUser.Login)
		CreateNotification(userID, "system", title, message, nil, nil, nil)
	}


	// Create session and redirect
	if err := createUserSession(w, userID); err != nil {
		log.Println("Failed to create session:", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Redirect to homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Separate function for email fetching
func fetchGitHubUserEmail(client *http.Client) string {
	emailResp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		log.Println("Failed to fetch GitHub user emails:", err)
		return ""
	}
	defer emailResp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
		log.Println("Failed to decode GitHub emails:", err)
		return ""
	}

	// Look for primary email first, then any email
	for _, e := range emails {
		if e.Primary {
			return e.Email
		}
	}

	// If no primary email, return the first one
	if len(emails) > 0 {
		return emails[0].Email
	}

	return ""
}

// Separate function for user creation/retrieval
func createOrGetUser(username, email string) (int, bool, error) {
	db := database.CreateTable()
	defer db.Close()

	var userID int
	err := db.QueryRow("SELECT user_id FROM Users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		// Create new user if not exists
		res, err := db.Exec("INSERT INTO Users (username, email, password_hash) VALUES (?, ?, ?)",
			username, email, "")
		if err != nil {
			return 0, false, err
		}
		lastID, _ := res.LastInsertId()
		userID = int(lastID)
		return userID, true, nil
	}

	return userID, false, nil 
}

// Separate function for session creation
func createUserSession(w http.ResponseWriter, userID int) error {
	db := database.CreateTable()
	defer db.Close()

	// Cleanup old sessions for this user (single-session enforcement)
	if _, err := db.Exec("DELETE FROM Sessions WHERE user_id = ?", userID); err != nil {
		log.Printf("Warning: Failed to cleanup old sessions for user %d: %v\n", userID, err)
	}

	// Create new session
	cookieValue := utils.GenerateCookieValue()
	if _, err := db.Exec("INSERT INTO Sessions (user_id, cookie_value, expiration_date) VALUES (?, ?, datetime('now', '+7 days'))",
		userID, cookieValue); err != nil {
		return err
	}

	// Set secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    cookieValue,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7, // 7 days
		HttpOnly: true,              // Prevent XSS attacks
		SameSite: http.SameSiteLaxMode, // CSRF protection
	})

	return nil
}