package translator

import (
	"translang/figma"
	"translang/openai"
)

type TranslatorClient struct {
	figmaClient  figma.FigmaClient
	openaiClient openai.OpenaiClient
}

func Client(figmaPAT string, openaiAPIKey string) TranslatorClient {
	figmaClient := figma.Client(figmaPAT)
	openaiClient := openai.Client(openaiAPIKey)

	return TranslatorClient{
		figmaClient:  figmaClient,
		openaiClient: openaiClient,
	}
}
