package handlers

import (
	"encoding/json"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CreateCommentHandler handles comment creation
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
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

	userID := utgetUserIDFromSession(cookie.Value)
	if userID == 0 {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Parse form data
	postIDStr := r.FormValue("post_id")
	content := strings.TrimSpace(r.FormValue("content"))

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	if content == "" {
		http.Error(w, "Comment content cannot be empty", http.StatusBadRequest)
		return
	}

	// Insert comment into database
	db := database.CreateTable()
	defer db.Close()

	_, err = db.Exec("INSERT INTO Comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
	if err != nil {
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// CommentsAPIHandler returns comments for a specific post
func CommentsAPIHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get current user ID if logged in
	var currentUserID int
	if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
		currentUserID = getUserIDFromSession(cookie.Value)
	}

	db := database.CreateTable()
	defer db.Close()

	query := `
		SELECT c.comment_id, c.post_id, c.content, c.creation_date, u.username
		FROM Comments c
		JOIN Users u ON c.user_id = u.user_id
		WHERE c.post_id = ?
		ORDER BY c.creation_date ASC`

	rows, err := db.Query(query, postID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []database.CommentResponse
	for rows.Next() {
		var c database.CommentResponse
		var creationDate time.Time

		err := rows.Scan(&c.ID, &c.PostID, &c.Content, &creationDate, &c.Author)
		if err != nil {
			continue
		}

		c.TimeAgo = formatTimeAgo(creationDate)

		// Get like/dislike counts and user's vote
		likeStats := getCommentLikeStats(db, c.ID, currentUserID)
		c.LikeCount = likeStats.LikeCount
		c.DislikeCount = likeStats.DislikeCount
		c.UserVote = likeStats.UserVote

		comments = append(comments, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// DeleteCommentHandler handles comment deletion (for comment author or admin)
func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
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

	userID := getUserIDFromSession(cookie.Value)
	if userID == 0 {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	commentIDStr := r.FormValue("comment_id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	// Check if user owns this comment
	var commentUserID int
	err = db.QueryRow("SELECT user_id FROM Comments WHERE comment_id = ?", commentID).Scan(&commentUserID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if commentUserID != userID {
		http.Error(w, "Unauthorized to delete this comment", http.StatusForbidden)
		return
	}

	// Delete comment likes first (foreign key constraint)
	db.Exec("DELETE FROM CommentLikes WHERE comment_id = ?", commentID)

	// Delete the comment
	_, err = db.Exec("DELETE FROM Comments WHERE comment_id = ?", commentID)
	if err != nil {
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
