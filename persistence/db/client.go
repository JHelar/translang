package db

import (
	"strconv"
	"translang/db"
	"translang/dto"
	"translang/persistence"
)

type DBPersistenceClient struct {
	db.DBClient
}

func NewClient() DBPersistenceClient {
	return DBPersistenceClient{
		db.NewClient(),
	}
}

func (client DBPersistenceClient) UpsertTranslation(figmaUrl string) (persistence.PersistenceTranslation, error) {
	translation, err := dto.UpsertTranslation(figmaUrl, client.DBClient)
	if err != nil {
		return DBPersistenceTranslation{}, err
	}

	return DBPersistenceTranslation{
		DBClient:    &client.DBClient,
		translation: translation,
	}, nil
}

func (client DBPersistenceClient) GetTranslationByID(translationIDString string) (persistence.PersistenceTranslation, error) {
	translationID, err := strconv.ParseInt(translationIDString, 10, 64)
	if err != nil {
		return DBPersistenceTranslation{}, err
	}

	translation, err := dto.GetTranslationByID(translationID, client.DBClient)
	if err != nil {
		return DBPersistenceTranslation{}, err
	}

	return DBPersistenceTranslation{
		DBClient:    &client.DBClient,
		translation: translation,
	}, nil
}

func (client DBPersistenceClient) GetAllTranslations() ([]persistence.PersistenceTranslation, error) {
	translations, err := dto.GetAllTranslations(client.DBClient)
	if err != nil {
		return nil, err
	}

	var persistenceTranslations []persistence.PersistenceTranslation
	for _, translation := range translations {
		persistenceTranslations = append(persistenceTranslations, DBPersistenceTranslation{
			DBClient:    &client.DBClient,
			translation: translation,
		})
	}

	return persistenceTranslations, nil
}

func (client DBPersistenceClient) DeleteTranslationByID(translationIDString string) error {
	translationID, err := strconv.ParseInt(translationIDString, 10, 64)
	if err != nil {
		return err
	}

	return dto.DeleteTranslation(translationID, client.DBClient)
}

func (client DBPersistenceClient) GetNodeFromSourceText(sourceText string) (persistence.PersistenceNode, error) {
	node, err := dto.GetTranslationNodeBySourceText(sourceText, &client.DBClient)
	if err != nil {
		return DBPersistenceNode{}, err
	}

	return DBPersistenceNode{
		DBClient: &client.DBClient,
		node:     node,
	}, nil
}
