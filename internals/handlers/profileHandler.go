package handlers

import (
	"database/sql"
	"encoding/json"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strings"
	"time"
)

// ProfileHandler serves the profile page and handles profile updates
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		// Serve profile page
		utils.FileService("profile.html", w, nil)
	} else if r.Method == "POST" {
		// Handle profile updates
		updateProfile(w, r)
	}
}

// ProfileAPIHandler returns user profile data as JSON
func ProfileAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := utils.GetUserIDFromSession(cookie.Value)
	if userID == 0 {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	// Get user profile data
	profile := getUserProfile(db, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// ActivityAPIHandler returns user activity data
func ActivityAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := utils.GetUserIDFromSession(cookie.Value)
	if userID == 0 {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	// Get user activity
	activity := getUserActivity(db, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activity)
}

// updateProfile handles profile update requests
func updateProfile(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromSession(getCookieValue(r))
	if userID == 0 {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Parse form data
	newUsername := strings.TrimSpace(r.FormValue("username"))
	newEmail := strings.TrimSpace(r.FormValue("email"))
	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")

	db := database.CreateTable()
	defer db.Close()

	// Validate current password if changing password
	if newPassword != "" {
		if currentPassword == "" {
			http.Error(w, "Current password required", http.StatusBadRequest)
			return
		}

		var storedHash string
		err := db.QueryRow("SELECT password_hash FROM Users WHERE user_id = ?", userID).Scan(&storedHash)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Verify current password (you'll need to implement password verification)
		if !verifyPassword(currentPassword, storedHash) {
			http.Error(w, "Incorrect current password", http.StatusBadRequest)
			return
		}

		// Hash new password
		hashedPassword := hashPassword(newPassword)
		_, err = db.Exec("UPDATE Users SET password_hash = ? WHERE user_id = ?", hashedPassword, userID)
		if err != nil {
			http.Error(w, "Failed to update password", http.StatusInternalServerError)
			return
		}
	}

	// Update username and email
	if newUsername != "" || newEmail != "" {
		// Check if username/email already exists
		if newUsername != "" {
			var exists int
			db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = ? AND user_id != ?", newUsername, userID).Scan(&exists)
			if exists > 0 {
				http.Error(w, "Username already taken", http.StatusBadRequest)
				return
			}
		}

		if newEmail != "" {
			if !utils.IsValidEmail(newEmail) {
				http.Error(w, "Invalid email format", http.StatusBadRequest)
				return
			}

			var exists int
			db.QueryRow("SELECT COUNT(*) FROM Users WHERE email = ? AND user_id != ?", newEmail, userID).Scan(&exists)
			if exists > 0 {
				http.Error(w, "Email already taken", http.StatusBadRequest)
				return
			}
		}

		// Update user info
		query := "UPDATE Users SET "
		args := []interface{}{}
		updates := []string{}

		if newUsername != "" {
			updates = append(updates, "username = ?")
			args = append(args, newUsername)
		}
		if newEmail != "" {
			updates = append(updates, "email = ?")
			args = append(args, newEmail)
		}

		query += strings.Join(updates, ", ") + " WHERE user_id = ?"
		args = append(args, userID)

		_, err := db.Exec(query, args...)
		if err != nil {
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Helper functions
func getUserProfile(db *sql.DB, userID int) database.UserProfile {
	var profile database.UserProfile
	var registrationDate time.Time

	// Get basic user info
	err := db.QueryRow(`
		SELECT user_id, username, email, registration_date 
		FROM Users 
		WHERE user_id = ?
	`, userID).Scan(&profile.UserID, &profile.Username, &profile.Email, &registrationDate)

	if err != nil {
		return profile
	}

	profile.JoinDate = registrationDate.Format("January 2, 2006")

	// Get post count
	db.QueryRow("SELECT COUNT(*) FROM Posts WHERE user_id = ?", userID).Scan(&profile.PostCount)

	// Get comment count
	db.QueryRow("SELECT COUNT(*) FROM Comments WHERE user_id = ?", userID).Scan(&profile.CommentCount)

	// Get likes given
	db.QueryRow("SELECT COUNT(*) FROM LikesDislikes WHERE user_id = ? AND vote = 1", userID).Scan(&profile.LikesGiven)

	// Get likes received on user's posts
	db.QueryRow(`
		SELECT COUNT(*) 
		FROM LikesDislikes ld 
		JOIN Posts p ON ld.post_id = p.post_id 
		WHERE p.user_id = ? AND ld.vote = 1
	`, userID).Scan(&profile.LikesReceived)

	return profile
}

func getUserActivity(db *sql.DB, userID int) database.UserActivity {
	var activity database.UserActivity

	// Get recent posts (last 5)
	postRows, err := db.Query(`
		SELECT p.post_id, p.title, p.content, p.creation_date,
		       (SELECT COUNT(*) FROM Comments WHERE post_id = p.post_id) as comment_count,
		       (SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = 1) as like_count
		FROM Posts p
		WHERE p.user_id = ?
		ORDER BY p.creation_date DESC
		LIMIT 5
	`, userID)

	if err == nil {
		defer postRows.Close()
		for postRows.Next() {
			var post database.PostResponse
			var creationDate time.Time
			postRows.Scan(&post.ID, &post.Title, &post.Content, &creationDate, &post.Comments, &post.Likes)

			post.TimeAgo = formatTimeAgo(creationDate)
			post.Excerpt = truncateText(post.Content, 100)
			post.Tags = getPostTags(db, post.ID)

			activity.RecentPosts = append(activity.RecentPosts, post)
		}
	}

	// Get recent comments (last 5)
	commentRows, err := db.Query(`
		SELECT c.comment_id, c.post_id, c.content, c.creation_date, p.title
		FROM Comments c
		JOIN Posts p ON c.post_id = p.post_id
		WHERE c.user_id = ?
		ORDER BY c.creation_date DESC
		LIMIT 5
	`, userID)

	if err == nil {
		defer commentRows.Close()
		for commentRows.Next() {
			var comment database.CommentActivity
			var creationDate time.Time
			commentRows.Scan(&comment.ID, &comment.PostID, &comment.Content, &creationDate, &comment.PostTitle)

			comment.TimeAgo = formatTimeAgo(creationDate)
			comment.Content = truncateText(comment.Content, 100)

			activity.RecentComments = append(activity.RecentComments, comment)
		}
	}

	return activity
}

func getCookieValue(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// Placeholder functions - implement these based on your password hashing method
func verifyPassword(password, hash string) bool {
	// Implement password verification (bcrypt, etc.)
	// This is a placeholder - replace with actual implementation
	return hashPassword(password) == hash
}

func hashPassword(password string) string {
	// Implement password hashing (bcrypt, etc.)
	// This is a placeholder - replace with actual implementation
	return password // DON'T USE THIS IN PRODUCTION
}
