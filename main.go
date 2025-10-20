package main

import (
	"fmt"
	"forum/internals/database"
	"forum/internals/handlers"
	"net/http"
)

func main() {
	// Initialize database
	database.InitializeDatabase()

	fmt.Println("Server running on http://localhost:8080")

	// Setup routes
	handlers.SetupRoutes()

	// Start server
	http.ListenAndServe(":8080", nil)
}
