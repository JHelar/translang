package figma

import (
	"log"
	"net/url"
	"strings"
)

type FigmaUrl struct {
	FileType string
	FileKey  string
	FileName string
	Query    url.Values
	RawUrl   string
}

func parseFigmaFileUrl(figmaFileUrl string) FigmaUrl {
	url, err := url.Parse(figmaFileUrl)
	if err != nil {
		log.Fatal("Failed to parse url", err)
	}

	pathParts := strings.Split(url.Path, "/")
	if len(pathParts) < 3+1 {
		log.Fatalf("Invalid figmaFileUrl: '%v'", figmaFileUrl)
	}

	return FigmaUrl{
		FileType: pathParts[1],
		FileKey:  pathParts[2],
		FileName: pathParts[3],
		Query:    url.Query(),
		RawUrl:   figmaFileUrl,
	}
}
