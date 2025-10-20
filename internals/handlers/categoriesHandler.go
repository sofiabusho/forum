package handlers

import (
	"encoding/json"
	"forum/internals/database"
	"net/http"
)

// CategoriesAPIHandler returns all categories as JSON
func CategoriesAPIHandler(w http.ResponseWriter, r *http.Request) {
	db := database.CreateTable()
	defer db.Close()

	// Debug: Check if we can connect to database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM Categories").Scan(&count)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	// Get all categories
	rows, err := db.Query("SELECT name FROM Categories ORDER BY name")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []database.CategoryResponse
	for rows.Next() {
		var categoryName string
		err := rows.Scan(&categoryName)
		if err != nil {
			continue
		}

		// Create category response with basic info
		category := database.CategoryResponse{
			Name:        categoryName,
			Description: getCategoryDescription(categoryName),
			Tags:        []string{categoryName}, 
		}

		categories = append(categories, category)
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// CategoryResponse struct
type CategoryResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// getCategoryDescription returns a description for each category
func getCategoryDescription(categoryName string) string {
	descriptions := map[string]string{
		"Succulents":       "Low-maintenance plants perfect for beginners",
		"Tropical Plants":  "Exotic plants that bring the tropics indoors",
		"Herb Garden":      "Edible plants for cooking and natural remedies",
		"Indoor Plants":    "Plants that thrive in indoor environments",
		"Plant Care Tips":  "General advice and tips for plant care",
		"Plant Diseases":   "Help with identifying and treating plant problems",
		"Propagation":      "Growing new plants from existing ones",
		"Flowering Plants": "Plants known for their beautiful blooms",
	}

	if desc, exists := descriptions[categoryName]; exists {
		return desc
	}
	return "Share your knowledge about " + categoryName
}
