package server

import (
	"fmt"
	"io"
	"net/http"
	"translang/server/sse"
	"translang/template"
	"translang/translator"
)

func (client ServerClient) TranslateRoute(w http.ResponseWriter, r *http.Request) {
	translations, err := client.persistence.GetAllTranslations()
	if err != nil {
		fmt.Printf("Error retreiving translation: %v\n", err)
		w.WriteHeader(500)
		return
	}

	results := []translator.ProcessResult{}
	for _, translation := range translations {
		result, err := translation.ToResult()
		if err != nil {
			w.WriteHeader(500)
			return
		}
		results = append(results, result)
	}

	props := template.TranslateProps{
		Results: results,
	}

	r.ParseForm()
	figmaUrl := r.Form.Get("figmaUrl")
	if figmaUrl != "" {
		translation, err := client.persistence.UpsertTranslation(figmaUrl)
		if err != nil {
			fmt.Printf("Error creating translation: %v\n", err)
		} else {
			props.Modal = template.TranslationModalProps{
				SSEUrl:         fmt.Sprintf("/translate/stream?id=%s", translation.GetID()),
				FigmaSourceUrl: figmaUrl,
			}
		}
	}

	template.Translate(props).Render(r.Context(), w)

}

func (client ServerClient) TranslateStreamRoute(w http.ResponseWriter, r *http.Request) {
	translation, err := client.persistence.GetTranslationByID(r.URL.Query().Get("id"))
	if err != nil {
		fmt.Printf("Error getting translation: %v\n", err)
		w.WriteHeader(404)
		return
	}

	sseClient := sse.NewClient(w)
	defer sseClient.Close()

	imageUrlChan := make(chan string)
	translationChan := make(chan translator.TranslationResult)
	errorChan := make(chan error)

	go func() {
		contextImageUrl, err := translation.GetContextImageUrl()
		if err == nil {
			imageUrlChan <- contextImageUrl
		} else {
			client.translator.ProcessContextImage(translation.GetFigmaSourceUrl(), imageUrlChan, errorChan)
		}
	}()

	go func() {
		defer close(translationChan)

		nodes, err := translation.GetAllNodes()
		if err != nil {
			errorChan <- err
			return
		}
		if len(nodes) > 0 {
			for _, node := range nodes {
				result, err := node.ToResult()
				if err != nil {
					errorChan <- err
					return
				}
				translationChan <- result
			}
			return
		}
		client.translator.ProcessTextTranslations(translation.GetFigmaSourceUrl(), translationChan, errorChan)
	}()

	moreTranslations := true
	imageReturned := false
	for moreTranslations || !imageReturned {
		select {
		case translationResult, done := <-translationChan:
			if done {
				go translation.UpsertNode(translationResult)

				sseClient.SendEvent("translation", func(w io.Writer) {
					template.TranslationNode(translationResult).Render(r.Context(), w)
				})
			} else {
				moreTranslations = false
			}
		case contextImageUrl := <-imageUrlChan:
			go translation.UpdateContextImage(contextImageUrl)

			sseClient.SendEvent("contextImage", func(w io.Writer) {
				template.TranslationContextImage(contextImageUrl).Render(r.Context(), w)
			})
			imageReturned = true
		case err := <-errorChan:
			fmt.Printf("Error generating translations: %v\n", err)
			moreTranslations = false
			imageReturned = true
		}
	}
}
