package sse

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SSEEventClient struct {
	sb strings.Builder
	w  http.ResponseWriter
}

func NewClient(w http.ResponseWriter) SSEEventClient {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	return SSEEventClient{
		w:  w,
		sb: strings.Builder{},
	}
}

func (client SSEEventClient) SendEvent(eventName string, writeData func(w io.Writer)) {
	defer client.sb.Reset()

	client.sb.WriteString(fmt.Sprintf("event:%s\ndata:", eventName))
	writeData(&client.sb)
	client.sb.WriteString("\n\n")

	client.w.Write([]byte(client.sb.String()))
	client.w.(http.Flusher).Flush()
}

func (client SSEEventClient) Close() {
	client.SendEvent("close", func(w io.Writer) {})
}
