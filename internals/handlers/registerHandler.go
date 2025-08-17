package handlers

import (
	"fmt"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

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

	formData := map[string]interface{}{
		"Username": username,
		"Email":    email,
	}

	// DEBUG: Print received values
	fmt.Printf("\n=== REGISTRATION DEBUG ===\n")
	fmt.Printf("Username: '%s'\n", username)
	fmt.Printf("Email: '%s'\n", email)
	fmt.Printf("Password length: %d\n", len(pass))
	fmt.Printf("Passwords match: %v\n", pass == confirm)

	if username == "" || email == "" || pass == "" || confirm == "" {
		formData["Message"] = "All fields are required"
		utils.FileService("register.html", w, formData)
		return
	}

	if pass != confirm {
		formData["Message"] = "Passwords do not match"
		utils.FileService("register.html", w, formData)
		return
	}

	// DEBUG: Test email validation step by step
	fmt.Printf("\n--- EMAIL VALIDATION DEBUG ---\n")
	fmt.Printf("Raw email: '%s'\n", email)
	fmt.Printf("Email length: %d\n", len(email))
	fmt.Printf("@ count: %d\n", strings.Count(email, "@"))
	emailValid := utils.IsValidEmail(email)

	if !emailValid {
		formData["Message"] = "Invalid email format"
		utils.FileService("register.html", w, formData)
		return
	
	}
	
	fmt.Printf("\n--- PASSWORD VALIDATION DEBUG ---\n")
	passValid := utils.IsValidPassword(pass)

	if !passValid {
		formData["Message"] = "Password must have: 8+ characters, 1 uppercase, 1 lowercase, 1 number, 1 symbol"
		utils.FileService("register.html", w, formData)
		return
	}

	// Check for duplicates
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM Users WHERE email = ?", email).Scan(&exists)
	if err != nil {
		formData["Message"] = "Database error occurred"
		utils.FileService("register.html", w, formData)		
		return
	}

	if exists > 0 {
		formData["Message"] = "This email is already registered"
		utils.FileService("register.html", w, formData)
		return
	}

	err = db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = ?", username).Scan(&exists)
	if err != nil {
		formData["Message"] = "Database error occurred"
		utils.FileService("register.html", w, formData)
		return
	}

	if exists > 0 {
		formData["Message"] = "This username is already taken"
		utils.FileService("register.html", w, formData)
		return
	}

	// Hash and insert
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		formData["Message"] = "Internal server error"
		utils.FileService("register.html", w, formData)
		return
	}

	_, err = db.Exec("INSERT INTO Users (username, email, password_hash) VALUES (?, ?, ?)", username, email, string(hash))
	if err != nil {
		formData["Message"] = "Failed to create user account"
		utils.FileService("register.html", w, formData)
		return
	}

	// Create welcome notification
	var newUserID int
	err = db.QueryRow("SELECT user_id FROM Users WHERE email = ?", email).Scan(&newUserID)
	if err == nil && newUserID > 0 {
		title := "Welcome to Plant Talk! 🌱"
		message := fmt.Sprintf("Welcome to our plant-loving community, %s! Start by creating your first post or exploring different plant categories. Happy growing!", username)

		CreateNotification(newUserID, "system", title, message, nil, nil, nil)
	}

	fmt.Println("✅ Registration successful!")

	http.Redirect(w, r, "/login?success=registration", http.StatusSeeOther)
}
