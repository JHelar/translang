package translator

import (
	"translang/figma"
	"translang/openai"
	"translang/persistence"
)

type TranslatorClient struct {
	figmaClient  figma.FigmaClient
	openaiClient openai.OpenaiClient
	persistence  persistence.PersistenceClient
}

func NewClient(figmaPAT string, openaiAPIKey string, persistence persistence.PersistenceClient) TranslatorClient {
	figmaClient := figma.NewClient(figmaPAT)
	openaiClient := openai.NewClient(openaiAPIKey)

	return TranslatorClient{
		figmaClient:  figmaClient,
		openaiClient: openaiClient,
		persistence:  persistence,
	}
}
