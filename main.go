package main

import (
	"log"
	"os"
	"translang/figma"
	"translang/openai"
	"translang/server"
)

func getTokens() (string, string) {
	figmaPAT := os.Getenv("FIGMA_PAT")
	if figmaPAT == "" {
		log.Fatalf("Missing FIGMA_PAT variable")
	}

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatalf("Missing OPENAI_API_KEY variable")
	}

	return figmaPAT, openaiAPIKey
}

type ProcessResult struct {
	ContextImageUrl string                `json:"contextImageUrl"`
	Translations    []openai.Translations `json:"translations"`
}

func Process(figmaUrl string, figmaClient *figma.FigmaClient, openaiClient *openai.OpenaiClient) ProcessResult {
	imageUrlChan := make(chan string)
	translationsChan := make(chan []openai.Translations)

	go func(figmaUrl string) {
		imageUrlChan <- figmaClient.GetImage(figmaUrl)
	}(figmaUrl)

	go func(figmaUrl string) {
		node := figmaClient.GetFileNodes(figmaUrl)
		textNodes := node.FindAllNodesOfType("TEXT")

		var translations []openai.Translations
		for _, textNode := range textNodes {
			translation := openaiClient.Translate(textNode.Characters)
			translations = append(translations, translation)
		}

		translationsChan <- translations
	}(figmaUrl)

	return ProcessResult{
		ContextImageUrl: <-imageUrlChan,
		Translations:    <-translationsChan,
	}
}

func main() {
	// figmaPAT, openaiAPIKey := getTokens()
	// figmaClient := figma.Client(figmaPAT)
	// openaiClient := openai.Client(openaiAPIKey)

	// if len(os.Args) < 2 {
	// 	log.Fatalf("Missing figmaUrl")
	// }

	// figmaUrl := os.Args[1]

	// result := Process(figmaUrl, &figmaClient, &openaiClient)

	// resultJSON, err := json.Marshal(result)
	// if err != nil {
	// 	log.Fatal("Failed to stringify", err)
	// }

	// fmt.Print(string(resultJSON))

	server.ListenAndServe()
}
