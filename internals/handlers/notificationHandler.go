package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

	// Check for pagination parameters
	page := getIntParam(r, "page", 1)
	limit := getIntParam(r, "limit", 20)
	if limit > 50 {
		limit = 50 // Prevent too large requests
	}

	db := database.CreateTable()
	defer db.Close()

	// Get unread notifications
	unreadNotifications := getNotificationsWithPagination(db, userID, false, page, limit)

	// Get read notifications
	readNotifications := getNotificationsWithPagination(db, userID, true, page, limit)

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

	// Verify notification belongs to user before marking as read
	var existingUserID int
	var isRead bool
	err = db.QueryRow("SELECT user_id, is_read FROM Notifications WHERE notification_id = ?", notificationID).Scan(&existingUserID, &isRead)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Notification not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if existingUserID != userID {
		http.Error(w, "Unauthorized to modify this notification", http.StatusForbidden)
		return
	}

	if isRead {
		// Already read, just return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		return
	}

	// Mark as read
	_, err = db.Exec("UPDATE Notifications SET is_read = TRUE WHERE notification_id = ?", notificationID)
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

	// Get count of unread notifications before marking
	var unreadCount int
	db.QueryRow("SELECT COUNT(*) FROM Notifications WHERE user_id = ? AND is_read = FALSE", userID).Scan(&unreadCount)

	if unreadCount == 0 {
		// No unread notifications
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"marked":  0,
		})
		return
	}

	// Mark all notifications as read for this user
	result, err := db.Exec("UPDATE Notifications SET is_read = TRUE WHERE user_id = ? AND is_read = FALSE", userID)
	if err != nil {
		http.Error(w, "Failed to mark notifications as read", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"marked":  rowsAffected,
	})
}

// CreateNotification creates a new notification for a user
func CreateNotification(userID int, notificationType, title, message string, relatedPostID, relatedCommentID, relatedUserID *int) error {
	// Do not create notification for user's own actions
	if userID == 0 || (relatedUserID != nil && *relatedUserID == userID) {
		return nil
	}

	db := database.CreateTable()
	defer db.Close()

	//Check if notification already exists
	if skipDuplicateNotification(db, userID, notificationType, relatedPostID, relatedCommentID, relatedUserID) {

	}

	_, err := db.Exec(`
		INSERT INTO Notifications (user_id, type, title, message, related_post_id, related_comment_id, related_user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, notificationType, title, message, relatedPostID, relatedCommentID, relatedUserID)

	return err
}

func CreateCommentNotification(postID int, commenterID int, commenterUsername string, postTitle string) {
	db := database.CreateTable()
	defer db.Close()

	//Get the author of the post
	var postAuthorID int
	err := db.QueryRow("SELECT user_id FROM Posts WHERE post_id = ?", postID).Scan(&postAuthorID)
	if err != nil || postAuthorID == commenterID {
		return // Do not notify the commenter about their own comment
	}

	title := "New Comment!"
	message := fmt.Sprintf("%s commented on your post '%s'", commenterUsername, truncateText(postTitle, 50))

	CreateNotification(postAuthorID, "comment", title, message, &postID, nil, &commenterID)
}

func CreateLikeNotification(postID int, likerID int, likerUsername string, postTitle string) {
	db := database.CreateTable()
	defer db.Close()

	// Get the author of the post
	var postAuthorID int
	err := db.QueryRow("SELECT user_id FROM Posts WHERE post_id = ?", postID).Scan(&postAuthorID)
	if err != nil || postAuthorID == likerID {
		return // Do not notify the liker about their own like
	}

	title := "New Like!"
	message := fmt.Sprintf("%s liked your post '%s'", likerUsername, truncateText(postTitle, 50))

	CreateNotification(postAuthorID, "liket", title, message, &postID, nil, &likerID)
}

// Helper function to get notifications from database
func getNotificationsWithPagination(db *sql.DB, userID int, isRead bool, page, limit int) []database.Notification {
	var notifications []database.Notification

	offset := (page - 1) * limit

	query := `
		SELECT notification_id, user_id, type, title, message, 
		       related_post_id, related_comment_id, related_user_id, 
		       is_read, creation_date
		FROM Notifications 
		WHERE user_id = ? AND is_read = ?
		ORDER BY creation_date DESC
		LIMIT 20`

	rows, err := db.Query(query, userID, isRead, limit, offset)
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

func skipDuplicateNotification(db *sql.DB, userID int, notificationType string, relatedPostID, relatedCommentID, relatedUserID *int) bool {
	// Check if a similar notification already exists
	query := `
	SELECT COUNT(*) FROM Notifications
	WHERE user_id = ? AND type = ? AND related_post_id = ? AND related_comment_id = ? AND related_user_id = ?
	AND creation_date > datetime('now', '-1 hour')`

	var count int
	err := db.QueryRow(query, userID, notificationType, relatedPostID, relatedCommentID, relatedUserID).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0

}

// GetUnreadNotificationCount returns the count of unread notifications for a user
func GetUnreadNotificationCount(userID int) int {
	db := database.CreateTable()
	defer db.Close()

	var count int
	db.QueryRow("SELECT COUNT(*) FROM Notifications WHERE user_id = ? AND is_read = FALSE", userID).Scan(&count)
	return count
}

func NotificationCountHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"count": 0})
		return
	}

	userID := utils.GetUserIDFromSession(cookie.Value)
	count := GetUnreadNotificationCount(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

// Delete old notifications older than 30 days
func DeleteOldNotifications() {
	db := database.CreateTable()
	defer db.Close()

	_, err := db.Exec(`
		DELETE FROM Notifications
		WHERE is_read = TRUE AND creation_date < datetime('now', '-30 days')
		`)
		if err != nil {
			fmt.Printf("Error deleting old notifications: %v\n", err)
	}
}

// Helper function to get integer parameter from request
func getIntParam(r *http.Request, param string, defaultValue int) int {
	valueStr := r.URL.Query().Get(param)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func SystemNotification(userIDs []int, title, message string) error {
	db := database.CreateTable()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	smtm, err := tx.Prepare(`
		INSERT INTO Notifications (user_id, type, title, message)
		VALUES (?, 'system;, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer smtm.Close()

	for _, userID := range userIDs {
		_, err = smtm.Exec(userID, title, message)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

