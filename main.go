package main

import (
	"log"
	"os"
	"strings"
	"translang/auth"
	"translang/figma"
	"translang/persistence/db"
	"translang/server"
	"translang/translator"

	_ "github.com/mattn/go-sqlite3"
)

type EnvVariables struct {
	FIGMA_PAT      string
	OPENAI_API_KEY string
	BASE_URL       string
}

func getEnvVariables() EnvVariables {
	bytes, err := os.ReadFile(".env")
	if err != nil {
		log.Fatal(err)
	}
	environmentsString := string(bytes)
	environmentsLines := strings.Split(environmentsString, "\n")

	var environments EnvVariables

	for _, line := range environmentsLines {
		keyVal := strings.Split(line, "=")
		if len(keyVal) != 2 {
			log.Fatal("Environment value is malformed")
		}
		if keyVal[0] == "FIGMA_PAT" {
			environments.FIGMA_PAT = strings.TrimSpace(keyVal[1])
		} else if keyVal[0] == "OPENAI_API_KEY" {
			environments.OPENAI_API_KEY = strings.TrimSpace(keyVal[1])
		} else if keyVal[0] == "BASE_URL" {
			environments.BASE_URL = strings.TrimSpace(keyVal[1])
		}
	}

	if environments.FIGMA_PAT == "" {
		log.Fatalf("Missing FIGMA_PAT variable")
	}

	if environments.OPENAI_API_KEY == "" {
		log.Fatalf("Missing OPENAI_API_KEY variable")
	}

	if environments.BASE_URL == "" {
		log.Fatalf("Missing BASE_URL variable")
	}

	return environments
}

func main() {
	env := getEnvVariables()

	authProvider := auth.NewAuthProvider()
	dbPersistenceClient := db.NewClient()

	passwordProvider := auth.NewPasswordProvider(&dbPersistenceClient.DBClient)
	authProvider.AddProvider(passwordProvider)

	figmaClient := figma.NewClient(env.FIGMA_PAT)
	translator := translator.NewClient(env.FIGMA_PAT, env.OPENAI_API_KEY, dbPersistenceClient)
	serverClient := server.NewClient(translator, dbPersistenceClient, figmaClient, env.BASE_URL)

	serverClient.ListenAndServe()
}
