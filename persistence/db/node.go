package db

import (
	"translang/db"
	"translang/dto"
	"translang/persistence"
	"translang/translator"
)

type DBPersistenceNode struct {
	*db.DBClient
	node dto.TranslationNode
}

func (client DBPersistenceNode) UpsertValue(language string, text string) (persistence.PersistenceValue, error) {
	value, err := client.node.UpsertValue(language, text, client.DBClient)
	if err != nil {
		return DBPersistenceValue{}, err
	}

	return DBPersistenceValue{
		DBClient: client.DBClient,
		value:    value,
	}, nil
}

func (client DBPersistenceNode) ToResult() (translator.TranslationResult, error) {
	return client.node.ToResult(client.DBClient)
}
