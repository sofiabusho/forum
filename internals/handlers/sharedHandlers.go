package handlers

import (
	"database/sql"
	"fmt"
	"forum/internals/database"
	"forum/internals/utils"
	"time"
)

// getUserIDFromSession returns the user ID for a given session cookie
func getUserIDFromSession(cookieValue string) int {
	db := database.CreateTable()
	defer db.Close()

	var userID int
	err := db.QueryRow("SELECT user_id FROM Sessions WHERE cookie_value = ? AND expiration_date > datetime('now')", cookieValue).Scan(&userID)
	if err != nil {
		return 0
	}
	return userID
}

// formatTimeAgo formats a time.Time into a human-readable "time ago" string
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)
	
	if duration.Minutes() < 1 {
		return "just now"
	} else if duration.Hours() < 1 {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration.Hours() < 24 {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	} else if duration.Hours() < 24*7 {
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	} else if duration.Hours() < 24*30 {
		return fmt.Sprintf("%d weeks ago", int(duration.Hours()/(24*7)))
	} else {
		return fmt.Sprintf("%d months ago", int(duration.Hours()/(24*30)))
	}
}

// truncateText shortens text to a maximum length and adds "..." if truncated
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

// getPostTags returns the category tags for a given post
func getPostTags(db *sql.DB, postID int) []string {
	rows, err := db.Query(`
		SELECT c.name FROM Categories c
		JOIN PostCategories pc ON c.category_id = pc.category_id
		WHERE pc.post_id = ?`, postID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		rows.Scan(&tag)
		tags = append(tags, tag)
	}
	return tags
}

// getPostViews returns the view count for a post (placeholder implementation)
func getPostViews(db *sql.DB, postID int) int {
	// Placeholder - you can implement view tracking later
	// For now, return a random-ish number based on post ID
	return (postID * 7) % 100
}

// getUsernameFromSession returns the username for a given session cookie
func GetUsernameFromSession(cookieValue string) string {
	return utils.GetUsernameFromSession(cookieValue)
}

// isValidSession checks if a session cookie is valid
func isValidSession(cookieValue string) bool {
	return utils.IsValidSession(cookieValue)
}

// checkAuthenticationRequired checks if user is logged in and returns userID
func checkAuthenticationRequired(cookieValue string) (int, bool) {
	if !utils.IsValidSession(cookieValue) {
		return 0, false
	}
	userID := getUserIDFromSession(cookieValue)
	return userID, userID > 0
}