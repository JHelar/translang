package translator

import (
	"translang/persistence"
)

type TranslationValue struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}

type TranslationResult struct {
	NodeId  string             `json:"nodeId"`
	Source  string             `json:"source"`
	CopyKey string             `json:"copyKey"`
	Values  []TranslationValue `json:"values"`
}

func nodeToTranslationResult(node persistence.PersistenceNode) (TranslationResult, error) {
	payload, err := node.ToPayload()
	if err != nil {
		return TranslationResult{}, err
	}

	result := TranslationResult{
		NodeId:  payload.NodeId,
		Source:  payload.Source,
		CopyKey: payload.CopyKey,
	}

	for _, value := range payload.Values {
		result.Values = append(result.Values, TranslationValue{
			Language: value.Language,
			Text:     value.Text,
		})
	}

	return result, nil
}

func translationResultToNodePayload(result TranslationResult) persistence.NodePayload {
	var values []persistence.ValuePayload
	for _, value := range result.Values {
		values = append(values, persistence.ValuePayload{
			Language: value.Language,
			Text:     value.Text,
		})
	}

	return persistence.NodePayload{
		NodeId:  result.NodeId,
		Source:  result.Source,
		CopyKey: result.CopyKey,
		Values:  values,
	}
}

type ProcessResult struct {
	FigmaSourceUrl  string              `json:"figmaSourceUrl"`
	ContextImageUrl string              `json:"contextImageUrl"`
	Translations    []TranslationResult `json:"translations"`
}

func (client TranslatorClient) ProcessContextImage(translation persistence.PersistenceTranslation) (<-chan string, <-chan error) {
	imageUrlChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		imageUrl, err := translation.GetContextImageUrl()
		if err == nil {
			imageUrlChan <- imageUrl
			return
		}

		imageUrl, err = client.figmaClient.GetImage(translation.GetFigmaSourceUrl())
		if err != nil {
			errorChan <- err
			return
		}

		imageUrlChan <- imageUrl
		if err := translation.UpdateContextImage(imageUrl); err != nil {
			errorChan <- err
			return
		}
	}()

	return imageUrlChan, errorChan
}

func (client TranslatorClient) ProcessTextTranslations(translation persistence.PersistenceTranslation) (<-chan TranslationResult, <-chan error) {
	translationResult := make(chan TranslationResult)
	errorChan := make(chan error)

	go func() {
		nodes, err := translation.GetAllNodes()
		defer close(translationResult)

		if err == nil && len(nodes) > 0 {
			for _, node := range nodes {
				result, err := nodeToTranslationResult(node)
				if err != nil {
					errorChan <- err
					return
				}

				translationResult <- result
			}
		}

		node, err := client.figmaClient.GetFileNodes(translation.GetFigmaSourceUrl())
		if err != nil {
			errorChan <- err
			return
		}

		textNodes := node.FindAllNodesOfType("TEXT")
		for _, textNode := range textNodes {
			node, err := client.persistence.GetNodeFromSourceText(textNode.Characters)
			var result TranslationResult

			if err == nil {
				payload, _ := node.ToPayload()
				result.NodeId = payload.NodeId
				result.Source = payload.Source
				result.CopyKey = payload.CopyKey

				for _, value := range payload.Values {
					result.Values = append(result.Values, TranslationValue{
						Language: value.Language,
						Text:     value.Text,
					})
				}
			} else {
				translation := client.openaiClient.Translate(textNode.Characters)
				result.NodeId = textNode.ID
				result.Source = translation.Source
				result.CopyKey = translation.CopyKey
				result.Values = []TranslationValue{
					{
						Language: "sv",
						Text:     translation.Swedish,
					},
					{
						Language: "en",
						Text:     translation.English,
					},
					{
						Language: "fi",
						Text:     translation.Finnish,
					},
				}
			}

			translationResult <- result
			if _, err := translation.UpsertNode(translationResultToNodePayload(result)); err != nil {
				errorChan <- err
			}
		}
	}()

	return translationResult, errorChan
}

func (client TranslatorClient) Process(figmaUrl string) (ProcessResult, error) {
	translation, err := client.persistence.UpsertTranslation(figmaUrl)
	if err != nil {
		return ProcessResult{}, err
	}
	imageUrlChan, imageErrorChan := client.ProcessContextImage(translation)
	translationChan, translationErrorChan := client.ProcessTextTranslations(translation)

	var contextImageUrl string
	var translations []TranslationResult

	moreTranslations := true
	for moreTranslations {
		select {
		case translation, moreTranslations := <-translationChan:
			if moreTranslations {
				translations = append(translations, translation)
			}
		case contextImageUrl = <-imageUrlChan:
		case err := <-imageErrorChan:
			return ProcessResult{}, err
		case err := <-translationErrorChan:
			return ProcessResult{}, err
		}
	}

	return ProcessResult{
		FigmaSourceUrl:  figmaUrl,
		ContextImageUrl: contextImageUrl,
		Translations:    translations,
	}, nil
}
