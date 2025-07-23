package utils

import (
	"fmt"
	"net/smtp"
)

func SendResetEmail(toEmail, token string) error {
	from := "plant.talk2025@gmail.com" // ➤ βάλε εδώ το Gmail σου
	password := "niicnftnethvawxf"     // ✅ Gmail App Password
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// ✅ Σωστό reset link προς add-newpassword.html
	resetLink := fmt.Sprintf("http://localhost:8080/add-newpassword.html?token=%s", token)

	// Μήνυμα email
	message := []byte(fmt.Sprintf(
		"Subject: Reset your password\n\nClick the link below to reset your password:\n%s", resetLink,
	))

	// SMTP Auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Αποστολή email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
	if err != nil {
		return err
	}
	return nil
}
