package translator

import (
	"translang/db"
	"translang/figma"
	"translang/openai"
)

type TranslatorClient struct {
	figmaClient  figma.FigmaClient
	openaiClient openai.OpenaiClient
	db           db.DBClient
}

func NewClient(figmaPAT string, openaiAPIKey string, dbClient db.DBClient) TranslatorClient {
	figmaClient := figma.NewClient(figmaPAT)
	openaiClient := openai.NewClient(openaiAPIKey)

	return TranslatorClient{
		figmaClient:  figmaClient,
		openaiClient: openaiClient,
		db:           dbClient,
	}
}
