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
		UnauthorizedHandler(w, r)
		return
	}

	// Parse form data
	err = r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		BadRequestHandler(w, r)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))

	categoryNames := r.Form["categories[]"]
	imageIDStr := r.FormValue("image_id")

	if title == "" || content == "" {
		utils.FileService("new-post.html", w, map[string]interface{}{"Error": "All fields are required"})
		return
	}

	if len(categoryNames) == 0 {
		utils.FileService("new-post.html", w, map[string]interface{}{
			"Error": "Please select at least one category",
		})
		return
	}

	// Validate categories and get their IDs
	db := database.CreateTable()
	defer db.Close()

	var categoryIDs []int
	for _, categoryName := range categoryNames {
		var categoryID int
		err = db.QueryRow("SELECT category_id FROM Categories WHERE name = ?", categoryName).Scan(&categoryID)
		if err != nil {
			utils.FileService("new-post.html", w, map[string]interface{}{
				"Error": fmt.Sprintf("Invalid category: %s", categoryName),
			})
			return
		}
		categoryIDs = append(categoryIDs, categoryID)
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

	// Start transaction for post creation
	tx, err := db.Begin()
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}
	defer tx.Rollback()

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

	// Associate post with all selected categories
	for _, categoryID := range categoryIDs {
		_, err = tx.Exec("INSERT INTO PostCategories (post_id, category_id) VALUES (?, ?)",
			postID, categoryID)
		if err != nil {
			http.Error(w, "Failed to associate categories", http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to save post", http.StatusInternalServerError)
		return
	}

	// Redirect to the new post or home page
	http.Redirect(w, r, fmt.Sprintf("/view-post?id=%d", postID), http.StatusSeeOther)
}

// RemovePostFromCategories removes all category associations for a post
func RemovePostFromCategories(db *sql.DB, postID int) error {
	_, err := db.Exec("DELETE FROM PostCategories WHERE post_id = ?", postID)
	return err
}

// ValidateCategories checks if all provided category names exist in the database
func ValidateCategories(db *sql.DB, categoryNames []string) ([]int, error) {
	var categoryIDs []int

	for _, name := range categoryNames {
		var categoryID int
		err := db.QueryRow("SELECT category_id FROM Categories WHERE name = ?", name).Scan(&categoryID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("category '%s' does not exist", name)
			}
			return nil, fmt.Errorf("database error while validating category '%s': %v", name, err)
		}
		categoryIDs = append(categoryIDs, categoryID)
	}

	return categoryIDs, nil
}

// PostsAPIHandler returns posts as JSON for dynamic loading (index.html)
func PostsAPIHandler(w http.ResponseWriter, r *http.Request) {

	db := database.CreateTable()
	defer db.Close()

	var currentUserID int
	if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
		currentUserID = utils.GetUserIDFromSession(cookie.Value)
	}

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
				   	(SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = -1) as dislike_count,

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
				   	(SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = -1) as dislike_count,
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
		var imageURL, thumbnailURL sql.NullString

		err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &creationDate, &p.Comments, &p.Likes, &p.Dislikes, &imageURL, &thumbnailURL)
		if err != nil {
			continue
		}

		p.TimeAgo = utils.FormatTimeAgo(creationDate)
		p.Excerpt = utils.TruncateText(p.Content, 150)
		p.Tags = getPostTags(db, p.ID)

		// Add image URLs if available
		if imageURL.Valid {
			p.ImageURL = imageURL.String
		}
		if thumbnailURL.Valid {
			p.ThumbnailURL = thumbnailURL.String
		}

		// Get user's vote status if logged in
		if currentUserID > 0 {
			var userVote int
			err := db.QueryRow("SELECT vote FROM LikesDislikes WHERE post_id = ? AND user_id = ?", p.ID, currentUserID).Scan(&userVote)
			if err != nil {
				userVote = 0 // No vote
			}
			p.UserVote = userVote
		}

		posts = append(posts, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// SinglePostAPIHandler returns a single post by ID (view-post.html)
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
		BadRequestHandler(w, r)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		BadRequestHandler(w, r)
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
        SELECT p.post_id, p.title, p.content, u.username, p.creation_date, p.user_id,
               (SELECT COUNT(*) FROM Comments WHERE post_id = p.post_id) as comment_count,
               (SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = 1) as like_count,
			   (SELECT COUNT(*) FROM LikesDislikes WHERE post_id = p.post_id AND vote = -1) as dislike_count,
               i.image_url, i.thumbnail_url
        FROM Posts p 
        JOIN Users u ON p.user_id = u.user_id
        LEFT JOIN Images i ON p.image_id = i.image_id
        WHERE p.post_id = ?`

	var post database.PostResponse
	var creationDate time.Time
	var imageURL, thumbnailURL *string
	var postAuthorID int

	err = db.QueryRow(query, postID).Scan(
		&post.ID, &post.Title, &post.Content, &post.Author,
		&creationDate, &postAuthorID, &post.Comments, &post.Likes, &post.Dislikes,
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
	post.TimeAgo = utils.FormatTimeAgo(creationDate)
	post.Tags = getPostTags(db, post.ID)

	// Add image URLs if available
	if imageURL != nil {
		post.ImageURL = *imageURL
	}
	if thumbnailURL != nil {
		post.ThumbnailURL = *thumbnailURL
	}

	// Get user's vote status if logged in
	var userVote int
	if currentUserID > 0 {
		err := db.QueryRow("SELECT vote FROM LikesDislikes WHERE post_id = ? AND user_id = ?", postID, currentUserID).Scan(&userVote)
		if err != nil {
			userVote = 0
		}
	}

	// Check if current user is the author
	isAuthor := currentUserID > 0 && currentUserID == postAuthorID

	response := map[string]interface{}{
		"id":           post.ID,
		"title":        post.Title,
		"content":      post.Content,
		"author":       post.Author,
		"timeAgo":      post.TimeAgo,
		"tags":         post.Tags,
		"comments":     post.Comments,
		"likes":        post.Likes,
		"dislikes":     post.Dislikes,
		"imageUrl":     post.ImageURL,
		"thumbnailUrl": post.ThumbnailURL,
		"userVote":     userVote,
		"isAuthor":     isAuthor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		postID := r.FormValue("post_id")
		title := strings.TrimSpace(r.FormValue("title"))
		content := strings.TrimSpace(r.FormValue("content"))
		categoryNames := r.Form["categories[]"]

		if title == "" || content == "" {
			http.Error(w, "Title and content required", http.StatusBadRequest)
			return
		}

		if len(categoryNames) == 0 {
			http.Error(w, "At least one category is required", http.StatusBadRequest)
			return
		}

		// Check authentication and ownership
		cookie, err := r.Cookie("session")
		if err != nil || !utils.IsValidSession(cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID := utils.GetUserIDFromSession(cookie.Value)

		// Get post and verify ownership
		db := database.CreateTable()
		defer db.Close()

		var authorID int
		err = db.QueryRow("SELECT user_id FROM Posts WHERE post_id = ?", postID).Scan(&authorID)
		if err != nil || authorID != userID {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		// Validate categories and get their IDs
		var categoryIDs []int
		for _, categoryName := range categoryNames {
			var categoryID int
			err = db.QueryRow("SELECT category_id FROM Categories WHERE name = ?", categoryName).Scan(&categoryID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid category: %s", categoryName), http.StatusBadRequest)
				return
			}
			categoryIDs = append(categoryIDs, categoryID)
		}

		// Start transaction for atomic updates
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Update post title and content
		_, err = tx.Exec("UPDATE Posts SET title = ?, content = ? WHERE post_id = ?", title, content, postID)
		if err != nil {
			http.Error(w, "Failed to update post", http.StatusInternalServerError)
			return
		}

		// Delete existing post-category associations
		_, err = tx.Exec("DELETE FROM PostCategories WHERE post_id = ?", postID)
		if err != nil {
			http.Error(w, "Failed to update categories", http.StatusInternalServerError)
			return
		}

		// Insert new post-category associations
		for _, categoryID := range categoryIDs {
			_, err = tx.Exec("INSERT INTO PostCategories (post_id, category_id) VALUES (?, ?)", postID, categoryID)
			if err != nil {
				http.Error(w, "Failed to update categories", http.StatusInternalServerError)
				return
			}
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Failed to save changes", http.StatusInternalServerError)
			return
		}

		// Return JSON response for API calls
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}
}

// DeletePostHandler handles post deletion
func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
		return
	}

	// Check authentication and ownership
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := utils.GetUserIDFromSession(cookie.Value)

	db := database.CreateTable()
	defer db.Close()

	var authorID int
	err = db.QueryRow("SELECT user_id FROM Posts WHERE post_id = ?", postID).Scan(&authorID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	if authorID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Delete related data first (foreign key constraints)
	db.Exec("DELETE FROM LikesDislikes WHERE post_id = ?", postID)
	db.Exec("DELETE FROM CommentLikes WHERE comment_id IN (SELECT comment_id FROM Comments WHERE post_id = ?)", postID)
	db.Exec("DELETE FROM Comments WHERE post_id = ?", postID)
	db.Exec("DELETE FROM PostCategories WHERE post_id = ?", postID)
	db.Exec("DELETE FROM Notifications WHERE related_post_id = ?", postID)

	// Delete the post
	_, err = db.Exec("DELETE FROM Posts WHERE post_id = ?", postID)
	if err != nil {
		http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// EditCommentHandler handles comment editing
func EditCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	commentID := r.FormValue("comment_id")
	content := strings.TrimSpace(r.FormValue("content"))

	if commentID == "" || content == "" {
		http.Error(w, "Comment ID and content required", http.StatusBadRequest)
		return
	}

	// Check authentication and ownership
	cookie, err := r.Cookie("session")
	if err != nil || !utils.IsValidSession(cookie.Value) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := utils.GetUserIDFromSession(cookie.Value)

	db := database.CreateTable()
	defer db.Close()

	var authorID int
	err = db.QueryRow("SELECT user_id FROM Comments WHERE comment_id = ?", commentID).Scan(&authorID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if authorID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Update comment
	_, err = db.Exec("UPDATE Comments SET content = ? WHERE comment_id = ?", content, commentID)
	if err != nil {
		http.Error(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
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
		WHERE user_id = ? And image_type = 'post'
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
