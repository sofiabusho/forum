package utils

import (
	"forum/internals/database"
	"html/template"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

// TemplateData holds data to pass to templates
type TemplateData struct {
	IsLoggedIn   bool
	Username     string
	UserID       int
	Message      string
	Error        string
	Data         interface{}
}

func FileService(filename string, w http.ResponseWriter, data any) {
	tmpl, err := template.ParseFiles("frontend/templates/" + filename)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), 500)
		return
	}
	tmpl.Execute(w, data)
}

// FileServiceWithAuth serves templates with authentication context
func FileServiceWithAuth(filename string, w http.ResponseWriter, r *http.Request, data interface{}) {
	templateData := &TemplateData{
		Data: data,
	}
	
	// Check if user is logged in
	if cookie, err := r.Cookie("session"); err == nil && IsValidSession(cookie.Value) {
		templateData.IsLoggedIn = true
		templateData.UserID = GetUserIDFromSession(cookie.Value)
		templateData.Username = GetUsernameFromSession(cookie.Value)
	}
	
	tmpl, err := template.ParseFiles("frontend/templates/" + filename)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), 500)
		return
	}
	tmpl.Execute(w, templateData)
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

// IsValidSession returns true if the given session cookie exists and is not expired.
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

// GetUserIDFromSession returns the user ID for a given session cookie
func GetUserIDFromSession(cookieValue string) int {
	db := database.CreateTable()
	defer db.Close()

	var userID int
	err := db.QueryRow("SELECT user_id FROM Sessions WHERE cookie_value = ? AND expiration_date > datetime('now')", cookieValue).Scan(&userID)
	if err != nil {
		return 0
	}
	return userID
}

// GetUsernameFromSession returns the username for a given session cookie
func GetUsernameFromSession(cookieValue string) string {
	db := database.CreateTable()
	defer db.Close()

	var username string
	err := db.QueryRow(`
		SELECT u.username 
		FROM Users u 
		JOIN Sessions s ON u.user_id = s.user_id 
		WHERE s.cookie_value = ? AND s.expiration_date > datetime('now')
	`, cookieValue).Scan(&username)
	if err != nil {
		return ""
	}
	return username
}

// CheckAuth is a middleware to check if user is authenticated
func CheckAuth(r *http.Request) (bool, int, string) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false, 0, ""
	}
	
	if !IsValidSession(cookie.Value) {
		return false, 0, ""
	}
	
	userID := GetUserIDFromSession(cookie.Value)
	username := GetUsernameFromSession(cookie.Value)
	
	return true, userID, username
}