package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/internals/database"
	"forum/internals/utils"
	"net/http"
	"strconv"
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
	userID := utils.GetUserIDFromSession(cookie.Value)
	if userID == 0 {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Parse form data
	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	categoryName := r.FormValue("categories")
	imageIDStr := r.FormValue("image_id")

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

	// Validate image ID if provided
	var imageID *int
	if imageIDStr != "" {
		// Verify the image exists and belongs to this user
		var imageUserID int
		var imageFilename string
		err = db.QueryRow("SELECT user_id, filename FROM Images WHERE filename = ?", imageIDStr).Scan(&imageUserID, &imageFilename)
		if err != nil {
			utils.FileService("new-post.html", w, map[string]interface{}{"Error": "Invalid image selected"})
			return
		}
		if imageUserID != userID {
			utils.FileService("new-post.html", w, map[string]interface{}{"Error": "You can only use your own images"})
			return
		}

		// Get the actual image ID
		var actualImageID int
		err = db.QueryRow("SELECT image_id FROM Images WHERE filename = ?", imageIDStr).Scan(&actualImageID)
		if err == nil {
			imageID = &actualImageID
		}
	}

	// Insert post into database with optional image
	var result sql.Result
	if imageID != nil {
		result, err = db.Exec("INSERT INTO Posts (user_id, title, content, image_id) VALUES (?, ?, ?, ?)", userID, title, content, *imageID)
	} else {
		result, err = db.Exec("INSERT INTO Posts (user_id, title, content) VALUES (?, ?, ?)", userID, title, content)
	}

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
			       (SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = 1) as like_count,
			       i.image_url, i.thumbnail_url
			FROM Posts p 
			JOIN Users u ON p.user_id = u.user_id
			LEFT JOIN Images i ON p.image_id = i.image_id
			JOIN PostCategories pc ON p.post_id = pc.post_id
			JOIN Categories c ON pc.category_id = c.category_id
			WHERE c.name = ?
			ORDER BY p.creation_date DESC`
		args = append(args, categoryValue)
	default:
		query = `
			SELECT p.post_id, p.title, p.content, u.username, p.creation_date,
			       (SELECT COUNT(*) FROM Comments WHERE post_id = p.post_id) as comment_count,
			       (SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = 1) as like_count,
			       i.image_url, i.thumbnail_url
			FROM Posts p 
			JOIN Users u ON p.user_id = u.user_id
			LEFT JOIN Images i ON p.image_id = i.image_id
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
		var imageURL, thumbnailURL *string

		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &creationDate, &p.Comments, &p.Likes, &imageURL, &thumbnailURL)
		if err != nil {
			continue
		}

		p.TimeAgo = formatTimeAgo(creationDate)
		p.Excerpt = truncateText(p.Content, 150)
		p.Tags = getPostTags(db, p.ID)
		p.Views = getPostViews(db, p.ID)

		// Add image URLs if available
		if imageURL != nil {
			p.ImageURL = *imageURL
		}
		if thumbnailURL != nil {
			p.ThumbnailURL = *thumbnailURL
		}

		posts = append(posts, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// SinglePostAPIHandler returns a single post by ID
func SinglePostAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path or query parameter
	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		// Try to extract from path if using /api/post/{id} pattern
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 3 {
			postIDStr = parts[len(parts)-1]
		}
	}

	if postIDStr == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get current user ID if logged in (for vote status)
	var currentUserID int
	if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
		currentUserID = utils.GetUserIDFromSession(cookie.Value)
	}

	db := database.CreateTable()
	defer db.Close()

	// Query for single post with image information
	query := `
        SELECT p.post_id, p.title, p.content, u.username, p.creation_date,
               (SELECT COUNT(*) FROM Comments WHERE post_id = p.post_id) as comment_count,
               (SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = 1) as like_count,
               i.image_url, i.thumbnail_url
        FROM Posts p 
        JOIN Users u ON p.user_id = u.user_id
        LEFT JOIN Images i ON p.image_id = i.image_id
        WHERE p.post_id = ?`

	var post database.PostResponse
	var creationDate time.Time
	var imageURL, thumbnailURL *string

	err = db.QueryRow(query, postID).Scan(
		&post.ID, &post.Title, &post.Content, &post.Author,
		&creationDate, &post.Comments, &post.Likes,
		&imageURL, &thumbnailURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Format the post data
	post.TimeAgo = formatTimeAgo(creationDate)
	post.Tags = getPostTags(db, post.ID)
	post.Views = getPostViews(db, post.ID)

	// Add image URLs if available
	if imageURL != nil {
		post.ImageURL = *imageURL
	}
	if thumbnailURL != nil {
		post.ThumbnailURL = *thumbnailURL
	}

	// Get user's vote status if logged in
	if currentUserID > 0 {
		var userVote int
		err := db.QueryRow("SELECT vote FROM LikesDislikes WHERE post_id = ? AND user_id = ?", postID, currentUserID).Scan(&userVote)
		if err == nil {
			// Add userVote to response (you might need to extend PostResponse struct)
			// For now, we'll add it as a separate field
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// GetUserImagesHandler returns images uploaded by a user
func GetUserImagesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get user's uploaded images
	rows, err := db.Query(`
		SELECT image_id, filename, original_name, file_size, file_type, 
		       image_url, thumbnail_url, upload_date
		FROM Images 
		WHERE user_id = ? 
		ORDER BY upload_date DESC`, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var images []database.ImageResponse
	for rows.Next() {
		var img database.ImageResponse
		var uploadDate time.Time

		err := rows.Scan(&img.ID, &img.Filename, &img.OriginalName, &img.FileSize,
			&img.FileType, &img.ImageURL, &img.ThumbnailURL, &uploadDate)
		if err != nil {
			continue
		}

		img.UploadDate = uploadDate.Format("2006-01-02 15:04:05")
		img.FileSizeFormatted = formatFileSize(img.FileSize)

		images = append(images, img)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

// Helper function to format file size
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
