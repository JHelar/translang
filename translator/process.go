package translator

import "translang/openai"

type TranslationResult struct {
	NodeId      string             `json:"nodeId"`
	Translation openai.Translation `json:"translation"`
}

type ProcessResult struct {
	ContextImageUrl string              `json:"contextImageUrl"`
	Translations    []TranslationResult `json:"translations"`
}

func (client TranslatorClient) Process(figmaUrl string) (ProcessResult, error) {
	imageUrlChan := make(chan string)
	translationsChan := make(chan []TranslationResult)
	errorChan := make(chan error)

	go func(figmaUrl string) {
		imageUrl, err := client.figmaClient.GetImage(figmaUrl)
		if err != nil {
			errorChan <- err
		} else {
			imageUrlChan <- imageUrl
		}
	}(figmaUrl)

	go func(figmaUrl string) {
		node, err := client.figmaClient.GetFileNodes(figmaUrl)
		if err != nil {
			errorChan <- err
			return
		}

		textNodes := node.FindAllNodesOfType("TEXT")

		var translations []TranslationResult
		for _, textNode := range textNodes {
			translation := client.openaiClient.Translate(textNode.Characters)
			translations = append(translations, TranslationResult{
				NodeId:      textNode.ID,
				Translation: translation,
			})
		}

		translationsChan <- translations
	}(figmaUrl)

	var contextImageUrl string
	var translations []TranslationResult

	for range 2 {
		select {
		case translations = <-translationsChan:
		case contextImageUrl = <-imageUrlChan:
		case err := <-errorChan:
			return ProcessResult{}, err
		}
	}

	return ProcessResult{
		ContextImageUrl: contextImageUrl,
		Translations:    translations,
	}, nil
}
