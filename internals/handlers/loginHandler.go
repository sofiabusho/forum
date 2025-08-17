package handlers

import (
	"fmt"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Check for success message from registration
		successType := r.URL.Query().Get("success")
		message := r.URL.Query().Get("message")
		
		data := make(map[string]interface{})
		
		if successType == "registration" {
			data["SuccessMessage"] = "Registration successful! Please log in with your new account."
		} else if message != "" {
			data["SuccessMessage"] = message
		}
		
		utils.FileService("login.html", w, data)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	emailOrUsername := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	var userID int
	var passwordHash string

	err := db.QueryRow(
		"SELECT user_id, password_hash FROM Users WHERE email = ? OR username = ?",
		emailOrUsername, emailOrUsername,
	).Scan(&userID, &passwordHash)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		utils.FileService("login.html", w, map[string]interface{}{
			"ErrorMessage": "Invalid email/username or password",
		})
		return
	}

	// Create secure session cookie
	cookieValue := utils.GenerateCookieValue()
	expiration := time.Now().Add(24 * time.Hour)

	// Set session cookie with SameSite
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    cookieValue,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, 
	})

	// Store session in database
	database.Insert(db, "Sessions", "(user_id, cookie_value, expiration_date)", userID, cookieValue, expiration)

	// Debug output
	fmt.Println("Login successful for user_id:", userID)
	fmt.Println("Session cookie set:", cookieValue)

	// Redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
