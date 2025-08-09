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

	if r.Method == "GET" {
		utils.FileService("profile.html", w, nil)
	} else if r.Method == "POST" {
		updateProfile(w, r)
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

	db := database.CreateTable()
	defer db.Close()

	profile := getUserProfile(db, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
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
		db.QueryRow("SELECT COUNT(*) FROM Users WHERE username = ? AND user_id != ?", newUsername, userID).Scan(&exists)
		if exists > 0 {
			http.Error(w, "Username already taken", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE Users SET username = ? WHERE user_id = ?", newUsername, userID)
		if err != nil {
			http.Error(w, "Failed to update username", http.StatusInternalServerError)
			return
		}
	}

	if newBio != "" {
		_, err := db.Exec("UPDATE Users SET bio = ? WHERE user_id = ?", newBio, userID)
		if err != nil {
			http.Error(w, "Failed to update bio", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"username": newUsername,
		"bio":      newBio,
	})
}

func getUserProfile(db *sql.DB, userID int) database.UserProfile {
	var profile database.UserProfile
	var registrationDate time.Time
	var bio string

	err := db.QueryRow(`
		SELECT user_id, username, email, registration_date, bio
		FROM Users 
		WHERE user_id = ?
	`, userID).Scan(&profile.UserID, &profile.Username, &profile.Email, &registrationDate, &bio)

	if err != nil {
		return profile
	}

	profile.Bio = bio
	profile.JoinDate = registrationDate.Format("January 2, 2006")
	db.QueryRow("SELECT COUNT(*) FROM Posts WHERE user_id = ?", userID).Scan(&profile.PostCount)
	db.QueryRow("SELECT COUNT(*) FROM Comments WHERE user_id = ?", userID).Scan(&profile.CommentCount)
	db.QueryRow("SELECT COUNT(*) FROM LikesDislikes WHERE user_id = ? AND vote = 1", userID).Scan(&profile.LikesGiven)
	db.QueryRow(`
		SELECT COUNT(*) 
		FROM LikesDislikes ld 
		JOIN Posts p ON ld.post_id = p.post_id 
		WHERE p.user_id = ? AND ld.vote = 1
	`, userID).Scan(&profile.LikesReceived)

	var thumbnailURL string
	db.QueryRow(`
		SELECT thumbnail_url 
		FROM Images 
		WHERE user_id = ? 
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

// ✅ ΝΕΑ FUNCTION για /api/user/posts
func UserPostsHandler(w http.ResponseWriter, r *http.Request) {
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
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &created)
		if err != nil {
			continue
		}
		post.TimeAgo = utils.FormatTimeAgo(created)
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// ✅ /api/user/comments — επιστρέφει τα σχόλιά μου (με τίτλο post & timeAgo)
func UserCommentsHandler(w http.ResponseWriter, r *http.Request) {
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
}

// ✅ /api/user/likes — επιστρέφει posts που έχω κάνει like (clickable σε κάθε post)
func UserLikesHandler(w http.ResponseWriter, r *http.Request) {
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
		// προαιρετικό μικρό απόσπασμα
		if len(pr.Content) > 160 {
			pr.Excerpt = pr.Content[:160] + "…"
		} else {
			pr.Excerpt = pr.Content
		}
		liked = append(liked, pr)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(liked)
}
