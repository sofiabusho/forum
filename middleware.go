package main

import (
	"forum/internals/handlers"
	"log"
	"net/http"
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

// CombinedMiddleware applies multiple middlewares
func CombinedMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return LoggingMiddleware(ErrorHandlingMiddleware(next))
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

// wrapHandler wraps a regular handler function with middleware
func wrapHandler(pattern string, handler http.HandlerFunc) {
	http.HandleFunc(pattern, CombinedMiddleware(handler))
}

// wrapHandlerNoLog wraps a handler without logging (for static files)
func wrapHandlerNoLog(pattern string, handler http.HandlerFunc) {
	http.HandleFunc(pattern, ErrorHandlingMiddleware(handler))
}
