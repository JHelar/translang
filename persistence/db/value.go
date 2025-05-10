package db

import (
	"translang/db"
	"translang/dto"
)

type DBPersistenceValue struct {
	*db.DBClient
	value dto.TranslationNodeValue
}
