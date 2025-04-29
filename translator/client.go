package translator

import (
	"translang/figma"
	"translang/openai"
)

type TranslatorClient struct {
	figmaClient  figma.FigmaClient
	openaiClient openai.OpenaiClient
}

func NewClient(figmaPAT string, openaiAPIKey string) TranslatorClient {
	figmaClient := figma.NewClient(figmaPAT)
	openaiClient := openai.NewClient(openaiAPIKey)

	return TranslatorClient{
		figmaClient:  figmaClient,
		openaiClient: openaiClient,
	}
}
