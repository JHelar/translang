package persistence

type PersistenceValue interface {
}

type ValuePayload struct {
	Language string
	Text     string
}

type PersistenceNode interface {
	UpsertValue(payload ValuePayload) (PersistenceValue, error)

	ToPayload() (NodePayload, error)
}

type NodePayload struct {
	NodeId  string
	Source  string
	CopyKey string
	Values  []ValuePayload
}

type PersistenceTranslation interface {
	GetContextImageUrl() (string, error)
	GetFigmaSourceUrl() string
	GetID() string

	UpdateContextImage(contextImageUrl string) error
	UpsertNode(payload NodePayload) (PersistenceNode, error)

	GetAllNodes() ([]PersistenceNode, error)
}

type PersistenceClient interface {
	UpsertTranslation(figmaUrl string) (PersistenceTranslation, error)

	GetTranslationByID(translationID string) (PersistenceTranslation, error)
	GetAllTranslations() ([]PersistenceTranslation, error)

	DeleteTranslationByID(translationID string) error

	GetNodeFromSourceText(sourceText string) (PersistenceNode, error)
}
