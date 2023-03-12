package sqlite

import (
	"database/sql"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func SetupAndOpen(path string) *sql.DB {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
