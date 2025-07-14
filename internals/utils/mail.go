package utils

import (
	"fmt"
	"net/smtp"
)

func SendResetEmail(toEmail, token string) error {
	from := "plant.talk2025@gmail.com" // ➤ βάλε εδώ το Gmail σου
	password := "niicnftnethvawxf"     // ✅ χωρίς κενά
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Φτιάχνουμε το reset link
	resetLink := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)

	// Μήνυμα
	message := []byte(fmt.Sprintf("Subject: Reset your password\n\nClick the link below to reset your password:\n%s", resetLink))

	// Auth object
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Αποστολή
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
	if err != nil {
		return err
	}
	return nil
}
