package handlers

import (
	"encoding/json"
	"fmt"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strings"
	"time"
)

// CreatePostHandler handles post creation
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.FileService("new-post.html", w, nil)
		return
	}

	// Check if user is logged in
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user ID from session
	userID := getUserIDFromSession(cookie.Value)
	if userID == 0 {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Parse form data
	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	categoryName := r.FormValue("categories")

	if title == "" || content == "" || categoryName == "" {
		utils.FileService("new-post.html", w, map[string]interface{}{"Error": "All fields are required"})
		return
	}

	// Get category ID from name
	db := database.CreateTable()
	defer db.Close()

	var categoryID int
	err = db.QueryRow("SELECT category_id FROM Categories WHERE name = ?", categoryName).Scan(&categoryID)
	if err != nil {
		utils.FileService("new-post.html", w, map[string]interface{}{"Error": "Invalid category selected"})
		return
	}

	// Insert post into database
	result, err := db.Exec("INSERT INTO Posts (user_id, title, content) VALUES (?, ?, ?)", userID, title, content)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	postID, _ := result.LastInsertId()

	// Associate post with category
	db.Exec("INSERT INTO PostCategories (post_id, category_id) VALUES (?, ?)", postID, categoryID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// PostsAPIHandler returns posts as JSON for dynamic loading
func PostsAPIHandler(w http.ResponseWriter, r *http.Request) {
	db := database.CreateTable()
	defer db.Close()

	filter := r.URL.Query().Get("filter")

	var query string
	var args []interface{}

	switch filter {
	case "categories":
		categoryValue := r.URL.Query().Get("value")
		query = `
			SELECT p.post_id, p.title, p.content, u.username, p.creation_date,
			       (SELECT COUNT(*) FROM Comments WHERE post_id = p.post_id) as comment_count,
			       (SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = 1) as like_count
			FROM Posts p 
			JOIN Users u ON p.user_id = u.user_id
			JOIN PostCategories pc ON p.post_id = pc.post_id
			JOIN Categories c ON pc.category_id = c.category_id
			WHERE c.name = ?
			ORDER BY p.creation_date DESC`
		args = append(args, categoryValue)
	default:
		query = `
			SELECT p.post_id, p.title, p.content, u.username, p.creation_date,
			       (SELECT COUNT(*) FROM Comments WHERE post_id = p.post_id) as comment_count,
			       (SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = 1) as like_count
			FROM Posts p 
			JOIN Users u ON p.user_id = u.user_id
			ORDER BY p.creation_date DESC`
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

		p.TimeAgo = formatTimeAgo(creationDate)
		p.Excerpt = truncateText(p.Content, 150)
		p.Tags = getPostTags(db, p.ID)
		p.Views = getPostViews(db, p.ID)

		posts = append(posts, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// CategoriesAPIHandler returns available categories
func CategoriesAPIHandler(w http.ResponseWriter, r *http.Request) {
	db := database.CreateTable()
	defer db.Close()

	rows, err := db.Query("SELECT category_id, name FROM Categories")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []database.CategoryResponse
	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			continue
		}

		categories = append(categories, database.CategoryResponse{
			Name:        name,
			Description: fmt.Sprintf("Posts about %s", name),
			Tags:        []string{name},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
