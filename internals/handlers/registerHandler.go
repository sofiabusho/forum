package handlers

import (
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
	if !utils.IsValidEmail(email) || !utils.IsValidPassword(pass) {
		utils.FileService("register.html", w, map[string]interface{}{"Messagesg": "Invalid input"})
		return
	}

	var exists int
	db.QueryRow("SELECT COUNT(*) FROM Users WHERE email=?", email).Scan(&exists)
	if exists > 0 {
		utils.FileService("register.html", w, map[string]interface{}{"Messagesg": "Email exists"})
		return
	}
	db.QueryRow("SELECT COUNT(*) FROM Users WHERE username=?", username).Scan(&exists)
	if exists > 0 {
		utils.FileService("register.html", w, map[string]interface{}{"Messagesg": "Username taken"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	database.Insert(db, "Users", "(username, email, password_hash)", username, email, string(hash))
	utils.FileService("login.html", w, map[string]interface{}{"Message": "Register successful"})
}
