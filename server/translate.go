package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"translang/dto"
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
	translation, err := dto.NewTranslation(figmaUrl, client.db)
	if err != nil {
		fmt.Printf("Error creating translation: %v\n", err)
		w.WriteHeader(500)
		return
	}

	fmt.Println(translation)

	template.TranslationBase(fmt.Sprintf("/translation/stream?id=%d", translation.ID)).Render(r.Context(), w)
}

func (client ServerClient) TranslationStream(w http.ResponseWriter, r *http.Request) {
	translationID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		w.WriteHeader(422)
		return
	}

	translation, err := dto.GetTranslationByID(translationID, client.db)
	if err != nil {
		fmt.Printf("Error getting translation (%d): %v", translationID, err)
		w.WriteHeader(404)
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

	go client.translator.ProcessContextImage(translation.FigmaSourceUrl, imageUrlChan, errorChan)
	go client.translator.ProcessTextTranslations(translation.FigmaSourceUrl, translationChan, errorChan)

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
