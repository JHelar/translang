package persistence

import "translang/translator"

type PersistenceValue interface {
}

type PersistenceNode interface {
	UpsertValue(language string, text string) (PersistenceValue, error)

	ToResult() (translator.TranslationResult, error)
}

type PersistenceTranslation interface {
	GetContextImageUrl() (string, error)
	GetFigmaSourceUrl() string
	GetID() string

	UpdateContextImage(contextImageUrl string) error
	UpsertNode(result translator.TranslationResult) (PersistenceNode, error)

	GetAllNodes() ([]PersistenceNode, error)

	ToResult() (translator.ProcessResult, error)
}

type Persistence interface {
	UpsertTranslation(figmaUrl string) (PersistenceTranslation, error)
	GetTranslationByID(translationID string) (PersistenceTranslation, error)
	GetAllTranslations() ([]PersistenceTranslation, error)
}
