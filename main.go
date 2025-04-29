package main

import (
	"log"
	"os"
	"translang/server"
	"translang/translator"

	_ "github.com/mattn/go-sqlite3"
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

func main() {
	figmaPAT, openaiAPIKey := getTokens()
	// figmaClient := figma.Client(figmaPAT)
	// openaiClient := openai.Client(openaiAPIKey)

	// if len(os.Args) < 2 {
	// 	log.Fatalf("Missing figmaUrl")
	// }

	// figmaUrl := os.Args[1]
	// fmt.Printf("Fetching translations for url: %v\n", figmaUrl)
	// result := Process(figmaUrl, &figmaClient, &openaiClient)

	// resultJSON, err := json.Marshal(result)
	// if err != nil {
	// 	log.Fatal("Failed to stringify", err)
	// }

	// fmt.Print(string(resultJSON))

	// db, err := sql.Open("sqlite3", "./foo.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	// sqlStmt := `
	// create table foo (id integer not null primary key, name text);
	// delete from foo;
	// `

	// _, err = db.Exec(sqlStmt)
	// if err != nil {
	// 	log.Printf("%q: %s\n", err, sqlStmt)
	// 	return
	// }
	translator := translator.Client(figmaPAT, openaiAPIKey)
	serverClient := server.NewClient(translator)

	serverClient.ListenAndServe()
}
