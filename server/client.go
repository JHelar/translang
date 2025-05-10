package server

import (
	"fmt"
	"net/http"
	"translang/persistence"
	"translang/translator"
)

type ServerClient struct {
	translator  translator.TranslatorClient
	persistence persistence.Persistence
}

func NewClient(translator translator.TranslatorClient, persistence persistence.Persistence) ServerClient {
	return ServerClient{
		translator:  translator,
		persistence: persistence,
	}
}

func (client ServerClient) ListenAndServe() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))
	http.HandleFunc("/", client.HomeRoute)
	http.HandleFunc("/translate", client.TranslateRoute)
	http.HandleFunc("/translate/stream", client.TranslateStreamRoute)

	fmt.Println("Listening on :3000")
	http.ListenAndServe("0.0.0.0:3000", nil)
}
