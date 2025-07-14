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
	wrapHandler("/login", handlers.LoginHandler)
	wrapHandler("/login.html", handlers.LoginHandler)
	wrapHandler("/register", handlers.RegisterHandler)
	wrapHandler("/register.html", handlers.RegisterHandler)
	wrapHandler("/logout", logoutHandler)

	// Post routes
	wrapHandler("/new-post", handlers.CreatePostHandler)
	wrapHandler("/new-post.html", handlers.CreatePostHandler)
	wrapHandler("/api/posts", handlers.PostsAPIHandler)

	// Comment routes
	wrapHandler("/api/comments/create", handlers.CreateCommentHandler)
	wrapHandler("/api/comments", handlers.CommentsAPIHandler)
	wrapHandler("/api/comments/delete", handlers.DeleteCommentHandler)

	// Like/Dislike routes
	wrapHandler("/api/posts/like", handlers.LikePostHandler)
	wrapHandler("/api/comments/like", handlers.LikeCommentHandler)

	// Category routes
	wrapHandler("/api/categories", handlers.CategoriesAPIHandler)
	wrapHandler("/categories", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("categories.html", w, nil)
	})
	wrapHandler("/categories.html", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("categories.html", w, nil)
	})

	// Forgot Password Page
	wrapHandler("/forgot-password.html", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("forgot-password.html", w, nil)
	})

	// Forgot Password Logic
	wrapHandler("/forgot-password", handlers.ForgotPasswordHandler)

	wrapHandler("/reset-password", handlers.ResetPasswordHandler)

	wrapHandler("/add-newpassword.html", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("add-newpassword.html", w, nil)
	})

	// Filter routes (for authenticated users)
	wrapHandler("/api/posts/filtered", handlers.FilteredPostsHandler)

	// Auth status check - ONLY ONE REGISTRATION
	wrapHandler("/api/auth/status", func(w http.ResponseWriter, r *http.Request) {
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
	wrapHandler("/api/notifications", handlers.NotificationsAPIHandler)
	wrapHandler("/api/notifications/mark-read", handlers.MarkNotificationReadHandler)
	wrapHandler("/api/notifications/mark-all-read", handlers.MarkAllNotificationsReadHandler)

	// Image upload and management
	wrapHandler("/api/upload-image", handlers.ImageUploadHandler)
	wrapHandler("/api/delete-image", handlers.DeleteImageHandler)
	wrapHandler("/api/user-images", handlers.GetUserImagesHandler)

	// Profile and notifications
	wrapHandler("/profile", func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
			utils.FileService("profile.html", w, nil)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})
	wrapHandler("/profile.html", func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
			utils.FileService("profile.html", w, nil)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})

	wrapHandler("/notifications", func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
			utils.FileService("notifications.html", w, nil)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})
	wrapHandler("/notifications.html", func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
			utils.FileService("notifications.html", w, nil)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})

	wrapHandler("/api/notifications/count", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || !utils.IsValidSession(cookie.Value) {
			json.NewEncoder(w).Encode(map[string]int{"count": 0})
			return
		}

		userID := utils.GetUserIDFromSession(cookie.Value)
		count := handlers.GetUnreadNotificationCount(userID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"count": count})
	})

	// Static pages
	wrapHandler("/about", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("about.html", w, nil)
	})
	wrapHandler("/about.html", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("about.html", w, nil)
	})

	wrapHandler("/terms", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("Terms&Conditions.html", w, nil)
	})
	wrapHandler("/terms.html", func(w http.ResponseWriter, r *http.Request) {
		utils.FileService("Terms&Conditions.html", w, nil)
	})

	// Error pages - ADD THESE NEW ROUTES
	wrapHandler("/404", handlers.NotFoundHandler)
	wrapHandler("/500", handlers.InternalServerErrorHandler)

	// Static files (CSS, images, JavaScript)
	fs := http.FileServer(http.Dir("frontend/"))
	http.Handle("/frontend/", http.StripPrefix("/frontend/", fs))

	// Homepage - IMPROVED 404 HANDLING
	wrapHandler("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle 404 for any path that's not exactly "/" or registered routes
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

// KEEP YOUR EXISTING FUNCTIONS - NO CHANGES NEEDED
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
