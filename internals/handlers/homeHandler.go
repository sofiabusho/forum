package handlers

import (
	"forum/internals/utils"
	"net/http"
)

// HomeHandler handles the main homepage route
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		NotFoundHandler(w, r)
		return
	}

	if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
		utils.FileService("index-signed.html", w, nil)
	} else {
		utils.FileService("index-unsigned.html", w, nil)
	}
}

// ViewPostHandler serves the view-post page
func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	utils.FileService("view-post.html", w, nil)
}

// CategoriesPageHandler serves the categories page
func CategoriesPageHandler(w http.ResponseWriter, r *http.Request) {
	utils.FileService("categories.html", w, nil)
}

// Static page handlers
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	utils.FileService("about.html", w, nil)
}

func TermsHandler(w http.ResponseWriter, r *http.Request) {
	utils.FileService("Terms&Conditions.html", w, nil)
}

func ForgotPasswordPageHandler(w http.ResponseWriter, r *http.Request) {
	utils.FileService("forgot-password.html", w, nil)
}

func ProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
		utils.FileService("profile.html", w, nil)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func NotificationsPageHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
		utils.FileService("notifications.html", w, nil)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
