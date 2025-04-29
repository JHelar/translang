package dto

import (
	"time"
	"translang/db"
)

type Translation struct {
	ID              string    `db:"id"`
	FigmaSourceUrl  string    `db:"figma_source_url"`
	ContextImageUrl string    `db:"context_image_url"`
	CreatedAt       time.Time `db:"created_at"`
	SyncedAt        time.Time `db:"synced_at"`
}

func UpsertTranslation(figmaUrl string, client db.DBClient) {

}
