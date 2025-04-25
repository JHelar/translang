package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"translang/template"
	"translang/translator"
)

func (client ServerClient) TranslateRoute(w http.ResponseWriter, r *http.Request) {

	template.Translate().Render(r.Context(), w)
}

func (client ServerClient) TranslationRoute(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	figmaUrl := r.Form.Get("figmaUrl")
	if figmaUrl == "" {
		w.WriteHeader(422)
		return
	}

	template.TranslationBase(fmt.Sprintf("/translation/stream?figmaUrl=%s", url.QueryEscape(figmaUrl))).Render(r.Context(), w)
}

func (client ServerClient) TranslationStream(w http.ResponseWriter, r *http.Request) {
	figmaUrl := r.URL.Query().Get("figmaUrl")
	if figmaUrl == "" {
		w.WriteHeader(422)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	imageUrlChan := make(chan string)
	translationChan := make(chan translator.TranslationResult)
	errorChan := make(chan error)

	var stringBuilder strings.Builder

	go client.translator.ProcessContextImage(figmaUrl, imageUrlChan, errorChan)
	go client.translator.ProcessTextTranslations(figmaUrl, translationChan, errorChan)

	moreTranslations := true
	for moreTranslations {
		select {
		case translation, moreTranslations := <-translationChan:
			if moreTranslations {
				stringBuilder.WriteString("event:translation\ndata:")
				template.TranslationNode(translation).Render(r.Context(), &stringBuilder)
				stringBuilder.WriteString("\n\n")

				w.Write([]byte(stringBuilder.String()))
				w.(http.Flusher).Flush()
				stringBuilder.Reset()
			}
		case contextImageUrl := <-imageUrlChan:
			stringBuilder.WriteString("event:contextImage\ndata:")
			template.TranslationContextImage(contextImageUrl).Render(r.Context(), &stringBuilder)
			stringBuilder.WriteString("\n\n")

			w.Write([]byte(stringBuilder.String()))
			w.(http.Flusher).Flush()
			stringBuilder.Reset()
		case err := <-errorChan:
			fmt.Printf("Error generating translations: %v\n", err)
			moreTranslations = false
		}
	}
}
