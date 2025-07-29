package handlers

import (
	"database/sql"
	"encoding/json"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strings"
	"time"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		utils.FileService("profile.html", w, nil)
	} else if r.Method == "POST" {
		updateProfile(w, r)
	}
}

func ProfileAPIHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := utils.GetUserIDFromSession(cookie.Value)
	if userID == 0 {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	profile := getUserProfile(db, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromSession(getCookieValue(r))
	if userID == 0 {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	newUsername := strings.TrimSpace(r.FormValue("username"))
	if newUsername == "" {
		http.Error(w, "Username cannot be empty", http.StatusBadRequest)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	var exists int
	db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = ? AND user_id != ?", newUsername, userID).Scan(&exists)
	if exists > 0 {
		http.Error(w, "Username already taken", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE Users SET username = ? WHERE user_id = ?", newUsername, userID)
	if err != nil {
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"username": newUsername,
	})
}

func getUserProfile(db *sql.DB, userID int) database.UserProfile {
	var profile database.UserProfile
	var registrationDate time.Time

	err := db.QueryRow(`
		SELECT user_id, username, email, registration_date 
		FROM Users 
		WHERE user_id = ?
	`, userID).Scan(&profile.UserID, &profile.Username, &profile.Email, &registrationDate)

	if err != nil {
		return profile
	}

	profile.JoinDate = registrationDate.Format("January 2, 2006")
	db.QueryRow("SELECT COUNT(*) FROM Posts WHERE user_id = ?", userID).Scan(&profile.PostCount)
	db.QueryRow("SELECT COUNT(*) FROM Comments WHERE user_id = ?", userID).Scan(&profile.CommentCount)
	db.QueryRow("SELECT COUNT(*) FROM LikesDislikes WHERE user_id = ? AND vote = 1", userID).Scan(&profile.LikesGiven)
	db.QueryRow(`
		SELECT COUNT(*) 
		FROM LikesDislikes ld 
		JOIN Posts p ON ld.post_id = p.post_id 
		WHERE p.user_id = ? AND ld.vote = 1
	`, userID).Scan(&profile.LikesReceived)

	return profile
}

func getCookieValue(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	return cookie.Value
}
