package server

import (
	"fmt"
	"net/http"
	"translang/db"
	"translang/translator"
)

type ServerClient struct {
	translator translator.TranslatorClient
	db         db.DBClient
}

func NewClient(translator translator.TranslatorClient, db db.DBClient) ServerClient {
	return ServerClient{
		translator: translator,
		db:         db,
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
