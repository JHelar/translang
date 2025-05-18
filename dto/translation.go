package dto

import (
	"database/sql"
	"time"
	"translang/db"
)

type Translation struct {
	ID              int64          `db:"id"`
	FigmaSourceUrl  string         `db:"figma_source_url"`
	ContextImageUrl sql.NullString `db:"context_image_url"`
	CreatedAt       time.Time      `db:"created_at"`
	SyncedAt        time.Time      `db:"synced_at"`
}

type TranslationNode struct {
	ID              int64     `db:"id"`
	FigmaTextNodeId string    `db:"figma_text_node_id"`
	SourceText      string    `db:"source_text"`
	CopyKey         string    `db:"copy_key"`
	CreatedAt       time.Time `db:"created_at"`
}

type TranslationNodeValue struct {
	TranslationNodeID int64     `db:"translation_node_id"`
	CopyText          string    `db:"copy_text"`
	CopyLanguage      string    `db:"copy_language"`
	CreatedAt         time.Time `db:"created_at"`
}

const SELECT_ALL_TRANSLATIONS = `
select id,figma_source_url,context_image_url,created_at,synced_at from translation
`

func GetAllTranslations(client db.DBClient) ([]Translation, error) {
	translations := []Translation{}
	if err := client.DB.Select(&translations, SELECT_ALL_TRANSLATIONS); err != nil {
		return nil, err
	}
	return translations, nil
}

const UPSERT_TRANSLATION_QUERY = `
insert into translation (figma_source_url) values (?) 
	on conflict (figma_source_url) do update set synced_at=current_timestamp
	returning id,figma_source_url,context_image_url,created_at,synced_at
`

func UpsertTranslation(figmaUrl string, client db.DBClient) (Translation, error) {
	translation := Translation{}

	tx := client.DB.MustBegin()
	row := tx.QueryRowx(UPSERT_TRANSLATION_QUERY, figmaUrl)
	if err := row.StructScan(&translation); err != nil {
		return Translation{}, err
	}
	if err := tx.Commit(); err != nil {
		return Translation{}, err
	}
	return translation, nil
}

const DELETE_TRANSLATION_QUERY = `
delete from translation where id=$1
`

const DELETE_TRANSLATION_NODE_CONNECTION = `
delete from translation_to_translation_node where translation_id=$1
`

func DeleteTranslation(translationID int64, client db.DBClient) error {
	tx := client.DB.MustBegin()

	tx.MustExec(DELETE_TRANSLATION_QUERY, translationID)
	tx.MustExec(DELETE_TRANSLATION_NODE_CONNECTION, translationID)

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

const GET_TRANSLATION_BY_ID_QUERY = `
select id,figma_source_url,context_image_url,created_at,synced_at from translation where id=$1
`

func GetTranslationByID(translationID int64, client db.DBClient) (Translation, error) {
	translation := Translation{}
	if err := client.DB.Get(&translation, GET_TRANSLATION_BY_ID_QUERY, translationID); err != nil {
		return Translation{}, err
	}
	return translation, nil
}

const GET_TRANSLATION_BY_FIGMA_SOURCE_URL = `
select id,figma_source_url,context_image_url,created_at,synced_at from translation where figma_source_url=$1
`

func GetTranslationByFigmaSourceUrl(figmaSourceUrl string, client db.DBClient) (Translation, error) {
	translation := Translation{}
	if err := client.DB.Get(&translation, GET_TRANSLATION_BY_ID_QUERY, figmaSourceUrl); err != nil {
		return Translation{}, err
	}
	return translation, nil
}

const GET_TRANSLATION_NODE_BY_SOURCE_TEXT = `
select id,figma_text_node_id,source_text,copy_key from translation_node where source_text=$1
`

func GetTranslationNodeBySourceText(sourceText string, client *db.DBClient) (TranslationNode, error) {
	node := TranslationNode{}
	if err := client.DB.Get(&node, GET_TRANSLATION_NODE_BY_SOURCE_TEXT, sourceText); err != nil {
		return TranslationNode{}, err
	}

	return node, nil
}

const GET_TRANSLATION_NODE_BY_ID = `
select id,figma_text_node_id,source_text,copy_key from translation_node where id=$1
`

func GetNodeByID(nodeID int64, client *db.DBClient) (TranslationNode, error) {
	node := TranslationNode{}
	if err := client.DB.Get(&node, GET_TRANSLATION_NODE_BY_ID, nodeID); err != nil {
		return TranslationNode{}, err
	}
	return node, nil
}

const UPDATE_CONTEXT_IMAGE = `
update translation
	set context_image_url=?
	where id=?
`

func (translation *Translation) UpdateContextImage(contextImageUrl string, client *db.DBClient) error {
	tx := client.DB.MustBegin()
	tx.MustExec(UPDATE_CONTEXT_IMAGE, contextImageUrl, translation.ID)
	if err := tx.Commit(); err != nil {
		return err
	}

	translation.ContextImageUrl = sql.NullString{
		String: contextImageUrl,
		Valid:  true,
	}

	return nil
}

const UPSERT_TRANSLATION_NODE = `
insert into translation_node (figma_text_node_id,source_text,copy_key) values (?,?,?)
	on conflict (figma_text_node_id) do update set source_text=excluded.source_text,copy_key=excluded.copy_key
	returning id,figma_text_node_id,source_text,copy_key,created_at
`

const UPSERT_TRANSLATION_NODE_CONNECTION = `
insert into translation_to_translation_node (translation_id,translation_node_id) values (?,?)
	on conflict (translation_id,translation_node_id) do nothing
`

func (translation *Translation) UpsertNode(figmaTextNodeId string, sourceText string, copyKey string, client *db.DBClient) (TranslationNode, error) {
	translationNode := TranslationNode{}

	tx := client.DB.MustBegin()
	if err := tx.Get(&translationNode, UPSERT_TRANSLATION_NODE, figmaTextNodeId, sourceText, copyKey); err != nil {
		tx.Rollback()
		return TranslationNode{}, err
	}
	if _, err := tx.Exec(UPSERT_TRANSLATION_NODE_CONNECTION, translation.ID, translationNode.ID); err != nil {
		tx.Rollback()
		return TranslationNode{}, err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return TranslationNode{}, err
	}

	return translationNode, nil
}

const SELECT_ALL_NODES_FOR_TRANSLATION = `
select id,figma_text_node_id,source_text,copy_key,created_at
	from translation_node
	left join translation_to_translation_node
		on translation_to_translation_node.translation_node_id=translation_node.id
	where translation_to_translation_node.translation_id=$1
`

func (translation *Translation) Nodes(client *db.DBClient) ([]TranslationNode, error) {
	nodes := []TranslationNode{}
	if err := client.DB.Select(&nodes, SELECT_ALL_NODES_FOR_TRANSLATION, translation.ID); err != nil {
		return nil, err
	}
	return nodes, nil
}

const UPSERT_VALUE = `
insert into translation_node_value (translation_node_id,copy_language,copy_text) values (?,?,?)
	on conflict (copy_language,translation_node_id) do update set copy_text=excluded.copy_text
	returning translation_node_id,copy_language,copy_text,created_at
`

func (node *TranslationNode) UpsertValue(copyLanguage string, copyText string, client *db.DBClient) (TranslationNodeValue, error) {
	value := TranslationNodeValue{}

	tx := client.DB.MustBegin()
	if err := tx.Get(&value, UPSERT_VALUE, node.ID, copyLanguage, copyText); err != nil {
		return TranslationNodeValue{}, err
	}
	if err := tx.Commit(); err != nil {
		return TranslationNodeValue{}, err
	}
	return value, nil
}

const SELECT_ALL_TRANSLATIONS_FOR_NODE = `
select id,figma_source_url,context_image_url,created_at,synced_at 
	from translation
	left join translation_to_translation_node
		on translation_to_translation_node.translation_id=translation.id
	where translation_to_translation_node.translation_node_id=$1
`

func (node *TranslationNode) GetTranslations(client *db.DBClient) ([]Translation, error) {
	translations := []Translation{}
	if err := client.DB.Select(&translations, SELECT_ALL_TRANSLATIONS_FOR_NODE, node.ID); err != nil {
		return nil, err
	}
	return translations, nil
}

const SELECT_ALL_VALUES = `
select translation_node_id,copy_language,copy_text,created_at from translation_node_value where translation_node_id=?
`

func (node *TranslationNode) Values(client *db.DBClient) ([]TranslationNodeValue, error) {
	values := []TranslationNodeValue{}
	if err := client.DB.Select(&values, SELECT_ALL_VALUES, node.ID); err != nil {
		return nil, err
	}
	return values, nil
}
