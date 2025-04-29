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
}

type TranslationNodeValue struct {
	TranslationNodeID int64  `db:"translation_node_id"`
	CopyKey           string `db:"copy_key"`
	CopyLanguage      string `db:"copy_language"`
	CopyText          string `db:"copy_text"`
}

const NEW_TRANSLATION_QUERY = `
insert into translation (figma_source_url) values (?) returning id,figma_source_url,context_image_url,created_at,synced_at
`

const GET_TRANSLATION_BY_ID_QUERY = `
select id,figma_source_url,context_image_url,created_at,synced_at from translation where id=$1
`

func NewTranslation(figmaUrl string, client db.DBClient) (Translation, error) {
	translation := Translation{}

	row := client.DB.QueryRowx(NEW_TRANSLATION_QUERY, figmaUrl)
	if err := row.StructScan(&translation); err != nil {
		return Translation{}, err
	}
	return translation, nil
}

func GetTranslationByID(translationID int64, client db.DBClient) (Translation, error) {
	translation := Translation{}
	if err := client.DB.Get(&translation, GET_TRANSLATION_BY_ID_QUERY, translationID); err != nil {
		return Translation{}, err
	}
	return translation, nil
}
