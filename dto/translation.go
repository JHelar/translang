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
	ID              int64  `db:"id"`
	FigmaTextNodeId string `db:"figma_text_node_id"`
	TranslationID   int64  `db:"translation_id"`
	SourceText      string `db:"source_text"`
	CopyKey         string `db:"copy_key"`
}

type TranslationNodeValue struct {
	TranslationNodeID int64  `db:"translation_node_id"`
	CopyText          string `db:"copy_text"`
	CopyLanguage      string `db:"copy_language"`
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
delete from translation where figma_source_url=$1
`

func DeleteTranslation(figmaUrl string, client db.DBClient) error {
	tx := client.DB.MustBegin()
	tx.MustExec(DELETE_TRANSLATION_QUERY, figmaUrl)
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
insert into translation_node (figma_text_node_id,translation_id,source_text,copy_key) values (?,?,?,?)
	on conflict (figma_text_node_id) do update set source_text=excluded.source_text,copy_key=excluded.copy_key
	returning id,figma_text_node_id,translation_id,source_text,copy_key
`

func (translation *Translation) UpsertNode(figmaTextNodeId string, sourceText string, copyKey string, client *db.DBClient) (TranslationNode, error) {
	translationNode := TranslationNode{}

	tx := client.DB.MustBegin()
	if err := tx.Get(&translationNode, UPSERT_TRANSLATION_NODE, figmaTextNodeId, translation.ID, sourceText, copyKey); err != nil {
		return TranslationNode{}, err
	}
	if err := tx.Commit(); err != nil {
		return TranslationNode{}, err
	}

	return translationNode, nil
}

const SELECT_ALL_NODES = `
select id,figma_text_node_id,translation_id,source_text,copy_key from translation_node where translation_id=?
`

func (translation *Translation) Nodes(client *db.DBClient) ([]TranslationNode, error) {
	nodes := []TranslationNode{}
	if err := client.DB.Select(&nodes, SELECT_ALL_NODES, translation.ID); err != nil {
		return nil, err
	}
	return nodes, nil
}

const UPSERT_VALUE = `
insert into translation_node_value (translation_node_id,copy_language,copy_text) values (?,?,?)
	on conflict (copy_language,translation_node_id) do update set copy_text=excluded.copy_text
	returning translation_node_id,copy_language,copy_text
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

const SELECT_ALL_VALUES = `
select translation_node_id,copy_language,copy_text from translation_node_value where translation_node_id=?
`

func (node *TranslationNode) Values(client *db.DBClient) ([]TranslationNodeValue, error) {
	values := []TranslationNodeValue{}
	if err := client.DB.Select(&values, SELECT_ALL_VALUES, node.ID); err != nil {
		return nil, err
	}
	return values, nil
}
