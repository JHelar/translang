package server

import (
	"net/http"
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

}
