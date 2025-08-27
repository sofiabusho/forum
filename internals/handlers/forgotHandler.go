package handlers

import (
	"database/sql"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// without POST, just show the forgot password page
		http.Redirect(w, r, "/forgot-password.html", http.StatusSeeOther)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if !utils.IsValidEmail(email) {
		// 404 if email is invalid
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	// if user exists, generate reset token
	var userID int
	err := db.QueryRow("SELECT user_id FROM Users WHERE email = ?", email).Scan(&userID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// create a reset token
	token := uuid.New().String()

	// save the token in the database
	_, _ = db.Exec("UPDATE Users SET reset_token = ? WHERE email = ?", token, email)

	// send email (best-effort)
	_ = SendResetEmail(email, token)

	// redirect to success page
	http.Redirect(w, r, "/forgot-password.html?sent=1", http.StatusSeeOther)
}
