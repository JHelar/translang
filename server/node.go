package server

import (
	"fmt"
	"net/http"
	"strconv"
	"translang/dto"
	"translang/template"
)

func (client ServerClient) NodesRoute(w http.ResponseWriter, r *http.Request) {
	nodes, err := dto.GetAllNodes(&client.db)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	props := template.NodesProp{}
	for _, node := range nodes {
		detailsUrl, err := client.router.Get("getNode").URL("id", strconv.FormatInt(node.ID, 10))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		props.Nodes = append(props.Nodes, template.NodeRowProp{
			SourceText: node.SourceText,
			CopyKey:    node.CopyKey,
			DetailsUrl: detailsUrl.String(),
		})
	}
	template.Nodes(props).Render(r.Context(), w)
}

func (client ServerClient) NodeDetailsRoute(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseInt(r.Form.Get("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	node, err := dto.GetNodeByID(nodeID, &client.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	values, err := node.Values(&client.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	props := template.NodeModalProps{
		TranslationNode: node,
		Values: []struct {
			dto.TranslationNodeValue
			UpdateValueURL string
		}{},
	}

	for _, value := range values {
		updateValueURL, err := client.router.Get("updateNodeValue").URL("id", strconv.FormatInt(node.ID, 10), "language", value.CopyLanguage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props.Values = append(props.Values, struct {
			dto.TranslationNodeValue
			UpdateValueURL string
		}{
			TranslationNodeValue: value,
			UpdateValueURL:       updateValueURL.String(),
		})
	}

	template.NodeModal(props).Render(r.Context(), w)
}

func (client ServerClient) UpdateTranslationValue(w http.ResponseWriter, r *http.Request) {
	nodeID, err := strconv.ParseInt(r.Form.Get("id"), 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting translation node: %v\n", err), 404)
		return
	}

	node, err := dto.GetNodeByID(nodeID, &client.db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting translation node: %v\n", err), 404)
		return
	}

	if _, err := node.UpsertValue(r.Form.Get("language"), r.Form.Get("text"), &client.db); err != nil {
		http.Error(w, fmt.Sprintf("Error updating node value: %v\n", err), 500)
		return
	}

	template.ToastSuccess(template.ToastProps{
		Message: "Updated translation",
	}).Render(r.Context(), w)
}
