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

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		utils.FileService("profile.html", w, nil)
	case http.MethodPost:
		updateProfile(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func ProfileAPIHandler(w http.ResponseWriter, r *http.Request) {
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

	switch r.Method {
	case http.MethodGet:
		db := database.CreateTable()
		defer db.Close()

		profile := getUserProfile(db, userID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromSession(getCookieValue(r))
	if userID == 0 {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	newUsername := strings.TrimSpace(r.FormValue("username"))
	newBio := strings.TrimSpace(r.FormValue("bio"))

	if newUsername != "" {
		var exists int
		_ = db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = ? AND user_id != ?", newUsername, userID).Scan(&exists)
		if exists > 0 {
			http.Error(w, "Username already taken", http.StatusBadRequest)
			return
		}
		if _, err := db.Exec("UPDATE Users SET username = ? WHERE user_id = ?", newUsername, userID); err != nil {
			http.Error(w, "Failed to update username", http.StatusInternalServerError)
			return
		}
	}

	if newBio != "" {
		if _, err := db.Exec("UPDATE Users SET bio = ? WHERE user_id = ?", newBio, userID); err != nil {
			http.Error(w, "Failed to update bio", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"username": newUsername,
		"bio":      newBio,
	})
}

func getUserProfile(db *sql.DB, userID int) database.UserProfile {
	var profile database.UserProfile
	var bio string

	// Date from SQL
	err := db.QueryRow(`
		SELECT 
			u.user_id,
			u.username,
			u.email,
			COALESCE( strftime('%Y-%m-%d', u.registration_date), '' ) AS join_date,
			COALESCE(u.bio, '')
		FROM Users u
		WHERE u.user_id = ?
	`, userID).Scan(&profile.UserID, &profile.Username, &profile.Email, &profile.JoinDate, &bio)

	if err != nil {
		return profile
	}
	profile.Bio = bio

	// Counts
	_ = db.QueryRow(`SELECT COUNT(*) FROM Posts WHERE user_id = ?`, userID).Scan(&profile.PostCount)
	_ = db.QueryRow(`SELECT COUNT(*) FROM Comments WHERE user_id = ?`, userID).Scan(&profile.CommentCount)

	// Likes Given / Received
	_ = db.QueryRow(`SELECT COUNT(*) FROM LikesDislikes WHERE user_id = ? AND vote = 1`, userID).Scan(&profile.LikesGiven)
	_ = db.QueryRow(`
		SELECT COUNT(*)
		FROM LikesDislikes ld 
		JOIN Posts p ON ld.post_id = p.post_id 
		WHERE p.user_id = ? AND ld.vote = 1
	`, userID).Scan(&profile.LikesReceived)

	// Dislikes Given / Received
	_ = db.QueryRow(`SELECT COUNT(*) FROM LikesDislikes WHERE user_id = ? AND vote = -1`, userID).Scan(&profile.DislikesGiven)
	_ = db.QueryRow(`
		SELECT COUNT(*)
		FROM LikesDislikes ld 
		JOIN Posts p ON ld.post_id = p.post_id 
		WHERE p.user_id = ? AND ld.vote = -1
	`, userID).Scan(&profile.DislikesReceived)

	// Profile image
	var thumbnailURL string
	_ = db.QueryRow(`
		SELECT thumbnail_url 
		FROM Images 
		WHERE user_id = ? AND image_type = 'profile'
		ORDER BY upload_date DESC 
		LIMIT 1
	`, userID).Scan(&thumbnailURL)
	profile.ProfileImage = thumbnailURL

	return profile
}

func getCookieValue(r *http.Request) string {
	cookie, err := r.Cookie("session")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// /api/user/posts
func UserPostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		userID := utils.GetUserIDFromSession(getCookieValue(r))
		if userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		db := database.CreateTable()
		defer db.Close()

		rows, err := db.Query(`
			SELECT post_id, title, content, creation_date 
			FROM Posts 
			WHERE user_id = ? 
			ORDER BY creation_date DESC
		`, userID)
		if err != nil {
			http.Error(w, "Database query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var posts []database.PostResponse
		for rows.Next() {
			var post database.PostResponse
			var created time.Time
			if err := rows.Scan(&post.ID, &post.Title, &post.Content, &created); err != nil {
				continue
			}
			post.TimeAgo = utils.FormatTimeAgo(created)
			if len(post.Content) > 160 {
				post.Excerpt = post.Content[:160] + "…"
			} else {
				post.Excerpt = post.Content
			}
			posts = append(posts, post)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// /api/user/comments
func UserCommentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		userID := utils.GetUserIDFromSession(getCookieValue(r))
		if userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		db := database.CreateTable()
		defer db.Close()

		rows, err := db.Query(`
			SELECT c.comment_id, c.post_id, p.title, c.content, c.creation_date
			FROM Comments c
			JOIN Posts p ON p.post_id = c.post_id
			WHERE c.user_id = ?
			ORDER BY c.creation_date DESC
		`, userID)
		if err != nil {
			http.Error(w, "Database query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type CommentItem struct {
			ID      int    `json:"id"`
			PostID  int    `json:"postId"`
			Title   string `json:"title"`
			Content string `json:"content"`
			TimeAgo string `json:"timeAgo"`
		}

		var out []CommentItem
		for rows.Next() {
			var it CommentItem
			var created time.Time
			if err := rows.Scan(&it.ID, &it.PostID, &it.Title, &it.Content, &created); err != nil {
				continue
			}
			it.TimeAgo = utils.FormatTimeAgo(created)
			out = append(out, it)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// /api/user/likes
func UserLikesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cookie, err := r.Cookie("session")
		if err != nil || !utils.IsValidSession(cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID := utils.GetUserIDFromSession(cookie.Value)
		if userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		db := database.CreateTable()
		defer db.Close()

		rows, err := db.Query(`
			SELECT p.post_id, p.title, p.content, p.creation_date
			FROM LikesDislikes ld
			JOIN Posts p ON p.post_id = ld.post_id
			WHERE ld.user_id = ? AND ld.vote = 1
			ORDER BY p.creation_date DESC
		`, userID)
		if err != nil {
			http.Error(w, "Database query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var liked []database.PostResponse
		for rows.Next() {
			var pr database.PostResponse
			var created time.Time
			if err := rows.Scan(&pr.ID, &pr.Title, &pr.Content, &created); err != nil {
				continue
			}
			pr.TimeAgo = utils.FormatTimeAgo(created)
			if len(pr.Content) > 160 {
				pr.Excerpt = pr.Content[:160] + "…"
			} else {
				pr.Excerpt = pr.Content
			}
			liked = append(liked, pr)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(liked)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// /api/user/dislikes
func UserDislikesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cookie, err := r.Cookie("session")
		if err != nil || !utils.IsValidSession(cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID := utils.GetUserIDFromSession(cookie.Value)
		if userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		db := database.CreateTable()
		defer db.Close()

		rows, err := db.Query(`
			SELECT p.post_id, p.title, p.content, p.creation_date
			FROM LikesDislikes ld
			JOIN Posts p ON p.post_id = ld.post_id
			WHERE ld.user_id = ? AND ld.vote = -1
			ORDER BY p.creation_date DESC
		`, userID)
		if err != nil {
			http.Error(w, "Database query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var disliked []database.PostResponse
		for rows.Next() {
			var pr database.PostResponse
			var created time.Time
			if err := rows.Scan(&pr.ID, &pr.Title, &pr.Content, &created); err != nil {
				continue
			}
			pr.TimeAgo = utils.FormatTimeAgo(created)
			if len(pr.Content) > 160 {
				pr.Excerpt = pr.Content[:160] + "…"
			} else {
				pr.Excerpt = pr.Content
			}
			disliked = append(disliked, pr)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(disliked)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
