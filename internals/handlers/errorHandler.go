package handlers

import (
	"encoding/json"
	"forum/internals/utils"
	"log"
	"net/http"
)

// NotFoundHandler handles 404 errors
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	utils.FileService("error404.html", w, nil)
}

// InternalServerErrorHandler handles 500 errors
func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	utils.FileService("error500.html", w, nil)
}

// BadRequestHandler handles 400 errors
func BadRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	// You can create a error400.html template or use a generic error page
	utils.FileService("error500.html", w, map[string]interface{}{
		"Error":   "Bad Request",
		"Message": "The request could not be processed.",
	})
}

// UnauthorizedHandler handles 401 errors
func UnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	// Redirect to login for unauthorized access
	http.Redirect(w, r, "/login?error=unauthorized", http.StatusSeeOther)
}

// ForbiddenHandler handles 403 errors
func ForbiddenHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	utils.FileService("error500.html", w, map[string]interface{}{
		"Error":   "Access Forbidden",
		"Message": "You don't have permission to access this resource.",
	})
}

// CustomErrorHandler is a middleware to catch errors and serve appropriate pages
func CustomErrorHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic occurred on %s %s: %v", r.Method, r.URL.Path, err)
				InternalServerErrorHandler(w, r)
			}
		}()
		next(w, r)
	}
}

// ErrorResponse sends a JSON error response for API endpoints
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Create and send JSON response
	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"status":  statusCode,
	}

	json.NewEncoder(w).Encode(response)
}

// HandleHTTPError handles different HTTP errors based on status code
func HandleHTTPError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	// Check if this is an API request (JSON response expected)
	if r.Header.Get("Content-Type") == "application/json" ||
		r.Header.Get("Accept") == "application/json" ||
		r.URL.Path[:4] == "/api" {
		ErrorResponse(w, statusCode, message)
		return
	}

	// Handle HTML error pages
	switch statusCode {
	case http.StatusNotFound:
		NotFoundHandler(w, r)
	case http.StatusBadRequest:
		BadRequestHandler(w, r)
	case http.StatusUnauthorized:
		UnauthorizedHandler(w, r)
	case http.StatusForbidden:
		ForbiddenHandler(w, r)
	case http.StatusInternalServerError:
		fallthrough
	default:
		InternalServerErrorHandler(w, r)
	}
}

// NotFoundMiddleware catches all unmatched routes
func NotFoundMiddleware() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		NotFoundHandler(w, r)
	}
}
