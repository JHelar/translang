package db

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const DATABASE_NAME = "translang.db"

type DBClient struct {
	DB *sqlx.DB
}

func NewClient() DBClient {
	db, err := sqlx.Open("sqlite3", DATABASE_NAME)
	if err != nil {
		log.Fatal(err)
	}

	schema, err := os.ReadFile("./db/schema.sql")
	if err != nil {
		log.Fatal(err)
	}

	db.MustExec(string(schema))

	return DBClient{
		DB: db,
	}
}
