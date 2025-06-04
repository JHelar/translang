package figma

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

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

func (client *FigmaClient) GetFileNodes(figmaFileUrl string) (FigmaNode, error) {
	url := parseFigmaFileUrl(figmaFileUrl)

	nodeId := strings.ReplaceAll(url.Query.Get("node-id"), "-", ":")
	if nodeId == "" {
		return FigmaNode{}, fmt.Errorf("missing node id")
	}

	path := fmt.Sprintf("/v1/files/%v/nodes?ids=%v&depth=5", url.FileKey, nodeId)

	request := client.request(path, http.MethodGet, http.NoBody)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return FigmaNode{}, fmt.Errorf("failed to get file nodes: '%v' %v", path, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return FigmaNode{}, fmt.Errorf("failed to read filed nodes response: '%v' %v", path, err)
	}

	fileResponse := FigmaNodeFile{}
	if err = json.Unmarshal(body, &fileResponse); err != nil {
		return FigmaNode{}, fmt.Errorf("failed to parse file nodes response: %v", err)
	}

	return fileResponse.Nodes[nodeId].Document, nil
}

func (client *FigmaClient) GetImage(figmaFileUrl string) (string, error) {
	url := parseFigmaFileUrl(figmaFileUrl)
	nodeId := strings.ReplaceAll(url.Query.Get("node-id"), "-", ":")
	if nodeId == "" {
		return "", fmt.Errorf("missing node id")
	}

	path := fmt.Sprintf("/v1/images/%v?ids=%v", url.FileKey, nodeId)

	request := client.request(path, http.MethodGet, http.NoBody)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to get image: '%v' %v", path, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse image response: '%v' %v", path, err)
	}

	fileImageResponse := FigmaImageResponse{}
	if err = json.Unmarshal(body, &fileImageResponse); err != nil {
		return "", fmt.Errorf("failed to parse image response: %v", err)
	}

	return fileImageResponse.Images[nodeId], nil
}
