package figma

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

const API_PATH = "https://api.figma.com"

type FigmaClient struct {
	token string
}

func NewClient(figmaPAT string) FigmaClient {
	return FigmaClient{
		token: figmaPAT,
	}
}

func (client *FigmaClient) request(path string, method string, body io.Reader) *http.Request {
	apiPath := fmt.Sprintf("%v%v", API_PATH, path)
	req, err := http.NewRequest(method, apiPath, body)

	if err != nil {
		log.Fatalf("Failed to create request: '%v' %v", path, err)
	}

	req.Header.Add("X-Figma-Token", client.token)

	return req
}
