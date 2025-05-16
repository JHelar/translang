package translator

import (
	"translang/figma"
	"translang/openai"
	"translang/persistence"
	"translang/persistence/db"
)

type TranslatorClient struct {
	figmaClient  figma.FigmaClient
	openaiClient openai.OpenaiClient
	persistence  persistence.PersistenceClient
}

func NewClient(figmaPAT string, openaiAPIKey string) TranslatorClient {
	figmaClient := figma.NewClient(figmaPAT)
	openaiClient := openai.NewClient(openaiAPIKey)
	persistenceClient := db.NewClient()

	return TranslatorClient{
		figmaClient:  figmaClient,
		openaiClient: openaiClient,
		persistence:  persistenceClient,
	}
}
