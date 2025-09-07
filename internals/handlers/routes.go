package handlers

import (
	"forum/internals/utils"
	"log"
	"net/http"
	"strings"
	"time"
)

func SetupRoutes() {
	// Public routes
	wrapHandler("/", HomeHandler)
	wrapHandler("/login", LoginHandler)
	wrapHandler("/login.html", LoginHandler)
	wrapHandler("/register", RegisterHandler)
	wrapHandler("/register.html", RegisterHandler)

	// OAuth routes (public)
	wrapHandler("/auth/google", GoogleLogin)
	wrapHandler("/auth/google/callback", GoogleCallback)
	wrapHandler("/auth/github", GitHubLogin)
	wrapHandler("/auth/github/callback", GitHubCallback)

	// Protected routes (require authentication)
	wrapProtectedHandler("/logout", LogoutHandler)
	wrapProtectedHandler("/profile", ProfileHandler)
	wrapProtectedHandler("/profile.html", ProfilePageHandler)
	wrapProtectedHandler("/api/user/profile", ProfileAPIHandler)
	wrapProtectedHandler("/new-post", CreatePostHandler)
	wrapProtectedHandler("/new-post.html", CreatePostHandler)
	wrapProtectedHandler("/api/posts/edit", EditPostHandler)
	wrapProtectedHandler("/api/posts/delete", DeletePostHandler)
	wrapProtectedHandler("/api/comments/create", CreateCommentHandler)
	wrapProtectedHandler("/api/comments/edit", EditCommentHandler)
	wrapProtectedHandler("/api/comments/delete", DeleteCommentHandler)
	wrapProtectedHandler("/api/posts/like", LikePostHandler)
	wrapProtectedHandler("/api/comments/like", LikeCommentHandler)
	wrapProtectedHandler("/api/user/posts", UserPostsHandler)
	wrapProtectedHandler("/api/user/comments", UserCommentsHandler)
	wrapProtectedHandler("/api/user/likes", UserLikesHandler)
	wrapProtectedHandler("/api/user/dislikes", UserDislikesHandler)
	wrapProtectedHandler("/api/posts/filtered", FilteredPostsHandler)
	wrapProtectedHandler("/api/notifications", NotificationsAPIHandler)
	wrapProtectedHandler("/api/notifications/mark-read", MarkNotificationReadHandler)
	wrapProtectedHandler("/api/notifications/mark-all-read", MarkAllNotificationsReadHandler)
	wrapProtectedHandler("/api/notifications/count", NotificationCountHandler)
	wrapProtectedHandler("/api/upload-image", ImageUploadHandler)

	// Semi-public routes (content varies by auth status)
	wrapHandler("/view-post", ViewPostHandler)
	wrapHandler("/view-post.html", ViewPostHandler)
	wrapHandler("/api/posts", PostsAPIHandler)
	wrapHandler("/api/post", SinglePostAPIHandler)
	wrapHandler("/api/comments", CommentsAPIHandler)
	wrapHandler("/api/categories", CategoriesAPIHandler)
	wrapHandler("/api/auth/status", AuthStatusHandler)

	// Category routes
	wrapHandler("/categories", CategoriesPageHandler)
	wrapHandler("/categories.html", CategoriesPageHandler)

	// Password reset routes
	wrapHandler("/forgot-password", ForgotPasswordHandler)
	wrapHandler("/forgot-password.html", ForgotPasswordPageHandler)
	wrapHandler("/reset-password", ResetPasswordHandler)
	wrapHandler("/add-newpassword.html", ShowResetFormHandler)

	// Error routes
	wrapHandler("/400", BadRequestHandler)
	wrapHandler("/404", NotFoundHandler)
	wrapHandler("/500", InternalServerErrorHandler)

	// Static files (no middleware needed)
	fs := http.FileServer(http.Dir("frontend/"))
	http.Handle("/frontend/", http.StripPrefix("/frontend/", fs))
}

// wrapHandler wraps handlers with error handling
func wrapHandler(path string, handler http.HandlerFunc) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Error handling
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic on %s %s: %v", r.Method, r.URL.Path, err)
				InternalServerErrorHandler(w, r)
			}
		}()

		// Call the actual handler
		handler(w, r)

		// Logging
		duration := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, 200, duration)
	})
}

// wrapProtectedHandler wraps handlers that require authentication
func wrapProtectedHandler(path string, handler http.HandlerFunc) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Authentication check
		cookie, err := r.Cookie("session")
		if err != nil || !utils.IsValidSession(cookie.Value) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login?error=unauthorized", http.StatusSeeOther)
			return
		}

		// Error handling
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic on %s %s: %v", r.Method, r.URL.Path, err)
				InternalServerErrorHandler(w, r)
			}
		}()

		// Call the actual handler
		handler(w, r)

		// Logging
		duration := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, 200, duration)
	})
}
