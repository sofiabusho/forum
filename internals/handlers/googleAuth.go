package handlers

import (
	"context"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleapi "google.golang.org/api/oauth2/v2"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     "185487260225-ak7onmfqeiug4044f9j5valjah7d421r.apps.googleusercontent.com",
	ClientSecret: "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", // ⚠️ το νέο full secret
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	},
	Endpoint: google.Endpoint,
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	oauth2Service, err := googleapi.New(googleOauthConfig.Client(context.Background(), token))
	if err != nil {
		http.Error(w, "Failed to create Google API client", http.StatusInternalServerError)
		return
	}

	userinfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	email := userinfo.Email
	username := userinfo.Name

	// === Εισαγωγή ή εύρεση χρήστη ===
	db := database.CreateTable()
	defer db.Close()

	var userID int
	err = db.QueryRow("SELECT user_id FROM Users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		// Χρήστης δεν υπάρχει, δημιουργούμε
		res, err := db.Exec("INSERT INTO Users (username, email, password_hash) VALUES (?, ?, ?)", username, email, "")
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		lastID, _ := res.LastInsertId()
		userID = int(lastID)
	}

	// === Δημιουργία session ===
	cookieValue := utils.GenerateCookieValue()
	_, err = db.Exec("INSERT INTO Sessions (user_id, cookie_value, expiration_date) VALUES (?, ?, datetime('now', '+7 days'))",
		userID, cookieValue)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Αποστολή cookie στον client
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  cookieValue,
		Path:   "/",
		MaxAge: 60 * 60 * 24 * 7,
	})

	// Τέλος: redirect στην αρχική
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
