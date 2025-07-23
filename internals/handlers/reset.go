package handlers

import (
	"database/sql"
	"forum/internals/database"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/add-newpassword.html", http.StatusSeeOther)
		return
	}

	token := r.FormValue("token")
	newPassword := r.FormValue("newPassword")

	db := database.CreateTable()
	defer db.Close()

	// ✅ Σωστό: πεδίο user_id (και ΟΧΙ UserID)
	var userID int
	err := db.QueryRow("SELECT user_id FROM Users WHERE reset_token = ?", token).Scan(&userID)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Δημιουργία hash του νέου κωδικού
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// ✅ Ενημέρωση password και διαγραφή του token
	_, err = db.Exec("UPDATE Users SET password_hash = ?, reset_token = NULL WHERE user_id = ?", string(hashedPassword), userID)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	// Redirect στον χρήστη για login
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}
