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

func (client ServerClient) TranslationsRoute(w http.ResponseWriter, r *http.Request) {
	translations, err := dto.GetAllTranslations(client.db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retreiving translation: %v\n", err), 500)
		return
	}

	var rows []template.TranslateRowProps
	for _, translation := range translations {
		contextImageUrl := translation.ContextImageUrl.String
		nodes, _ := translation.Nodes(&client.db)
		detailsUrl, _ := client.router.Get("getTranslation").URL("id", strconv.FormatInt(translation.ID, 10))

		rows = append(rows, template.TranslateRowProps{
			ContextImageUrl:  contextImageUrl,
			FigmaSourceUrl:   translation.FigmaSourceUrl,
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

	translation, err := dto.UpsertTranslation(figmaUrl, client.db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating translation: %v\n", err), http.StatusInternalServerError)
		return
	}

	sseURL, _ := client.router.Get("streamTranslation").URL("id", strconv.FormatInt(translation.ID, 10))
	props := template.TranslationModalProps{
		SSEUrl: sseURL.String(),
	}

	template.TranslationModal(props).Render(r.Context(), w)
}

func (client ServerClient) TranslationDetailsRoute(w http.ResponseWriter, r *http.Request) {
	translationID, err := strconv.ParseInt(r.Form.Get("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	translation, err := dto.GetTranslationByID(translationID, client.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sseURL, _ := client.router.Get("streamTranslation").URL("id", strconv.FormatInt(translation.ID, 10))
	props := template.TranslationModalProps{
		SSEUrl: sseURL.String(),
	}

	template.TranslationModal(props).Render(r.Context(), w)
}

func (client ServerClient) DeleteTranslationRoute(w http.ResponseWriter, r *http.Request) {
	translationID, err := strconv.ParseInt(r.Form.Get("id"), 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting translation: %v\n", err), 404)
		return
	}

	if err := dto.DeleteTranslation(translationID, client.db); err != nil {
		http.Error(w, fmt.Sprintf("Error deleting translation: %v\n", err), 404)
		return
	}
}

func (client ServerClient) TranslateStreamRoute(w http.ResponseWriter, r *http.Request) {
	translationID, err := strconv.ParseInt(r.Form.Get("id"), 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing translation id: %v\n", err), http.StatusBadRequest)
		return
	}

	translation, err := dto.GetTranslationByID(translationID, client.db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting translation: %v\n", err), http.StatusNotFound)
		return
	}

	sseClient := sse.NewClient(w, r)
	defer sseClient.Close()

	imageUrlChan, imageUrlErrorChan := client.translator.ProcessContextImage(translation)
	translationChan, translationErrorChan := client.translator.ProcessTextTranslations(translation)

	hasSentTranslations := false
	hasSentImage := false
	for !hasSentTranslations || !hasSentImage {
		if !hasSentTranslations {
			select {
			case translationResult, ok := <-translationChan:
				if ok {
					sseClient.SendEvent("translation", func(w io.Writer) {
						props := template.TranslationNodeProps{
							TranslationResult: translationResult,
						}
						for _, value := range translationResult.Values {
							updateValueURL, err := client.router.Get("updateNodeValue").URL("id", translationResult.ID, "language", value.Language)
							if err != nil {
								fmt.Printf("Error generating URL: %s\n", err.Error())
							}

							props.Values = append(props.Values, struct {
								translator.TranslationValue
								UpdateValueURL string
							}{
								TranslationValue: value,
								UpdateValueURL:   updateValueURL.String(),
							})
						}
						template.TranslationNode(props).Render(r.Context(), w)
					})
				} else {
					hasSentTranslations = true
				}
			default:
			}
		}

		if !hasSentImage {
			select {
			case contextImageUrl, ok := <-imageUrlChan:
				if ok {
					sseClient.SendEvent("contextImage", func(w io.Writer) {
						template.TranslationContextImage(contextImageUrl).Render(r.Context(), w)
					})
				} else {
					hasSentImage = true
				}
			default:
			}
		}

		select {
		case imageError := <-imageUrlErrorChan:
			fmt.Printf("Error with generating image: %s", imageError.Error())
			return
		case translationError := <-translationErrorChan:
			fmt.Printf("Error with generating translation: %s", translationError.Error())
			return
		default:
		}
	}
}
