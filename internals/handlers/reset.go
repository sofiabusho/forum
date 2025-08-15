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
	// GET: δείξε τη φόρμα με το token από το URL (π.χ. /reset-password?token=XYZ)
	if r.Method == http.MethodGet {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		// σερβίρουμε template και περνάμε το token για το hidden input {{.Token}}
		utils.FileService("add-newpassword.html", w, map[string]interface{}{"Token": token})
		return
	}

	// POST: αποθήκευση νέου κωδικού
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimSpace(r.FormValue("token"))
	newPassword := r.FormValue("newPassword")

	if token == "" || newPassword == "" {
		http.Error(w, "Invalid or missing token/password", http.StatusBadRequest)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	// Βρες τον χρήστη από το token
	var userID int
	err := db.QueryRow("SELECT user_id FROM Users WHERE reset_token = ?", token).Scan(&userID)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Hash νέου κωδικού
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Ενημέρωση password & καθάρισμα token
	_, err = db.Exec("UPDATE Users SET password_hash = ?, reset_token = NULL WHERE user_id = ?", string(hashedPassword), userID)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	// Επιτυχία → login
	http.Redirect(w, r, "/login.html", http.StatusSeeOther)
}
