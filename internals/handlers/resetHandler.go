package handlers

import (
	"database/sql"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// GET: serve the reset form with token from the URL (?token=...)
	if r.Method == http.MethodGet {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		// Render template and pass token to {{.Token}}
		// (Assumes utils.FileService uses Go templates.)
		utils.FileService("add-newpassword.html", w, map[string]interface{}{"Token": token})
		return
	}

	// POST: perform the reset
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimSpace(r.FormValue("token"))
	newPassword := r.FormValue("newPassword")
	confirm := r.FormValue("confirmPassword")

	if token == "" || newPassword == "" || confirm == "" {
		http.Error(w, "Invalid or missing token/password", http.StatusBadRequest)
		return
	}
	if newPassword != confirm {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	// find user by token
	var userID int
	err := db.QueryRow("SELECT user_id FROM Users WHERE reset_token = ?", token).Scan(&userID)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// hash and store new password; clear token
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	_, err = db.Exec("UPDATE Users SET password_hash = ?, reset_token = NULL WHERE user_id = ?", string(hashed), userID)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}
