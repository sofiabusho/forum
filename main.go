package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/internals/database"
	"forum/internals/handlers"
	"forum/internals/utils"
	"io/ioutil"
	"net/http"
)

func main() {
	// Initialize database
	initializeDatabase()

	fmt.Println("Server running on http://localhost:8080")

	// Authentication routes
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/login.html", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/register.html", handlers.RegisterHandler)
	http.HandleFunc("/logout", logoutHandler)

	// Post routes
	http.HandleFunc("/new-post", handlers.CreatePostHandler)
	http.HandleFunc("/new-post.html", handlers.CreatePostHandler)
	http.HandleFunc("/api/posts", handlers.PostsAPIHandler)

	// Comment routes
	http.HandleFunc("/api/comments/create", handlers.CreateCommentHandler)
	http.HandleFunc("/api/comments", handlers.CommentsAPIHandler)
	http.HandleFunc("/api/comments/delete", handlers.DeleteCommentHandler)

	// Like/Dislike routes
	http.HandleFunc("/api/posts/like", handlers.LikePostHandler)
	http.HandleFunc("/api/comments/like", handlers.LikeCommentHandler)

	// Category routes
	http.HandleFunc("/api/categories", handlers.CategoriesAPIHandler)
	http.HandleFunc("/categories", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("categories.html", w, nil)
	})
	http.HandleFunc("/categories.html", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("categories.html", w, nil)
	})

	// Filter routes (for authenticated users)
	http.HandleFunc("/api/posts/filtered", handlers.FilteredPostsHandler)

	// Auth status check - ONLY ONE REGISTRATION
	http.HandleFunc("/api/auth/status", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		isLoggedIn := err == nil && utils.IsValidSession(cookie.Value)

		w.Header().Set("Content-Type", "application/json")
		if isLoggedIn {
			userID := utils.GetUserIDFromSession(cookie.Value)
			username := utils.GetUsernameFromSession(cookie.Value)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"loggedIn": true,
				"userID":   userID,
				"username": username,
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"loggedIn": false,
			})
		}
	})

	// Notifications API
	http.HandleFunc("/api/notifications", handlers.NotificationsAPIHandler)

	// Profile and notifications
	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
			utils.FileService("profile.html", w, nil)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})
	http.HandleFunc("/profile.html", func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
			utils.FileService("profile.html", w, nil)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
			utils.FileService("notifications.html", w, nil)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})
	http.HandleFunc("/notifications.html", func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
			utils.FileService("notifications.html", w, nil)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})

	// Static pages
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("about.html", w, nil)
	})
	http.HandleFunc("/about.html", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("about.html", w, nil)
	})

	http.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("Terms&Conditions.html", w, nil)
	})
	http.HandleFunc("/terms.html", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("Terms&Conditions.html", w, nil)
	})

	// Error pages
	http.HandleFunc("/404", handlers.NotFoundHandler)
	http.HandleFunc("/500", handlers.InternalServerErrorHandler)

	// Static files (CSS, images, JavaScript)
	fs := http.FileServer(http.Dir("frontend/"))
	http.Handle("/frontend/", http.StripPrefix("/frontend/", fs))

	// Homepage
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle 404 for any path that's not exactly "/"
		if r.URL.Path != "/" {
			handlers.NotFoundHandler(w, r)
			return
		}

		// Check if user is logged in
		if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
			utils.FileService("index-signed.html", w, nil)
		} else {
			utils.FileService("index-unsigned.html", w, nil)
		}
	})

	// Start server
	http.ListenAndServe(":8080", nil)
}

// Initialize database and create tables
func initializeDatabase() {
	db := database.CreateTable()
	defer db.Close()

	// Read and execute SQL schema
	sqlContent, err := ioutil.ReadFile("internals/database/table.sql")
	if err != nil {
		fmt.Printf("Warning: Could not read table.sql: %v\n", err)
		return
	}

	// Execute SQL commands
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		fmt.Printf("Warning: Error executing SQL schema: %v\n", err)
	}

	// Insert default categories if they don't exist
	insertDefaultCategories(db)
}

// Insert default categories
func insertDefaultCategories(db *sql.DB) {
	categories := []string{
		"Succulents",
		"Tropical Plants",
		"Herb Garden",
		"Indoor Plants",
		"Plant Care Tips",
		"Plant Diseases",
		"Propagation",
		"Flowering Plants",
	}

	for _, category := range categories {
		var exists int
		db.QueryRow("SELECT COUNT(*) FROM Categories WHERE name = ?", category).Scan(&exists)
		if exists == 0 {
			database.Insert(db, "Categories", "(name)", category)
		}
	}
}

// Logout handler
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get the session cookie
	cookie, err := r.Cookie("session")
	if err == nil {
		// Delete session from database
		db := database.CreateTable()
		defer db.Close()
		db.Exec("DELETE FROM Sessions WHERE cookie_value = ?", cookie.Value)
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
