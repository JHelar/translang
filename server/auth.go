package server

import (
	"fmt"
	"net/http"
	"translang/auth"
)

func (client ServerClient) SignIn(w http.ResponseWriter, r *http.Request) {
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	if email == "" {
		http.Error(w, "missing email", http.StatusBadRequest)
		return
	}
	if password == "" {
		http.Error(w, "missing password", http.StatusBadRequest)
		return
	}

	payload := auth.NewPasswordUserPayload(email, password)
	user, err := client.auth.SignIn(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Signed in as user: %d", user.ID)
}
