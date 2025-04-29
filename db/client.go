package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const DATABASE_NAME = "translang.db"

type DBClient struct {
	db *sql.DB
}

func NewClient() DBClient {
	db, err := sql.Open("sqlite3", DATABASE_NAME)
	if err != nil {
		log.Fatal(err)
	}

	return DBClient{
		db: db,
	}
}
