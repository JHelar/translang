package translator

import "translang/persistence"

type TranslationValue struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}

type TranslationResult struct {
	NodeId  string             `json:"nodeId"`
	Source  string             `json:"source"`
	CopyKey string             `json:"copuKey"`
	Values  []TranslationValue `json:"values"`
}

type ProcessResult struct {
	FigmaSourceUrl  string              `json:"figmaSourceUrl"`
	ContextImageUrl string              `json:"contextImageUrl"`
	Translations    []TranslationResult `json:"translations"`
}

func (client TranslatorClient) ProcessContextImage(figmaUrl string, imageUrlChan chan<- string, errorChan chan<- error) {
	imageUrl, err := client.figmaClient.GetImage(figmaUrl)
	if err != nil {
		errorChan <- err
	} else {
		imageUrlChan <- imageUrl
	}
}

func (client TranslatorClient) ProcessTextTranslations(figmaUrl string, translationResult chan<- TranslationResult, errorChan chan<- error, persistenceClient persistence.PersistenceClient) {
	node, err := client.figmaClient.GetFileNodes(figmaUrl)
	if err != nil {
		errorChan <- err
		return
	}

	textNodes := node.FindAllNodesOfType("TEXT")
	for _, textNode := range textNodes {
		node, err := persistenceClient.GetNodeFromSourceText(textNode.Characters)
		if err == nil {
			payload, _ := node.ToPayload()
			var values []TranslationValue
			for _, value := range payload.Values {
				values = append(values, TranslationValue{
					Language: value.Language,
					Text:     value.Text,
				})
			}
			translationResult <- TranslationResult{
				NodeId:  payload.NodeId,
				Source:  payload.Source,
				CopyKey: payload.CopyKey,
				Values:  values,
			}
			continue
		}
		translation := client.openaiClient.Translate(textNode.Characters)
		translationResult <- TranslationResult{
			NodeId:  textNode.ID,
			Source:  translation.Source,
			CopyKey: translation.CopyKey,
			Values: []TranslationValue{
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
			},
		}
	}
}

func (client TranslatorClient) Process(figmaUrl string) (ProcessResult, error) {
	imageUrlChan := make(chan string)
	translationChan := make(chan TranslationResult)
	errorChan := make(chan error)

	go client.ProcessContextImage(figmaUrl, imageUrlChan, errorChan)
	go client.ProcessTextTranslations(figmaUrl, translationChan, errorChan, nil)

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
		case err := <-errorChan:
			return ProcessResult{}, err
		}
	}

	return ProcessResult{
		FigmaSourceUrl:  figmaUrl,
		ContextImageUrl: contextImageUrl,
		Translations:    translations,
	}, nil
}
