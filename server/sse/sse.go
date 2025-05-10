package sse

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type SSEEventClient struct {
	sb           strings.Builder
	w            http.ResponseWriter
	rc           http.ResponseController
	clientClosed <-chan struct{}
	ticker       *time.Ticker
}

func NewClient(w http.ResponseWriter, r *http.Request) SSEEventClient {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(time.Millisecond * 500)

	return SSEEventClient{
		w:            w,
		sb:           strings.Builder{},
		rc:           *http.NewResponseController(w),
		clientClosed: r.Context().Done(),
		ticker:       ticker,
	}
}

func (client SSEEventClient) SendEvent(eventName string, writeData func(w io.Writer)) {
	defer client.sb.Reset()

	select {
	case <-client.clientClosed:
		return
	case <-client.ticker.C:
		client.sb.WriteString(fmt.Sprintf("event:%s\ndata:", eventName))
		writeData(&client.sb)
		client.sb.WriteString("\n\n")

		_, err := fmt.Fprint(client.w, client.sb.String())
		if err != nil {
			return
		}
		err = client.rc.Flush()
		if err != nil {
			return
		}
	}
}

func (client SSEEventClient) Close() {
	client.SendEvent("close", func(w io.Writer) {})
}
