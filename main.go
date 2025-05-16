package main

import (
	"log"
	"os"
	"strings"
	"translang/persistence/db"
	"translang/server"
	"translang/translator"

	_ "github.com/mattn/go-sqlite3"
)

func getTokens() (string, string) {
	bytes, err := os.ReadFile(".env")
	if err != nil {
		log.Fatal(err)
	}
	environmentsString := string(bytes)
	environmentsLines := strings.Split(environmentsString, "\n")

	var environments struct {
		FIGMA_PAT      string
		OPENAI_API_KEY string
	}

	for _, line := range environmentsLines {
		keyVal := strings.Split(line, "=")
		if len(keyVal) != 2 {
			log.Fatal("Environment value is malformed")
		}
		if keyVal[0] == "FIGMA_PAT" {
			environments.FIGMA_PAT = strings.TrimSpace(keyVal[1])
		} else if keyVal[0] == "OPENAI_API_KEY" {
			environments.OPENAI_API_KEY = strings.TrimSpace(keyVal[1])
		}
	}

	if environments.FIGMA_PAT == "" {
		log.Fatalf("Missing FIGMA_PAT variable")
	}

	if environments.OPENAI_API_KEY == "" {
		log.Fatalf("Missing OPENAI_API_KEY variable")
	}

	return environments.FIGMA_PAT, environments.OPENAI_API_KEY
}

func main() {
	figmaPAT, openaiAPIKey := getTokens()
	dbPersistenceClient := db.NewClient()
	translator := translator.NewClient(figmaPAT, openaiAPIKey, dbPersistenceClient)
	serverClient := server.NewClient(translator, dbPersistenceClient)

	serverClient.ListenAndServe()
}
