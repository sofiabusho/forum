package handlers

import (
	"forum/internals/utils" // 👈 import το πακέτο utils
	"net/http"
	"strings"
)

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/forgot-password.html", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	if !strings.Contains(email, "@") {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	// ➤ Εδώ μπορείς να δημιουργήσεις πραγματικό token (π.χ. uuid ή jwt)
	token := "dummy-reset-token"

	// ➤ Αποστολή email
	err := utils.SendResetEmail(email, token)
	if err != nil {
		http.Error(w, "Failed to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Αν όλα πάνε καλά, κάνε redirect ή εμφάνισε επιβεβαίωση
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
