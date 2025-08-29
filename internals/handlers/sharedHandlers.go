package handlers

import (
	"database/sql"
	"forum/internals/utils"
)

// getPostTags returns the category tags for a given post
func getPostTags(db *sql.DB, postID int) []string {
	rows, err := db.Query(`
		SELECT c.name FROM Categories c
		JOIN PostCategories pc ON c.category_id = pc.category_id
		WHERE pc.post_id = ?`, postID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		rows.Scan(&tag)
		tags = append(tags, tag)
	}
	return tags
}

// getUsernameFromSession returns the username for a given session cookie
func GetUsernameFromSession(cookieValue string) string {
	return utils.GetUsernameFromSession(cookieValue)
}

