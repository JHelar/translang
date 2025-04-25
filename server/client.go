package server

import (
	"fmt"
	"net/http"
	"translang/translator"
)

type ServerClient struct {
	translator translator.TranslatorClient
}

func Client(translator translator.TranslatorClient) ServerClient {
	return ServerClient{
		translator: translator,
	}
}

func (client ServerClient) ListenAndServe() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))
	http.HandleFunc("/", client.HomeRoute)
	http.HandleFunc("/translate", client.TranslateRoute)
	http.HandleFunc("/translation", client.TranslationRoute)
	http.HandleFunc("/translation/stream", client.TranslationStream)

	fmt.Println("Listening on :3000")
	http.ListenAndServe("0.0.0.0:3000", nil)
}
