package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func routeVarsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		vars := mux.Vars(r)

		for key, value := range vars {
			r.Form.Add(key, value)
		}

		next.ServeHTTP(w, r)
	})
}
