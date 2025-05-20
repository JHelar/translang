package server

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"translang/persistence"
	"translang/translator"

	"github.com/gorilla/mux"
)

type ServerClient struct {
	translator  translator.TranslatorClient
	persistence persistence.PersistenceClient
	router      *mux.Router
	server      *http.Server
}

func NewClient(translator translator.TranslatorClient, persistence persistence.PersistenceClient) ServerClient {
	client := ServerClient{
		translator:  translator,
		persistence: persistence,
		router:      mux.NewRouter(),
	}

	client.router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	client.router.Use(routeVarsMiddleware)
	client.router.HandleFunc("/", client.HomeRoute).Name("home")
	client.router.HandleFunc("/translate", client.TranslationsRoute).Methods(http.MethodGet).Name("listTranslations")
	client.router.HandleFunc("/translate", client.CreateTranslationRoute).Methods(http.MethodPost).Name("createTranslation")
	client.router.HandleFunc("/translate/{id}", client.TranslationDetailsRoute).Methods(http.MethodGet).Name("getTranslation")
	client.router.HandleFunc("/translate/{id}", client.DeleteTranslationRoute).Methods(http.MethodDelete).Name("deleteTranslation")
	client.router.HandleFunc("/translate/{id}/stream", client.TranslateStreamRoute).Methods(http.MethodGet).Name("streamTranslation")

	client.router.HandleFunc("/node", client.NodesRoute).Methods(http.MethodGet).Name("listNodes")
	client.router.HandleFunc("/node/{id}", client.NodeDetailsRoute).Methods(http.MethodGet).Name("getNode")
	client.router.HandleFunc("/node/{id}/{language}", client.UpdateTranslationValue).Methods(http.MethodPatch).Name("updateNodeValue")

	client.server = &http.Server{
		Handler:      client.router,
		Addr:         "0.0.0.0:3000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return client
}

func (client ServerClient) ListenAndServe() {
	fmt.Printf("Listening on %s\n", client.server.Addr)
	if err := client.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
