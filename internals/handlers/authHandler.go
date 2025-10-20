package handlers

import (
	"encoding/json"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
)

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		db := database.CreateTable()
		defer db.Close()
		db.Exec("DELETE FROM Sessions WHERE cookie_value = ?", cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// AuthStatusHandler checks authentication status for API calls
func AuthStatusHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	isLoggedIn := err == nil && utils.IsValidSession(cookie.Value)

	w.Header().Set("Content-Type", "application/json")
	if isLoggedIn {
		userID := utils.GetUserIDFromSession(cookie.Value)
		username := utils.GetUsernameFromSession(cookie.Value)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"loggedIn": true,
			"userID":   userID,
			"username": username,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"loggedIn": false,
		})
	}
}
