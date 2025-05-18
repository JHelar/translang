package db

import (
	"strconv"
	"translang/db"
	"translang/dto"
	"translang/persistence"
)

type DBPersistenceNode struct {
	*db.DBClient
	node dto.TranslationNode
}

func (client DBPersistenceNode) UpsertValue(payload persistence.ValuePayload) (persistence.PersistenceValue, error) {
	value, err := client.node.UpsertValue(payload.Language, payload.Text, client.DBClient)
	if err != nil {
		return DBPersistenceValue{}, err
	}

	return DBPersistenceValue{
		DBClient: client.DBClient,
		value:    value,
	}, nil
}

func (client DBPersistenceNode) ToPayload() (persistence.NodePayload, error) {
	values, err := client.node.Values(client.DBClient)
	if err != nil {
		return persistence.NodePayload{}, err
	}

	var valuesPayload []persistence.ValuePayload
	for _, value := range values {
		valuesPayload = append(valuesPayload, persistence.ValuePayload{
			Language: value.CopyLanguage,
			Text:     value.CopyText,
		})
	}

	return persistence.NodePayload{
		NodeId:  client.node.FigmaTextNodeId,
		Source:  client.node.SourceText,
		CopyKey: client.node.CopyKey,
		Values:  valuesPayload,
	}, nil
}

func (client DBPersistenceNode) GetID() string {
	return strconv.FormatInt(client.node.ID, 10)
}
