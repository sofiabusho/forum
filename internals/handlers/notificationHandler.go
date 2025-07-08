package handlers

import (
	"database/sql"
	"encoding/json"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strconv"
)

// NotificationsAPIHandler returns real user notifications from database
func NotificationsAPIHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get unread notifications
	unreadNotifications := getNotifications(db, userID, false)

	// Get read notifications (limit to last 10)
	readNotifications := getNotifications(db, userID, true)

	response := database.NotificationResponse{
		Unread: unreadNotifications,
		Read:   readNotifications,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MarkNotificationReadHandler marks a notification as read
func MarkNotificationReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	notificationIDStr := r.FormValue("notification_id")
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	// Verify notification belongs to user and mark as read
	_, err = db.Exec("UPDATE Notifications SET is_read = TRUE WHERE notification_id = ? AND user_id = ?", notificationID, userID)
	if err != nil {
		http.Error(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// MarkAllNotificationsReadHandler marks all notifications as read for a user
func MarkAllNotificationsReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// Mark all notifications as read for this user
	_, err = db.Exec("UPDATE Notifications SET is_read = TRUE WHERE user_id = ? AND is_read = FALSE", userID)
	if err != nil {
		http.Error(w, "Failed to mark notifications as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// Helper function to get notifications from database
func getNotifications(db *sql.DB, userID int, isRead bool) []database.Notification {
	var notifications []database.Notification

	query := `
		SELECT notification_id, user_id, type, title, message, 
		       related_post_id, related_comment_id, related_user_id, 
		       is_read, creation_date
		FROM Notifications 
		WHERE user_id = ? AND is_read = ?
		ORDER BY creation_date DESC
		LIMIT 20`

	rows, err := db.Query(query, userID, isRead)
	if err != nil {
		return notifications
	}
	defer rows.Close()

	for rows.Next() {
		var n database.Notification
		err := rows.Scan(
			&n.NotificationID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Message,
			&n.RelatedPostID,
			&n.RelatedCommentID,
			&n.RelatedUserID,
			&n.IsRead,
			&n.CreationDate,
		)
		if err != nil {
			continue
		}

		// Format time ago
		n.TimeAgo = formatTimeAgo(n.CreationDate)
		notifications = append(notifications, n)
	}

	return notifications
}

// CreateNotification creates a new notification for a user
func CreateNotification(userID int, notificationType, title, message string, relatedPostID, relatedCommentID, relatedUserID *int) error {
	db := database.CreateTable()
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO Notifications (user_id, type, title, message, related_post_id, related_comment_id, related_user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, notificationType, title, message, relatedPostID, relatedCommentID, relatedUserID)

	return err
}

// GetUnreadNotificationCount returns the count of unread notifications for a user
func GetUnreadNotificationCount(userID int) int {
	db := database.CreateTable()
	defer db.Close()

	var count int
	db.QueryRow("SELECT COUNT(*) FROM Notifications WHERE user_id = ? AND is_read = FALSE", userID).Scan(&count)
	return count
}
