package handlers

import (
	"context"
	"encoding/json"
	"forum/internals/database"
	"forum/internals/utils"
	"io"
	"log"
	"net/http"

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
	url := githubOauthConfig.AuthCodeURL("state-token")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	// Ανταλλαγή του code με access token
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		body, _ := io.ReadAll(r.Body)
		log.Println("GitHub token exchange error:", err)
		log.Println("Body:", string(body))
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	// Λήψη δεδομένων χρήστη από GitHub
	client := githubOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var githubUser struct {
		Login string `json:"login"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		http.Error(w, "Failed to parse user", http.StatusInternalServerError)
		return
	}

	// Εναλλακτική λήψη email αν είναι null
	if githubUser.Email == "" {
		emailResp, _ := client.Get("https://api.github.com/user/emails")
		defer emailResp.Body.Close()
		var emails []struct {
			Email   string `json:"email"`
			Primary bool   `json:"primary"`
		}
		if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
			for _, e := range emails {
				if e.Primary {
					githubUser.Email = e.Email
					break
				}
			}
		}
	}

	if githubUser.Email == "" {
		http.Error(w, "GitHub email not found", http.StatusInternalServerError)
		return
	}

	// === Βάση δεδομένων ===
	db := database.CreateTable()
	defer db.Close()

	var userID int
	err = db.QueryRow("SELECT user_id FROM Users WHERE email = ?", githubUser.Email).Scan(&userID)
	if err != nil {
		// Δημιουργία νέου χρήστη
		res, err := db.Exec("INSERT INTO Users (username, email, password_hash) VALUES (?, ?, ?)",
			githubUser.Login, githubUser.Email, "")
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		lastID, _ := res.LastInsertId()
		userID = int(lastID)
	}

	// === Δημιουργία Session ===
	cookieValue := utils.GenerateCookieValue()
	_, err = db.Exec("INSERT INTO Sessions (user_id, cookie_value, expiration_date) VALUES (?, ?, datetime('now', '+7 days'))",
		userID, cookieValue)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  cookieValue,
		Path:   "/",
		MaxAge: 60 * 60 * 24 * 7,
	})

	// Ανακατεύθυνση στην αρχική
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
