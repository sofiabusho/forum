package handlers

import (
	"encoding/json"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"time"
)

// FilteredPostsHandler handles filtering posts by user's created posts and liked posts
func FilteredPostsHandler(w http.ResponseWriter, r *http.Request) {
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

	filter := r.URL.Query().Get("filter")
	db := database.CreateTable()
	defer db.Close()

	var query string
	var args []interface{}

	switch filter {
	case "my-posts":
		// Get posts created by the user
		query = `
			SELECT p.post_id, p.title, p.content, u.username, p.creation_date,
			       (SELECT COUNT(*) FROM Comments WHERE post_id = p.post_id) as comment_count,
			       (SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = 1) as like_count
			FROM Posts p 
			JOIN Users u ON p.user_id = u.user_id
			WHERE p.user_id = ?
			ORDER BY p.creation_date DESC`
		args = append(args, userID)

	case "my-likes":
		// Get posts liked by the user
		query = `
			SELECT p.post_id, p.title, p.content, u.username, p.creation_date,
			       (SELECT COUNT(*) FROM Comments WHERE post_id = p.post_id) as comment_count,
			       (SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = 1) as like_count
			FROM Posts p 
			JOIN Users u ON p.user_id = u.user_id
			JOIN LikesDislikes ld ON p.post_id = ld.post_id
			WHERE ld.user_id = ? AND ld.vote = 1
			ORDER BY p.creation_date DESC`
		args = append(args, userID)

	default:
		http.Error(w, "Invalid filter", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []database.PostResponse
	for rows.Next() {
		var p database.PostResponse
		var creationDate time.Time
		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &creationDate, &p.Comments, &p.Likes)
		if err != nil {
			continue
		}

		p.TimeAgo = utils.FormatTimeAgo(creationDate)
		p.Excerpt = utils.TruncateText(p.Content, 150)
		p.Tags = getPostTags(db, p.ID)

		posts = append(posts, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
