package handlers

import (
	"fmt"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// isValidPassword enforces:
// - at least 1 lowercase
// - at least 1 uppercase
// - at least 1 number
// - at least 1 symbol
// - length: >= 8 characters
func isValidPassword(password string) bool {
	pattern := `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[\W_]).{8,}$`
	ok, _ := regexp.MatchString(pattern, password)
	return ok
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.FileService("register.html", w, nil)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	pass := r.FormValue("password")
	confirm := r.FormValue("confirmPassword")

	if username == "" || email == "" || pass == "" || confirm == "" {
		utils.FileService("register.html", w, map[string]interface{}{"Messagesg": "All fields required"})
		return
	}

	if pass != confirm {
		utils.FileService("register.html", w, map[string]interface{}{"Messagesg": "Passwords do not match"})
		return
	}

	// Email and password validation
	if !utils.IsValidEmail(email) || !isValidPassword(pass) {
		utils.FileService("register.html", w, map[string]interface{}{"Messagesg": "Invalid input"})
		return
	}

	// Check for duplicates
	var exists int
	db.QueryRow("SELECT COUNT(*) FROM Users WHERE email = ?", email).Scan(&exists)
	if exists > 0 {
		utils.FileService("register.html", w, map[string]interface{}{"Messagesg": "Email exists"})
		return
	}
	db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = ?", username).Scan(&exists)
	if exists > 0 {
		utils.FileService("register.html", w, map[string]interface{}{"Messagesg": "Username taken"})
		return
	}

	// Hash and insert
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	database.Insert(db, "Users", "(username, email, password_hash)", username, email, string(hash))

	// Create welcome notification
	var newUserID int
	err := db.QueryRow("SELECT user_id FROM Users WHERE email = ?", email).Scan(&newUserID)
	if err == nil && newUserID > 0 {
		title := "Welcome to Plant Talk! 🌱"
		message := fmt.Sprintf("Welcome to our plant-loving community, %s! Start by creating your first post or exploring different plant categories. Happy growing!", username)

		CreateNotification(newUserID, "system", title, message, nil, nil, nil)
	}

	utils.FileService("login.html", w, map[string]interface{}{"Message": "Registration successful! Please log in."})
}
