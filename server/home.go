package server

import (
	"net/http"
	"translang/template"
)

func (client ServerClient) HomeRoute(w http.ResponseWriter, r *http.Request) {
	template.Home().Render(r.Context(), w)
}
