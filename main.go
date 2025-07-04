package main

import (
	"fmt"
	"forum/internals/handlers"
	"forum/internals/utils"
	"net/http"
)

func main() {
	fmt.Println("🚀 Server running on http://localhost:8080")

	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)

	// Static files (CSS, images)
	fs := http.FileServer(http.Dir("frontend/"))
	http.Handle("/frontend/", http.StripPrefix("/frontend/", fs))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("frontend/css"))))

	// Homepage
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// try to read the “session” cookie
		if cookie, err := r.Cookie("session"); err == nil && utils.IsValidSession(cookie.Value) {
			utils.FileService("index-signed.html", w, nil)
		} else {
			utils.FileService("index-unsigned.html", w, nil)
		}
	})
	http.ListenAndServe(":8080", nil)
}
