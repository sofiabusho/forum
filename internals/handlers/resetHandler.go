package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"forum/internals/database"
	"forum/internals/utils"
	"html/template"
	"net/http"
	"net/smtp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// GET: serve the reset form with token from the URL (?token=...)
	if r.Method == http.MethodGet {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			// Show initial reset request form
			utils.FileService("request-reset.html", w, nil)
			return
		}
		// Show password reset form with token
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
		showErrorOnPage(w, token, "All fields are required")
		return
	}
	if newPassword != confirm {
		showErrorOnPage(w, token, "Passwords do not match")
		return
	}

	db := database.CreateTable()
	defer db.Close()

	// find user by token and get current password
	var userID int
	var currentHashedPassword string
	err := db.QueryRow("SELECT user_id, password_hash FROM Users WHERE reset_token = ?", token).Scan(&userID, &currentHashedPassword)
	if err == sql.ErrNoRows {
		showErrorOnPage(w, token, "Invalid or expired reset link")
		return
	} else if err != nil {
		showErrorOnPage(w, token, "Database error. Please try again.")
		return
	}

	if currentHashedPassword == "" {
		showErrorOnPage(w, token, "User password not found")
		return
	}

	// Check if new password is the same as current password
	compareErr := bcrypt.CompareHashAndPassword([]byte(currentHashedPassword), []byte(newPassword))
	if compareErr == nil {
		// Passwords match - new password is the same as current
		showErrorOnPage(w, token, "New password cannot be the same as current password")
		return
	}

	// hash and store new password; clear token
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		showErrorOnPage(w, token, "Failed to process password")
		return
	}
	_, err = db.Exec("UPDATE Users SET password_hash = ?, reset_token = NULL WHERE user_id = ?", string(hashed), userID)
	if err != nil {
		showErrorOnPage(w, token, "Failed to update password")
		return
	}

	http.Redirect(w, r, "/login.html?message=password_reset_success", http.StatusSeeOther)
}

func SendResetEmail(toEmail, token string) error {

	from := "plant.talk2025@gmail.com"
	password := "niicnftnethvawxf"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Reset link
	resetLink := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)

	// Parse and execute the HTML template
	tmpl, err := template.ParseFiles("frontend/templates/email-reset.html")
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body bytes.Buffer
	data := struct {
		ResetLink string
	}{
		ResetLink: resetLink,
	}

	// Execute template into buffer
	err = tmpl.Execute(&body, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Create proper email headers with HTML content
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = toEmail
	headers["Subject"] = "Password Reset Request"
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Build the email message
	var message bytes.Buffer
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.Write(body.Bytes())

	// SMTP Auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send email
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// Helper function to show errors on the same page
func showErrorOnPage(w http.ResponseWriter, token, errorMessage string) {
	data := map[string]interface{}{
		"Token":        token,
		"ErrorMessage": errorMessage,
	}
	utils.FileService("add-newpassword.html", w, data)
}
