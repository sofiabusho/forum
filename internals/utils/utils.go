package utils

import (
	"forum/internals/database"
	"html/template"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

func FileService(filename string, w http.ResponseWriter, data any) {
	tmpl, err := template.ParseFiles("frontend/templates/" + filename)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), 500)
		return
	}
	tmpl.Execute(w, data)
}

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func IsValidPassword(password string) bool {
	return len(password) >= 5
}

func GenerateCookieValue() string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// isValid returns true if the given session cookie exists and is not expired.
func IsValidSession(cookieValue string) bool {
	db := database.CreateTable()
	defer db.Close()

	var expiration time.Time
	err := db.QueryRow(
		"SELECT expiration_date FROM Sessions WHERE cookie_value = ?",
		cookieValue,
	).Scan(&expiration)
	if err != nil {
		return false // not found, or some other DB error
	}
	return time.Now().Before(expiration)
}
