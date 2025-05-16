package server

import (
	"fmt"
	"io"
	"net/http"
	"translang/persistence"
	"translang/server/sse"
	"translang/template"
)

func (client ServerClient) TranslationsRoute(w http.ResponseWriter, r *http.Request) {
	translations, err := client.persistence.GetAllTranslations()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retreiving translation: %v\n", err), 500)
		return
	}

	var rows []template.TranslateRowProps
	for _, translation := range translations {
		contextImageUrl, _ := translation.GetContextImageUrl()
		nodes, _ := translation.GetAllNodes()
		detailsUrl, _ := client.router.Get("getTranslation").URL("id", translation.GetID())

		rows = append(rows, template.TranslateRowProps{
			ContextImageUrl:  contextImageUrl,
			FigmaSourceUrl:   translation.GetFigmaSourceUrl(),
			TranslationCount: fmt.Sprint(len(nodes)),
			DetailsUrl:       detailsUrl.String(),
		})
	}

	createTranslationUrl, _ := client.router.Get("createTranslation").URL()

	props := template.TranslateProps{
		Rows:                 rows,
		CreateTranslationUrl: createTranslationUrl.String(),
	}

	template.Translate(props).Render(r.Context(), w)
}

func (client ServerClient) CreateTranslationRoute(w http.ResponseWriter, r *http.Request) {
	figmaUrl := r.Form.Get("figmaUrl")
	if figmaUrl == "" {
		http.Error(w, "Missing figmaUrl", http.StatusBadRequest)
		return
	}

	translation, err := client.persistence.UpsertTranslation(figmaUrl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating translation: %v\n", err), http.StatusInternalServerError)
		return
	}

	sseURL, _ := client.router.Get("streamTranslation").URL("id", translation.GetID())
	props := template.TranslationModalProps{
		SSEUrl: sseURL.String(),
	}

	template.TranslationModal(props).Render(r.Context(), w)
}

func (client ServerClient) TranslationDetailsRoute(w http.ResponseWriter, r *http.Request) {
	translation, err := client.persistence.GetTranslationByID(r.Form.Get("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sseURL, _ := client.router.Get("streamTranslation").URL("id", translation.GetID())
	props := template.TranslationModalProps{
		SSEUrl: sseURL.String(),
	}

	template.TranslationModal(props).Render(r.Context(), w)
}

func (client ServerClient) DeleteTranslationRoute(w http.ResponseWriter, r *http.Request) {
	if err := client.persistence.DeleteTranslationByID(r.Form.Get("id")); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting translation: %v\n", err), 404)
		return
	}
}

func (client ServerClient) TranslateStreamRoute(w http.ResponseWriter, r *http.Request) {
	translation, err := client.persistence.GetTranslationByID(r.Form.Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting translation: %v\n", err), 404)
		return
	}

	sseClient := sse.NewClient(w, r)
	defer sseClient.Close()

	imageUrlChan, imageUrlErrorChan := client.translator.ProcessContextImage(translation)
	translationChan, translationErrorChan := client.translator.ProcessTextTranslations(translation)

	moreTranslations := true
	imageReturned := false
	for moreTranslations || !imageReturned {
		select {
		case translationResult := <-translationChan:
			// TODO: move upsert to translator
			var values []persistence.ValuePayload
			for _, value := range translationResult.Values {
				values = append(values, persistence.ValuePayload{
					Language: value.Language,
					Text:     value.Text,
				})
			}
			_, err := translation.UpsertNode(persistence.NodePayload{
				NodeId:  translationResult.NodeId,
				Source:  translationResult.Source,
				CopyKey: translationResult.CopyKey,
				Values:  values,
			})
			if err != nil {
				fmt.Print(err)
			}

			sseClient.SendEvent("translation", func(w io.Writer) {
				template.TranslationNode(translationResult).Render(r.Context(), w)
			})
		case contextImageUrl := <-imageUrlChan:
			// TODO: move upsert to translator
			err := translation.UpdateContextImage(contextImageUrl)
			if err != nil {
				fmt.Print(err)
			}

			sseClient.SendEvent("contextImage", func(w io.Writer) {
				template.TranslationContextImage(contextImageUrl).Render(r.Context(), w)
			})
			imageReturned = true
		case err := <-imageUrlErrorChan:
			fmt.Printf("Error generating context image: %v\n", err)
			moreTranslations = false
			imageReturned = true
		case err := <-translationErrorChan:
			fmt.Printf("Error generating translation: %v\n", err)
			moreTranslations = false
			imageReturned = true
		}
	}
}
