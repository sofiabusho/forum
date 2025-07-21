package handlers

import (
	"html/template"
	"net/http"
)

var tmpl = template.Must(template.ParseGlob("frontend/templates/*.html")) // ή προσαρμόστε το path

func ShowResetFormHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	data := struct {
		Token string
	}{
		Token: token,
	}

	tmpl.ExecuteTemplate(w, "add-newpassword.html", data)
}
