package database

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func NewDB(driver string, dsn string) *sqlx.DB {

	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
