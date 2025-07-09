package handlers

import (
	"fmt"
	"forum/internals/database"
	"forum/internals/utils"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"
)

const (
	MaxFileSize   = 20 * 1024 * 1024 // 20MB
	UploadDir     = "frontend/uploads/images"
	ThumbnailDir  = "frontend/uploads/thumbnails"
	ThumbnailSize = 300
)

// ImageUploadHandler handles image upload for posts
func ImageUploadHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse multipart form (32MB max memory)
	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Get the file from form
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file size
	if fileHeader.Size > MaxFileSize {
		http.Error(w, "File too large. Maximum size is 20MB", http.StatusBadRequest)
		return
	}

	// Validate file type
	fileType, err := validateImageType(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Reset file pointer to beginning
	file.Seek(0, 0)

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%d_%s", userID, timestamp, sanitizeFilename(fileHeader.Filename))

	// Ensure upload directories exist
	os.MkdirAll(UploadDir, 0755)
	os.MkdirAll(ThumbnailDir, 0755)

	// Save original image
	imagePath := filepath.Join(UploadDir, filename)
	dst, err := os.Create(imagePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		os.Remove(imagePath) // Clean up on error
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Create thumbnail
	thumbnailPath, err := createThumbnail(imagePath, filename, fileType)
	if err != nil {
		// Log error but don't fail the upload
		fmt.Printf("Warning: Could not create thumbnail: %v\n", err)
	}

	// Save image info to database
	db := database.CreateTable()
	defer db.Close()

	imageURL := "/frontend/uploads/images/" + filename
	thumbnailURL := ""
	if thumbnailPath != "" {
		thumbnailURL = "/frontend/uploads/thumbnails/" + filename
	}

	_, err = db.Exec(`
		INSERT INTO Images (user_id, filename, original_name, file_size, file_type, image_url, thumbnail_url, upload_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, userID, filename, fileHeader.Filename, fileHeader.Size, fileType, imageURL, thumbnailURL)

	if err != nil {
		// Clean up files on database error
		os.Remove(imagePath)
		if thumbnailPath != "" {
			os.Remove(thumbnailPath)
		}
		http.Error(w, "Error saving image info", http.StatusInternalServerError)
		return
	}

	// Return success response with image URL
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"success": true,
		"filename": "%s",
		"imageUrl": "%s",
		"thumbnailUrl": "%s",
		"fileSize": %d,
		"fileType": "%s"
	}`, filename, imageURL, thumbnailURL, fileHeader.Size, fileType)
}

// validateImageType checks if the uploaded file is a valid image type
func validateImageType(file io.ReadSeeker) (string, error) {
	// Read first 512 bytes for type detection
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("error reading file")
	}

	// Reset file pointer
	file.Seek(0, 0)

	// Detect content type
	contentType := http.DetectContentType(buffer)

	switch contentType {
	case "image/jpeg":
		return "JPEG", nil
	case "image/png":
		return "PNG", nil
	case "image/gif":
		return "GIF", nil
	default:
		return "", fmt.Errorf("unsupported file type. Only JPEG, PNG, and GIF are allowed")
	}
}

// createThumbnail creates a thumbnail version of the uploaded image
func createThumbnail(originalPath, filename, fileType string) (string, error) {
	// Open original image
	originalFile, err := os.Open(originalPath)
	if err != nil {
		return "", err
	}
	defer originalFile.Close()

	// Decode image based on type
	var img image.Image
	switch fileType {
	case "JPEG":
		img, err = jpeg.Decode(originalFile)
	case "PNG":
		img, err = png.Decode(originalFile)
	case "GIF":
		img, err = gif.Decode(originalFile)
	default:
		return "", fmt.Errorf("unsupported image type for thumbnail")
	}

	if err != nil {
		return "", err
	}

	// Calculate thumbnail dimensions (maintain aspect ratio)
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	var newWidth, newHeight int
	if width > height {
		newWidth = ThumbnailSize
		newHeight = (height * ThumbnailSize) / width
	} else {
		newHeight = ThumbnailSize
		newWidth = (width * ThumbnailSize) / height
	}

	// Create thumbnail image
	thumbnail := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.NearestNeighbor.Scale(thumbnail, thumbnail.Rect, img, bounds, draw.Over, nil)

	// Save thumbnail
	thumbnailPath := filepath.Join(ThumbnailDir, filename)
	thumbnailFile, err := os.Create(thumbnailPath)
	if err != nil {
		return "", err
	}
	defer thumbnailFile.Close()

	// Encode thumbnail (always save as JPEG for consistency)
	err = jpeg.Encode(thumbnailFile, thumbnail, &jpeg.Options{Quality: 80})
	if err != nil {
		os.Remove(thumbnailPath)
		return "", err
	}

	return thumbnailPath, nil
}

// sanitizeFilename removes potentially dangerous characters from filename
func sanitizeFilename(filename string) string {
	// Remove path separators and other dangerous characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	filename = strings.ReplaceAll(filename, " ", "_")

	// Limit filename length
	if len(filename) > 100 {
		ext := filepath.Ext(filename)
		name := filename[:100-len(ext)]
		filename = name + ext
	}

	return filename
}

// DeleteImageHandler handles image deletion
func DeleteImageHandler(w http.ResponseWriter, r *http.Request) {
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
	filename := r.FormValue("filename")

	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	db := database.CreateTable()
	defer db.Close()

	// Verify user owns this image
	var imageUserID int
	var imageURL, thumbnailURL string
	err = db.QueryRow("SELECT user_id, image_url, thumbnail_url FROM Images WHERE filename = ?", filename).Scan(&imageUserID, &imageURL, &thumbnailURL)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	if imageUserID != userID {
		http.Error(w, "Unauthorized to delete this image", http.StatusForbidden)
		return
	}

	// Delete files from filesystem
	if imageURL != "" {
		imagePath := strings.TrimPrefix(imageURL, "/frontend/uploads/images/")
		os.Remove(filepath.Join(UploadDir, imagePath))
	}
	if thumbnailURL != "" {
		thumbnailPath := strings.TrimPrefix(thumbnailURL, "/frontend/uploads/thumbnails/")
		os.Remove(filepath.Join(ThumbnailDir, thumbnailPath))
	}

	// Delete from database
	_, err = db.Exec("DELETE FROM Images WHERE filename = ?", filename)
	if err != nil {
		http.Error(w, "Error deleting image record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"success": true}`)
}
