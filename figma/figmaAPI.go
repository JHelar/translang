package figma

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const API_PATH = "https://api.figma.com"

type FigmaClient struct {
	token string
}

type FigmaUrl struct {
	FileType string
	FileKey  string
	FileName string
	Query    url.Values
	rawUrl   url.URL
}

type FigmaNode struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Characters string      `json:"characters,omitempty"`
	Children   []FigmaNode `json:"children"`
}

type FigmaNodePath struct {
	FigmaNode
	Path string
}

type FigmaDocumentNode struct {
	Document FigmaNode `json:"document"`
}

type FigmaNodeFile struct {
	Name         string                       `json:"name"`
	ThumbnailUrl string                       `json:"thumbnailUrl"`
	Nodes        map[string]FigmaDocumentNode `json:"nodes"`
}

type FigmaImageResponse struct {
	Error  string            `json:"err,string"`
	Images map[string]string `json:"images"`
}

func (node FigmaNodePath) String() string {
	return fmt.Sprintf("%v (%v): '%v'", node.Path, node.Type, node.Characters)
}

func (rootNode *FigmaNode) FindAllNodesOfType(nodeType string) []FigmaNodePath {
	var targetNodes []FigmaNodePath

	visitNodes := make([]FigmaNodePath, 0)

	visitNodes = append(visitNodes, FigmaNodePath{
		Path:      rootNode.Name,
		FigmaNode: *rootNode,
	})

	for {
		if len(visitNodes) == 0 {
			break
		}

		visitNode := visitNodes[0]
		visitNodes = visitNodes[1:]

		if visitNode.Type == nodeType {
			targetNodes = append(targetNodes, visitNode)
		}

		for _, child := range visitNode.Children {
			visitNodes = append(visitNodes, FigmaNodePath{
				Path:      fmt.Sprintf("%v/%v", visitNode.Path, child.Name),
				FigmaNode: child,
			})
		}
	}

	return targetNodes
}

func Client(figmaPAT string) FigmaClient {
	return FigmaClient{
		token: figmaPAT,
	}
}

func (client *FigmaClient) request(path string) *http.Request {
	apiPath := fmt.Sprintf("%v%v", API_PATH, path)
	req, err := http.NewRequest(http.MethodGet, apiPath, http.NoBody)

	if err != nil {
		log.Fatalf("Failed to create request: '%v' %v", path, err)
	}

	req.Header.Add("X-Figma-Token", client.token)

	return req
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
		rawUrl:   *url,
	}
}

func (client *FigmaClient) GetFileNodes(figmaFileUrl string) FigmaNode {
	url := parseFigmaFileUrl(figmaFileUrl)

	nodeId := strings.ReplaceAll(url.Query.Get("node-id"), "-", ":")
	if nodeId == "" {
		log.Fatal("Missing node id")
	}

	path := fmt.Sprintf("/v1/files/%v/nodes?ids=%v&depth=5", url.FileKey, nodeId)

	request := client.request(path)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatalf("Failed to get file nodes: '%v' %v", path, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to read filed nodes response: '%v' %v", path, err)
	}

	fileResponse := FigmaNodeFile{}
	if err = json.Unmarshal(body, &fileResponse); err != nil {
		log.Fatalf("Failed to parse file nodes response: %v", err)
	}

	return fileResponse.Nodes[nodeId].Document
}

func (client *FigmaClient) GetImage(figmaFileUrl string) string {
	url := parseFigmaFileUrl(figmaFileUrl)
	nodeId := strings.ReplaceAll(url.Query.Get("node-id"), "-", ":")
	if nodeId == "" {
		log.Fatal("Missing node id")
	}

	path := fmt.Sprintf("/v1/images/%v?ids=%v", url.FileKey, nodeId)

	request := client.request(path)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatalf("Failed to get image: '%v' %v", path, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to parse image response: '%v' %v", path, err)
	}

	fileImageResponse := FigmaImageResponse{}
	if err = json.Unmarshal(body, &fileImageResponse); err != nil {
		log.Fatalf("Failed to parse image response: %v", err)
	}

	return fileImageResponse.Images[nodeId]
}
