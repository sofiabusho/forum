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

	// Βρες τον χρήστη με βάση το token
	var userID int
	err := db.QueryRow("SELECT UserID FROM Users WHERE reset_token = ?", token).Scan(&userID)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Hash του νέου password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Ενημέρωσε τον χρήστη και αφαίρεσε το token
	_, err = db.Exec("UPDATE Users SET passwordHash = ?, reset_token = NULL WHERE UserID = ?", string(hashedPassword), userID)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}
