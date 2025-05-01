package dto

import (
	"fmt"
	"translang/db"
	"translang/translator"
)

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
