package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
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

	if unreadNotifications == nil {
		unreadNotifications = make([]database.Notification, 0)
	}
	if readNotifications == nil {
		readNotifications = make([]database.Notification, 0)
	}

	response := database.NotificationResponse{Unread: unreadNotifications, Read: readNotifications}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
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

	if !isRead {
		if _, err := db.Exec("UPDATE Notifications SET is_read = 1 WHERE notification_id = ?", notificationID); err != nil {
			http.Error(w, "Failed to mark read", http.StatusInternalServerError)
			return
		}
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

	if _, err := db.Exec("UPDATE Notifications SET is_read = 1 WHERE user_id = ? AND (is_read = 0 OR is_read IS NULL)", userID); err != nil {
		http.Error(w, "Failed to mark all as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func setBusyTimeout(db *sql.DB) {
	_, _ = db.Exec("PRAGMA busy_timeout=5000;")
}

func execWithRetry(db *sql.DB, query string, args ...any) (sql.Result, error) {
	var lastErr error
	for i := 0; i < 5; i++ {
		res, err := db.Exec(query, args...)
		if err == nil {
			return res, nil
		}
		msg := err.Error()
		if strings.Contains(msg, "database is locked") || strings.Contains(msg, "busy") {
			// exponential-ish backoff: 100ms, 200ms, 400ms, 800ms, 1200ms
			time.Sleep(time.Duration(100*(1<<min(i, 3))) * time.Millisecond)
			lastErr = err
			continue
		}
		return nil, err
	}
	if lastErr == nil {
		lastErr = errors.New("execWithRetry: unknown error")
	}
	return nil, lastErr
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CreateNotification creates a new notification for a user
func CreateNotification(userID int, notificationType, title, message string, relatedPostID, relatedCommentID, relatedUserID *int) error {
	// Do not create notification for user's own actions
	if userID == 0 || (relatedUserID != nil && *relatedUserID == userID) {
		return nil
	}

	db := database.CreateTable()
	defer db.Close()

	setBusyTimeout(db)

	//Check if notification already exists
	if skipDuplicateNotification(db, userID, notificationType, relatedPostID, relatedCommentID, relatedUserID) {
		return nil
	}

	_, err := execWithRetry(db, `
		INSERT INTO Notifications (user_id, type, title, message, related_post_id, related_comment_id, related_user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, notificationType, title, message, relatedPostID, relatedCommentID, relatedUserID)
	return err
}

func CreateCommentNotification(postID int, commentID int, commenterID int, commenterUsername string, postTitle string) {
	db := database.CreateTable()
	defer db.Close()

	//Get the author of the post
	var postAuthorID int
	if err := db.QueryRow("SELECT user_id FROM Posts WHERE post_id = ?", postID).Scan(&postAuthorID); err != nil {
		return
	}
	if postAuthorID == commenterID {
		return
	}

	title := "New Comment!"
	message := fmt.Sprintf("%s commented on your post '%s'", commenterUsername, utils.TruncateText(postTitle, 50))

	_ = CreateNotification(postAuthorID, "comment", title, message, &postID, &commentID, &commenterID)
}

func CreateLikeNotification(postID int, likerID int, likerUsername string, postTitle string) {
	db := database.CreateTable()
	defer db.Close()

	// Get the author of the post
	var postAuthorID int
	if err := db.QueryRow("SELECT user_id FROM Posts WHERE post_id = ?", postID).Scan(&postAuthorID); err != nil {
		return
	}
	if postAuthorID == likerID {
		return
	}

	title := "New Like!"
	message := fmt.Sprintf("%s liked your post '%s'", likerUsername, utils.TruncateText(postTitle, 50))

	_ = CreateNotification(postAuthorID, "like", title, message, &postID, nil, &likerID)
}

// CreateDislikeNotification creates notification for post dislikes
func CreateDislikeNotification(postID int, dislikerID int, dislikerUsername string, postTitle string) {
	db := database.CreateTable()
	defer db.Close()

	// Get the author of the post
	var postAuthorID int
	if err := db.QueryRow("SELECT user_id FROM Posts WHERE post_id = ?", postID).Scan(&postAuthorID); err != nil {
		return
	}
	if postAuthorID == dislikerID {
		return
	}

	title := "Someone disagreed with your post"
	message := fmt.Sprintf("%s disliked your post '%s'", dislikerUsername, utils.TruncateText(postTitle, 50))

	_ = CreateNotification(postAuthorID, "dislike", title, message, &postID, nil, &dislikerID)
}

func CreateCommentLikeNotification(commentID, likerID int, likerUsername, postTitle string, postID int) {
	db := database.CreateTable()
	defer db.Close()

	var commentAuthorID int
	if err := db.QueryRow("SELECT user_id FROM Comments WHERE comment_id = ?", commentID).Scan(&commentAuthorID); err != nil {
		return
	}
	if commentAuthorID == likerID {
		return
	}

	title := "Your comment got a like!"
	message := fmt.Sprintf("%s liked your comment on '%s'", likerUsername, utils.TruncateText(postTitle, 50))
	_ = CreateNotification(commentAuthorID, "like", title, message, &postID, &commentID, &likerID)
}

// CreateCommentDislikeNotification creates notification for comment dislikes
func CreateCommentDislikeNotification(commentID, dislikerID int, dislikerUsername, postTitle string, postID int) {
	db := database.CreateTable()
	defer db.Close()

	var commentAuthorID int
	if err := db.QueryRow("SELECT user_id FROM Comments WHERE comment_id = ?", commentID).Scan(&commentAuthorID); err != nil {
		return
	}
	if commentAuthorID == dislikerID {
		return
	}

	title := "Someone disagreed with your comment"
	message := fmt.Sprintf("%s disliked your comment on '%s'", dislikerUsername, utils.TruncateText(postTitle, 50))
	_ = CreateNotification(commentAuthorID, "dislike", title, message, &postID, &commentID, &dislikerID)
}

// CreateFollowupCommentNotifications notifies ALL previous commenters on a post (except author & current commenter)
// This implements "watching" functionality - once you comment on a post, you follow that discussion
func CreateFollowupCommentNotifications(postID, commentID, commenterID int, commenterUsername, postTitle string) {
	db := database.CreateTable()
	defer db.Close()

	// 1) Find the post author (they don't get notified here, they get separate notification)
	var postAuthorID int
	if err := db.QueryRow("SELECT user_id FROM Posts WHERE post_id = ?", postID).Scan(&postAuthorID); err != nil {
		fmt.Printf("[FollowupNotify] post=%d: cannot load postAuthorID: %v\n", postID, err)
		return
	}

	// 2) Find DISTINCT previous commenters on this post:
	//    - Same post
	//    - NOT the current commenter
	//    - NOT the post author
	//    DISTINCT ensures 1 notification per user even if they have multiple comments
	rows, err := db.Query(`
		SELECT DISTINCT c.user_id
		FROM Comments AS c
		WHERE c.post_id = ?
		  AND c.user_id != ?
		  AND c.user_id != ?
	`, postID, commenterID, postAuthorID)
	if err != nil {
		fmt.Printf("[FollowupNotify] post=%d: query watchers failed: %v\n", postID, err)
		return
	}
	defer rows.Close()

	title := "New activity on a post you commented"
	msg := fmt.Sprintf("%s also commented on '%s'", commenterUsername, utils.TruncateText(postTitle, 50))
	var count int

	for rows.Next() {
		var watcherID int
		if err := rows.Scan(&watcherID); err != nil {
			fmt.Printf("[FollowupNotify] post=%d: scan watcher err: %v\n", postID, err)
			continue
		}

		// 3) Send notification to each watcher
		//    related_post_id = postID
		//    related_comment_id = the NEW comment (for anti-spam uniqueness)
		//    related_user_id = commenterID (who made the new comment)
		err := CreateNotification(
			watcherID,
			"comment",
			title,
			msg,
			&postID,
			&commentID,
			&commenterID,
		)
		if err != nil {
			fmt.Printf("[FollowupNotify] post=%d newComment=%d -> watcher=%d ERROR sending: %v\n",
				postID, commentID, watcherID, err)
			continue
		}
		count++
	}

	if count == 0 {
	}
}

// CreateDirectReplyNotification notifies the author of a parent comment when someone replies directly to them
func CreateDirectReplyNotification(parentCommentID, newCommentID, replierID int, replierUsername, postTitle string, postID int) {
	db := database.CreateTable()
	defer db.Close()

	var parentAuthorID int
	if err := db.QueryRow("SELECT user_id FROM Comments WHERE comment_id = ?", parentCommentID).Scan(&parentAuthorID); err != nil {
		return
	}

	// Don't notify yourself
	if parentAuthorID == replierID {
		return
	}

	title := "New reply to your comment"
	message := fmt.Sprintf("%s replied to your comment on '%s'", replierUsername, utils.TruncateText(postTitle, 50))
	_ = CreateNotification(parentAuthorID, "comment", title, message, &postID, &newCommentID, &replierID)
}

// Helper function to get notifications from database
func getNotificationsWithPagination(db *sql.DB, userID int, isRead bool, page, limit int) []database.Notification {
	list := make([]database.Notification, 0, limit)

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	var rows *sql.Rows
	var err error

	if isRead {
		rows, err = db.Query(`
            SELECT notification_id, user_id, type, title, message,
                   related_post_id, related_comment_id, related_user_id,
                   COALESCE(is_read, 0) AS is_read, creation_date
            FROM Notifications
            WHERE user_id = ? AND is_read = 1
            ORDER BY creation_date DESC
            LIMIT ? OFFSET ?`,
			userID, limit, offset)
	} else {
		rows, err = db.Query(`
            SELECT notification_id, user_id, type, title, message,
                   related_post_id, related_comment_id, related_user_id,
                   COALESCE(is_read, 0) AS is_read, creation_date
            FROM Notifications
            WHERE user_id = ? AND (is_read = 0 OR is_read IS NULL)
            ORDER BY creation_date DESC
            LIMIT ? OFFSET ?`,
			userID, limit, offset)
	}

	if err != nil {
		return list
	}
	defer rows.Close()

	for rows.Next() {
		var n database.Notification
		if err := rows.Scan(
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
		); err != nil {
			continue
		}
		n.TimeAgo = utils.FormatTimeAgo(n.CreationDate)
		list = append(list, n)
	}
	return list
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
	err := db.QueryRow("SELECT COUNT(*) FROM Notifications WHERE user_id = ? AND (is_read = 0 OR is_read = false OR is_read IS NULL)", userID).Scan(&count)
	if err != nil {
		fmt.Printf("Error getting notification count for user %d: %v\n", userID, err)
		return 0
	}
	return count
}

func NotificationCountHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"count": 0})
		return
	}

	userID := utils.GetUserIDFromSession(cookie.Value)
	count := GetUnreadNotificationCount(userID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{"count": count})
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
		VALUES (?, 'system', ?, ?)
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
