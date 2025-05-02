package dto

import (
	"fmt"
	"translang/db"
	"translang/translator"
)

func (translation *Translation) ToResult(client db.DBClient) (translator.ProcessResult, error) {
	result := translator.ProcessResult{
		FigmaSourceUrl:  translation.FigmaSourceUrl,
		ContextImageUrl: translation.ContextImageUrl.String,
		Translations:    []translator.TranslationResult{},
	}

	nodes, err := translation.Nodes(client)
	if err != nil {
		return translator.ProcessResult{}, err
	}

	for _, node := range nodes {
		translationResult, err := node.ToResult(client)
		if err != nil {
			return translator.ProcessResult{}, err
		}

		result.Translations = append(result.Translations, translationResult)
	}

	return result, nil
}

func (node *TranslationNode) ToResult(client db.DBClient) (translator.TranslationResult, error) {
	result := translator.TranslationResult{
		NodeId:  node.FigmaTextNodeId,
		Source:  node.SourceText,
		CopyKey: node.CopyKey,
	}

	values, err := node.Values(client)
	if err != nil {
		return translator.TranslationResult{}, err
	}

	if len(values) == 0 {
		return translator.TranslationResult{}, fmt.Errorf("missing node values")
	}

	for _, value := range values {
		result.Values = append(result.Values, translator.TranslationValue{
			Language: value.CopyLanguage,
			Text:     value.CopyText,
		})
	}

	return result, nil
}
