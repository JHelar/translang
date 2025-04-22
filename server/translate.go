package server

import (
	"net/http"
	"translang/template"
)

func (client ServerClient) TranslateRoute(w http.ResponseWriter, r *http.Request) {
	figmaUrl := r.URL.Query().Get("figmaUrl")
	if figmaUrl == "" {
		w.WriteHeader(422)
		return
	}

	result, err := client.translator.Process(figmaUrl)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	template.Translate(template.TranslateProps{
		ProcessResult: result,
	}).Render(r.Context(), w)
}
