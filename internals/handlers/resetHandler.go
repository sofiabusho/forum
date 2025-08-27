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

func SendResetEmail(toEmail, token string) error {
	fmt.Printf("DEBUG: Attempting to send email to: %s\n", toEmail)
	fmt.Printf("DEBUG: Reset token: %s\n", token)

	from := "Plant Talk"
	password := "niicnftnethvawxf"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Reset link
	resetLink := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)
	fmt.Printf("DEBUG: Reset link: %s\n", resetLink)

	// Parse and execute the HTML template
	tmpl, err := template.ParseFiles("frontend/templates/email-reset.html")
	if err != nil {
		fmt.Printf("ERROR: Failed to parse template: %v\n", err)
		return fmt.Errorf("failed to parse email template: %v", err)
	}
	fmt.Println("DEBUG: Template parsed successfully")

	var body bytes.Buffer
	data := struct {
		ResetLink string
	}{
		ResetLink: resetLink,
	}

	// Execute template into buffer
	err = tmpl.Execute(&body, data)
	if err != nil {
		fmt.Printf("ERROR: Failed to execute template: %v\n", err)
		return fmt.Errorf("failed to execute template: %v", err)
	}
	fmt.Println("DEBUG: Template executed successfully")

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

	fmt.Printf("DEBUG: Attempting to connect to SMTP server: %s:%s\n", smtpHost, smtpPort)

	// SMTP Auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send email
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message.Bytes())
	if err != nil {
		fmt.Printf("ERROR: SMTP send failed: %v\n", err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	fmt.Println("DEBUG: Email sent successfully!")
	return nil
}
