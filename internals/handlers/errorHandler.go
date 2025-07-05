package handlers

import (
	"forum/internals/utils"
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

// CustomErrorHandler is a middleware to catch errors and serve appropriate pages
func CustomErrorHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				InternalServerErrorHandler(w, r)
			}
		}()
		next(w, r)
	}
}
