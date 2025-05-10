package db

import (
	"fmt"
	"translang/db"
	"translang/dto"
	"translang/persistence"
	"translang/translator"
)

type DBPersistenceTranslation struct {
	*db.DBClient
	translation dto.Translation
}

func (client DBPersistenceTranslation) UpdateContextImage(contextImageUrl string) error {
	return client.translation.UpdateContextImage(contextImageUrl, client.DBClient)
}

func (client DBPersistenceTranslation) UpsertNode(result translator.TranslationResult) (persistence.PersistenceNode, error) {
	node, err := client.translation.UpsertNode(result.NodeId, result.Source, result.CopyKey, client.DBClient)
	if err != nil {
		return DBPersistenceNode{}, err
	}

	for _, value := range result.Values {
		if _, err = node.UpsertValue(value.Language, value.Text, client.DBClient); err != nil {
			return DBPersistenceNode{}, err
		}
	}

	return DBPersistenceNode{
		DBClient: client.DBClient,
		node:     node,
	}, nil
}

func (client DBPersistenceTranslation) GetAllNodes() ([]persistence.PersistenceNode, error) {
	nodes, err := client.translation.Nodes(client.DBClient)
	if err != nil {
		return nil, err
	}

	var persistenceNodes []persistence.PersistenceNode
	for _, node := range nodes {
		persistenceNodes = append(persistenceNodes, DBPersistenceNode{
			DBClient: client.DBClient,
			node:     node,
		})
	}

	return persistenceNodes, nil
}

func (client DBPersistenceTranslation) ToResult() (translator.ProcessResult, error) {
	return client.translation.ToResult(client.DBClient)
}

func (client DBPersistenceTranslation) GetContextImageUrl() (string, error) {
	if client.translation.ContextImageUrl.Valid && client.translation.ContextImageUrl.String != "" {
		return client.translation.ContextImageUrl.String, nil
	}

	return "", fmt.Errorf("missing ContextImageUrl")
}

func (client DBPersistenceTranslation) GetFigmaSourceUrl() string {
	return client.translation.FigmaSourceUrl
}

func (client DBPersistenceTranslation) GetID() string {
	return fmt.Sprint(client.translation.ID)
}
