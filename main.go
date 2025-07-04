package main

import (
	"fmt"
	"forum/internals/handlers"
	"net/http"
)

func main() {
	fmt.Println("🚀 Server running on http://localhost:8080")

	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)

	// Static files (CSS, images)
	fs := http.FileServer(http.Dir("frontend"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Homepage
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/templates/index-unsigned.html")
	})

	http.ListenAndServe(":8080", nil)
}
