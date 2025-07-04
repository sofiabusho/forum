package handlers

import (
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.FileService("login.html", w, nil)
		return
	}

	db := database.CreateTable()
	emailOrUsername := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	var id int
	var hash string
	err := db.QueryRow("SELECT user_id, password_hash FROM Users WHERE email = ? OR username = ?", emailOrUsername, emailOrUsername).Scan(&id, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		utils.FileService("login.html", w, map[string]interface{}{"Messagelg": "Invalid credentials"})
		return
	}

	// Create cookie
	cookieValue := utils.GenerateCookieValue()
	exp := time.Now().Add(24 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    cookieValue,
		Path:     "/",
		Expires:  exp,
		HttpOnly: true,
	})
	database.Insert(db, "Sessions", "(user_id, cookie_value, expiration_date)", id, cookieValue, exp)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
