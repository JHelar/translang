package persistence

type PersistenceValue interface {
}

type ValuePayload struct {
	Language string
	Text     string
}

type PersistenceNode interface {
	GetID() string

	UpsertValue(payload ValuePayload) (PersistenceValue, error)

	ToPayload() (NodePayload, error)
}

type NodePayload struct {
	Source  string
	CopyKey string
	Values  []ValuePayload
}

type PersistenceTranslation interface {
	GetContextImageUrl() (string, error)
	GetFigmaSourceUrl() string
	GetID() string

	UpdateContextImage(contextImageUrl string) error
	UpsertNode(figmaTextNodeID string, payload NodePayload) (PersistenceNode, error)

	GetAllNodes() ([]PersistenceNode, error)
}

type PersistenceClient interface {
	UpsertTranslation(figmaUrl string) (PersistenceTranslation, error)

	GetTranslationByID(translationID string) (PersistenceTranslation, error)
	GetAllTranslations() ([]PersistenceTranslation, error)

	DeleteTranslationByID(translationID string) error

	GetNodeFromSourceText(sourceText string) (PersistenceNode, error)
	GetNodeByID(nodeID string) (PersistenceNode, error)

	GetAllNodes() ([]PersistenceNode, error)
}
