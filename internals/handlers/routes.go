package handlers

import (
	"fmt"
	"net/http"
)

func SetupRoutes() {
	// Authentication routes
	wrapHandler("/login",  LoginHandler)
	wrapHandler("/login.html",  LoginHandler)
	wrapHandler("/register",  RegisterHandler)
	wrapHandler("/register.html",  RegisterHandler)
	wrapHandler("/logout",  LogoutHandler)

	// Post routes
	wrapHandler("/new-post",  CreatePostHandler)
	wrapHandler("/new-post.html",  CreatePostHandler)
	wrapHandler("/api/posts",  PostsAPIHandler)
	wrapHandler("/api/posts/edit",  EditPostHandler)
	wrapHandler("/api/posts/delete",  DeletePostHandler)

	// Single post view
	wrapHandler("/view-post",  ViewPostHandler)
	wrapHandler("/view-post.html",  ViewPostHandler)
	wrapHandler("/api/post",  SinglePostAPIHandler) // API for single post, Since Go doesn't support URL parameters like /api/post/{id} natively

	// Comment routes
	wrapHandler("/api/comments/create",  CreateCommentHandler)
	wrapHandler("/api/comments",  CommentsAPIHandler)
	wrapHandler("/api/comments/edit",  EditCommentHandler)
	wrapHandler("/api/comments/delete",  DeleteCommentHandler)

	// Like/Dislike routes
	wrapHandler("/api/posts/like",  LikePostHandler)
	wrapHandler("/api/comments/like",  LikeCommentHandler)

	// User data routes
	http.HandleFunc("/api/user/posts",  UserPostsHandler)

	wrapHandler("/api/user/comments",  UserCommentsHandler)
	wrapHandler("/api/user/likes",  UserLikesHandler)
	wrapHandler("/api/user/dislikes", UserDislikesHandler)


	// Google OAuth routes
	http.HandleFunc("/auth/google",  GoogleLogin)
	http.HandleFunc("/auth/google/callback",  GoogleCallback)

	// GitHub OAuth routes
	http.HandleFunc("/auth/github",  GitHubLogin)
	http.HandleFunc("/auth/github/callback",  GitHubCallback)

	// Category routes
	wrapHandler("/api/categories",  CategoriesAPIHandler)
	wrapHandler("/categories",  CategoriesPageHandler)
	wrapHandler("/categories.html",  CategoriesPageHandler)

	// Forgot - Reset Password routs
	wrapHandler("/forgot-password",  ForgotPasswordHandler)
	wrapHandler("/forgot-password.html",  ForgotPasswordPageHandler)
	wrapHandler("/reset-password",  ResetPasswordHandler)
	wrapHandler("/add-newpassword.html",  ShowResetFormHandler)

	// Filter routes (for authenticated users)
	wrapHandler("/api/posts/filtered",  FilteredPostsHandler)

	// Auth status check - ONLY ONE REGISTRATION
	wrapHandler("/api/auth/status",  AuthStatusHandler)

	// Notifications API
	wrapHandler("/api/notifications", NotificationsAPIHandler)
	wrapHandler("/api/notifications/mark-read", MarkNotificationReadHandler)
	wrapHandler("/api/notifications/mark-all-read",  MarkAllNotificationsReadHandler)
	wrapHandler("/api/notifications/count",  NotificationCountHandler)

	// Image upload and management
	wrapHandler("/api/upload-image",  ImageUploadHandler)
	http.HandleFunc("/upload-image",  ImageUploadHandler)
	wrapHandler("/api/delete-image",  DeleteImageHandler)
	wrapHandler("/api/user-images",  GetUserImagesHandler)

	// Profile routes
	wrapHandler("/api/user/profile",  ProfileAPIHandler)
	wrapHandler("/profile",  ProfileHandler)
	wrapHandler("/profile.html",  ProfilePageHandler)

	// Notifications routes
	wrapHandler("/notifications", NotificationsPageHandler)
	wrapHandler("/notifications.html", NotificationsPageHandler)

	// Static pages
	wrapHandler("/about", AboutHandler)
	wrapHandler("/about.html", AboutHandler)
	wrapHandler("/terms", TermsHandler)
	wrapHandler("/terms.html", TermsHandler)

	// Error routes
	wrapHandler("/400", BadRequestHandler)
	wrapHandler("/404", NotFoundHandler)
	wrapHandler("/500", InternalServerErrorHandler)

	// Static files (CSS, images, JavaScript)
	fs := http.FileServer(http.Dir("frontend/"))
	http.Handle("/frontend/", http.StripPrefix("/frontend/", fs))

	// Homepage last to avoid conflicts
	wrapHandler("/", HomeHandler)
}

// wrapHandler wraps handlers with error handling
func wrapHandler(path string, handler http.HandlerFunc) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Panic on %s: %v\n", path, err)
				InternalServerErrorHandler(w, r)
			}
		}()
		handler(w, r)
	})
}
