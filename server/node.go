package server

import (
	"fmt"
	"net/http"
	"translang/persistence"
	"translang/template"
)

func (client ServerClient) NodesRoute(w http.ResponseWriter, r *http.Request) {
	nodes, err := client.persistence.GetAllNodes()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	props := template.NodesProp{}
	for _, node := range nodes {
		payload, err := node.ToPayload()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		detailsUrl, err := client.router.Get("getNode").URL("id", node.GetID())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		props.Nodes = append(props.Nodes, template.NodeRowProp{
			SourceText: payload.Source,
			CopyKey:    payload.CopyKey,
			DetailsUrl: detailsUrl.String(),
		})
	}
	template.Nodes(props).Render(r.Context(), w)
}

func (client ServerClient) NodeDetailsRoute(w http.ResponseWriter, r *http.Request) {
	node, err := client.persistence.GetNodeByID(r.Form.Get("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload, err := node.ToPayload()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	props := template.NodeModalProps{
		NodePayload: payload,
	}

	for _, value := range payload.Values {
		updateValueURL, err := client.router.Get("updateNodeValue").URL("id", node.GetID(), "language", value.Language)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props.Values = append(props.Values, struct {
			persistence.ValuePayload
			UpdateValueURL string
		}{
			ValuePayload:   value,
			UpdateValueURL: updateValueURL.String(),
		})
	}

	template.NodeModal(props).Render(r.Context(), w)
}

func (client ServerClient) UpdateTranslationValue(w http.ResponseWriter, r *http.Request) {
	node, err := client.persistence.GetNodeByID(r.Form.Get("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting translation node: %v\n", err), 404)
		return
	}

	if _, err := node.UpsertValue(persistence.ValuePayload{
		Language: r.Form.Get("language"),
		Text:     r.Form.Get("text"),
	}); err != nil {
		http.Error(w, fmt.Sprintf("Error updating node value: %v\n", err), 500)
		return
	}

	template.ToastSuccess(template.ToastProps{
		Message: "Updated translation",
	}).Render(r.Context(), w)
}
