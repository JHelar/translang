package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"translang/dto"
	"translang/server/sse"
	"translang/template"
	"translang/translator"
)

func (client ServerClient) TranslateRoute(w http.ResponseWriter, r *http.Request) {
	translations, err := dto.GetAllTranslations(client.db)
	if err != nil {
		fmt.Printf("Error retreiving translation: %v\n", err)
		w.WriteHeader(500)
		return
	}

	results := []translator.ProcessResult{}
	for _, translation := range translations {
		result, err := translation.ToResult(client.db)
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
		// if r.Form.Get("delete") == "true" {
		// 	if err := dto.DeleteTranslation(figmaUrl, client.db); err != nil {
		// 		fmt.Printf("Error deleting translation: %v\n", err)
		// 	}
		// } else {

		// }
		translation, err := dto.UpsertTranslation(figmaUrl, client.db)
		if err != nil {
			fmt.Printf("Error creating translation: %v\n", err)
		} else {
			props.Modal = template.TranslationModalProps{
				SSEUrl:         fmt.Sprintf("/translate/stream?id=%d", translation.ID),
				FigmaSourceUrl: figmaUrl,
			}
		}
	}

	template.Translate(props).Render(r.Context(), w)

}

func (client ServerClient) TranslateStreamRoute(w http.ResponseWriter, r *http.Request) {
	translationID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		w.WriteHeader(422)
		return
	}

	translation, err := dto.GetTranslationByID(translationID, client.db)
	if err != nil {
		fmt.Printf("Error getting translation (%d): %v\n", translationID, err)
		w.WriteHeader(404)
		return
	}
	sseClient := sse.NewClient(w)
	defer sseClient.Close()

	imageUrlChan := make(chan string)
	translationChan := make(chan translator.TranslationResult)
	errorChan := make(chan error)

	go func() {
		if translation.ContextImageUrl.Valid && translation.ContextImageUrl.String != "" {
			imageUrlChan <- translation.ContextImageUrl.String
		} else {
			client.translator.ProcessContextImage(translation.FigmaSourceUrl, imageUrlChan, errorChan)
		}
	}()

	go func() {
		defer close(translationChan)

		nodes, err := translation.Nodes(client.db)
		if err != nil {
			errorChan <- err
			return
		}
		if len(nodes) > 0 {
			for _, node := range nodes {
				result, err := node.ToResult(client.db)
				if err != nil {
					errorChan <- err
					return
				}
				translationChan <- result
			}
			return
		}
		client.translator.ProcessTextTranslations(translation.FigmaSourceUrl, translationChan, errorChan)
	}()

	moreTranslations := true
	imageReturned := false
	for moreTranslations || !imageReturned {
		select {
		case translationResult, done := <-translationChan:
			if done {
				go func() {
					node, err := translation.UpsertNode(translationResult.NodeId, translationResult.Source, translationResult.CopyKey, client.db)
					if err != nil {
						errorChan <- err
						return
					}
					for _, value := range translationResult.Values {
						if _, err = node.UpsertValue(value.Language, value.Text, client.db); err != nil {
							errorChan <- err
							return
						}
					}
				}()

				sseClient.SendEvent("translation", func(w io.Writer) {
					template.TranslationNode(translationResult).Render(r.Context(), w)
				})
			} else {
				moreTranslations = false
			}
		case contextImageUrl := <-imageUrlChan:
			go translation.UpdateContextImage(contextImageUrl, client.db)

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
