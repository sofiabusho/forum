package handlers

import (
	"database/sql"
	"encoding/json"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strconv"
)

// LikePostHandler handles liking/disliking posts
func LikePostHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse Rquest
	postIDStr := r.FormValue("post_id")
	voteStr := r.FormValue("vote") // "like" is 1 and "dislike is -1"

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	vote, err := strconv.Atoi(voteStr)
	if err != nil || (vote != 1 && vote != -1) {
		http.Error(w, "Invalid vote", http.StatusBadRequest)
		return

	}

	db := database.CreateTable()
	defer db.Close()

	// Check if user already voted on this post
	var existingVote int
	err = db.QueryRow("SELECT vote FROM PostVotes WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&existingVote)

	var isNewLike bool = false

	if err != nil {
		// User already voted
		if existingVote == vote {
			// Same vote - remove it (toggle off)
			db.Exec("DELETE FROM LikesDislikes WHERE post_id = ? AND user_id = ?", postID, userID)
		} else {
			// Different vote - update it
			db.Exec("UPDATE LikesDislikes SET vote = ? WHERE post_id = ? AND user_id = ?", vote, postID, userID)
			if vote == 1 {
				isNewLike = true
			}
		}
	} else {
		// No existing vote - create new one
		db.Exec("INSERT INTO LikesDislikes (post_id, user_id, vote) VALUES (?, ?, ?)", postID, userID, vote)
		if vote == 1 {
			isNewLike = true
		}
	}

	// Create notification for new likes only
	if isNewLike {
		var postAuthorID int
		var postTitle string
		err := db.QueryRow("SELECT user_id, title FROM Posts WHERE post_id = ?", postID).Scan(&postAuthorID, &postTitle)

		if err == nil && postAuthorID != userID { // Don't notify yourself
			likerUsername := utils.GetUsernameFromSession(cookie.Value)
			CreateLikeNotification(postID, userID, likerUsername, postTitle)
		}
	}

	// Get updated counts and user's current vote
	response := getLikeStats(db, postID, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LikeCommentHandler handles liking/disliking comments
func LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse request
	commentIDStr := r.FormValue("comment_id")
	voteStr := r.FormValue("vote")

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	vote, err := strconv.Atoi(voteStr)
	if err != nil || (vote != 1 && vote != -1) {
		http.Error(w, "Invalid vote", http.StatusBadRequest)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	// Check if user already voted on this comment
	var existingVote int
	err = db.QueryRow("SELECT vote FROM CommentLikes WHERE comment_id = ? AND user_id = ?", commentID, userID).Scan(&existingVote)

	if err == nil {
		// User already voted
		if existingVote == vote {
			// Same vote - remove it
			db.Exec("DELETE FROM CommentLikes WHERE comment_id = ? AND user_id = ?", commentID, userID)
		} else {
			// Different vote - update it
			db.Exec("UPDATE CommentLikes SET vote = ? WHERE comment_id = ? AND user_id = ?", vote, commentID, userID)
		}
	} else {
		// No existing vote - create new one
		db.Exec("INSERT INTO CommentLikes (comment_id, user_id, vote) VALUES (?, ?, ?)", commentID, userID, vote)
	}

	// Get updated counts
	response := getCommentLikeStats(db, commentID, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions
func getLikeStats(db *sql.DB, postID, userID int) database.LikeResponse {
	var likeCount, dislikeCount int

	// Get like count
	db.QueryRow("SELECT COUNT(*) FROM LikesDislikes WHERE post_id = ? AND vote = 1", postID).Scan(&likeCount)

	// Get dislike count
	db.QueryRow("SELECT COUNT(*) FROM LikesDislikes WHERE post_id = ? AND vote = -1", postID).Scan(&dislikeCount)

	// Get user's current vote
	var userVote int
	err := db.QueryRow("SELECT vote FROM LikesDislikes WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&userVote)
	if err != nil {
		userVote = 0 // No vote
	}

	return database.LikeResponse{
		Success:      true,
		LikeCount:    likeCount,
		DislikeCount: dislikeCount,
		UserVote:     userVote,
	}
}

func getCommentLikeStats(db *sql.DB, commentID, userID int) database.LikeResponse {
	var likeCount, dislikeCount int

	// Get like count
	db.QueryRow("SELECT COUNT(*) FROM CommentLikes WHERE comment_id = ? AND vote = 1", commentID).Scan(&likeCount)

	// Get dislike count
	db.QueryRow("SELECT COUNT(*) FROM CommentLikes WHERE comment_id = ? AND vote = -1", commentID).Scan(&dislikeCount)

	// Get user's current vote
	var userVote int
	err := db.QueryRow("SELECT vote FROM CommentLikes WHERE comment_id = ? AND user_id = ?", commentID, userID).Scan(&userVote)
	if err != nil {
		userVote = 0 // No vote
	}

	return database.LikeResponse{
		Success:      true,
		LikeCount:    likeCount,
		DislikeCount: dislikeCount,
		UserVote:     userVote,
	}
}
