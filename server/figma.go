package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type EventPayload struct {
	EventType string `json:"event_type"`
}

func (client ServerClient) HandleFigmaEvent(w http.ResponseWriter, r *http.Request) {
	var payload EventPayload
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&payload); err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("Got Figma event: '%s'\n", payload.EventType)

	w.WriteHeader(http.StatusOK)
}
