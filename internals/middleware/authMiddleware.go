package middleware

import (
	"context"
	"forum/internals/handlers"
	"forum/internals/utils"
	"log"
	"net/http"
	"strings"
	"time"
)

// LoggingMiddleware logs all requests
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(lrw, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.statusCode, duration)
	}
}

// ErrorHandlingMiddleware catches panics and serves 500 error page
func ErrorHandlingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic occurred: %v", err)
				handlers.InternalServerErrorHandler(w, r)
			}
		}()
		next(w, r)
	}
}

// SecurityMiddleware adds security headers
func SecurityMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next(w, r)
	}
}

// RequireAuth middleware ensures user is authenticated
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || !utils.IsValidSession(cookie.Value) {
			// Handle differently for API vs HTML requests
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login?error=unauthorized", http.StatusSeeOther)
			return
		}

		// Add user info to request context
		userID := utils.GetUserIDFromSession(cookie.Value)
		ctx := context.WithValue(r.Context(), "userID", userID)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// OptionalAuth middleware adds user info if available but doesn't require auth
func OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err == nil && utils.IsValidSession(cookie.Value) {
			userID := utils.GetUserIDFromSession(cookie.Value)
			ctx := context.WithValue(r.Context(), "userID", userID)
			r = r.WithContext(ctx)
		}
		next(w, r)
	}
}

// Chain applies multiple middlewares in order
func Chain(h http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// Common middleware stack for all routes
func CommonStack(next http.HandlerFunc) http.HandlerFunc {
	return Chain(next,
		LoggingMiddleware,
		ErrorHandlingMiddleware,
		SecurityMiddleware,
	)
}

// Protected middleware stack (requires authentication)
func ProtectedStack(next http.HandlerFunc) http.HandlerFunc {
	return Chain(next,
		LoggingMiddleware,
		ErrorHandlingMiddleware,
		SecurityMiddleware,
		RequireAuth,
	)
}

// Optional auth middleware stack
func OptionalAuthStack(next http.HandlerFunc) http.HandlerFunc {
	return Chain(next,
		LoggingMiddleware,
		ErrorHandlingMiddleware,
		SecurityMiddleware,
		OptionalAuth,
	)
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
