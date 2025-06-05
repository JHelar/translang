package translator

import (
	"strconv"
	"translang/db"
	"translang/dto"
)

type TranslationValue struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}

type TranslationResult struct {
	ID      string             `json:"id"`
	NodeId  string             `json:"nodeId"`
	Source  string             `json:"source"`
	CopyKey string             `json:"copyKey"`
	Values  []TranslationValue `json:"values"`
}

func nodeToTranslationResult(translation dto.Translation, node dto.TranslationNode, db db.DBClient) (TranslationResult, error) {
	values, err := node.Values(&db)
	if err != nil {
		return TranslationResult{}, err
	}

	result := TranslationResult{
		ID:      strconv.FormatInt(node.ID, 10),
		Source:  node.SourceText,
		CopyKey: node.CopyKey,
	}

	for _, value := range values {
		result.Values = append(result.Values, TranslationValue{
			Language: value.CopyLanguage,
			Text:     value.CopyText,
		})
	}

	return result, nil
}

func translationResultToNodePayload(result TranslationResult) dto.TranslationNode {
	return dto.TranslationNode{
		SourceText: result.Source,
		CopyKey:    result.CopyKey,
	}
}

type ProcessResult struct {
	FigmaSourceUrl  string              `json:"figmaSourceUrl"`
	ContextImageUrl string              `json:"contextImageUrl"`
	Translations    []TranslationResult `json:"translations"`
}

func (client TranslatorClient) ProcessContextImage(translation dto.Translation) (<-chan string, <-chan error) {
	imageUrlChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		if translation.ContextImageUrl.Valid && translation.ContextImageUrl.String != "" {
			imageUrlChan <- translation.ContextImageUrl.String
			return
		}

		imageUrl, err := client.figmaClient.GetImage(translation.FigmaSourceUrl)
		if err != nil {
			errorChan <- err
			return
		}

		imageUrlChan <- imageUrl
		if err := translation.UpdateContextImage(imageUrl, &client.db); err != nil {
			errorChan <- err
			return
		}
	}()

	return imageUrlChan, errorChan
}

func (client TranslatorClient) ProcessTextTranslations(translation dto.Translation) (<-chan TranslationResult, <-chan error) {
	translationResult := make(chan TranslationResult)
	errorChan := make(chan error)

	go func() {
		nodes, err := translation.Nodes(&client.db)
		defer close(translationResult)

		if err == nil && len(nodes) > 0 {
			for _, node := range nodes {
				result, err := nodeToTranslationResult(translation, node, client.db)
				if err != nil {
					errorChan <- err
					return
				}
				translationResult <- result
			}
			return
		}

		node, err := client.figmaClient.GetFileNodes(translation.FigmaSourceUrl)
		if err != nil {
			errorChan <- err
			return
		}

		textNodes := node.FindAllNodesOfType("TEXT")
		for _, textNode := range textNodes {
			node, err := dto.GetTranslationNodeBySourceText(textNode.Characters, &client.db)
			if err == nil {
				translation := client.openaiClient.Translate(textNode.Characters)
				node.SourceText = translation.Source
				node.CopyKey = translation.CopyKey
			}

			node, err = translation.UpsertNode(textNode.ID, node.SourceText, node.CopyKey, &client.db)
			if err != nil {
				errorChan <- err
				return
			}

			values, err := node.Values(&client.db)
			if err != nil {
				errorChan <- err
				return
			}

			result := TranslationResult{
				ID:      strconv.FormatInt(node.ID, 10),
				NodeId:  textNode.ID,
				Source:  node.SourceText,
				CopyKey: node.CopyKey,
			}

			for _, value := range values {
				result.Values = append(result.Values, TranslationValue{
					Language: value.CopyLanguage,
					Text:     value.CopyText,
				})
			}
			translationResult <- result
		}
	}()

	return translationResult, errorChan
}

func (client TranslatorClient) Process(figmaUrl string) (ProcessResult, error) {
	translation, err := dto.UpsertTranslation(figmaUrl, client.db)
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
