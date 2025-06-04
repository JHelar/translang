package figma

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	WebhookPingEventType              = "PING"
	WebhookFileUpdateEventType        = "FILE_UPDATE"
	WebhookFileVersionUpdateEventType = "FILE_VERSION_UPDATE"
	WebhookFileDeleteEventType        = "FILE_DELETE"
	WebhookLibraryPublishEventType    = "LIBRARY_PUBLISH"
	WebhookFileCommentEventType       = "FILE_COMMENT"
)

const (
	DemoTeamID = "1354746843273557935"
)

type WebhookSetupPayload struct {
	EventType string `json:"event_type"`
	Context   string `json:"context"`
	ContextID string `json:"context_id"`
	Endpoint  string `json:"endpoint"`
	Passcode  string `json:"passcode"`
}

func (client FigmaClient) SetupWebhook(payload WebhookSetupPayload) error {
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	writer := bytes.NewBuffer(jsonBytes)
	request := client.request("/v2/webhooks", http.MethodPost, writer)
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to setup webhook (status %d): %s", response.StatusCode, body)
	}

	return nil
}
