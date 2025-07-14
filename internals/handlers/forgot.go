package handlers

import (
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/forgot-password.html", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	if !strings.Contains(email, "@") {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	// Δημιουργία reset token
	token := uuid.New().String()

	db := database.CreateTable()
	defer db.Close()

	// Αποθήκευση token στον χρήστη
	_, err := db.Exec("UPDATE Users SET reset_token = ? WHERE email = ?", token, email)
	if err != nil {
		http.Error(w, "Failed to store reset token", http.StatusInternalServerError)
		return
	}

	// Αποστολή email
	err = utils.SendResetEmail(email, token)
	if err != nil {
		http.Error(w, "Failed to send reset email", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
