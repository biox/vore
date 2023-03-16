package sqlite

import (
	"database/sql"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

type DB struct {
	sql *sql.DB
}

// New opens a sqlite database, populates it with tables, and
// returns a ready-to-use *sqlite.DB object which is used for
// abstracting database queries.
func New(path string) *DB {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (     
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT UNIQUE NOT NULL,
                password TEXT NOT NULL,
                session_token TEXT UNIQUE
        )`)
	if err != nil {
		panic(err)
	}

	wrapper := DB{
		sql: db,
	}

	return &wrapper
}

// TODO: evaluate error cases

func (s *DB) GetUsernameBySessionToken(token string) string {
	var username string
	err := s.sql.QueryRow("SELECT username FROM users WHERE session_token=?", token).Scan(&username)
	if err == sql.ErrNoRows {
		return ""
	}
	if err != nil {
		panic(err)
	}
	return username
}

func (s *DB) GetPassword(username string) string {
	var password string
	err := s.sql.QueryRow("SELECT password FROM users WHERE username=?", username).Scan(&password)
	if err == sql.ErrNoRows {
		return ""
	}
	if err != nil {
		panic(err)
	}
	return password
}
func (s *DB) SetSessionToken(username string, token string) {
	_, err := s.sql.Exec("UPDATE users SET session_token=? WHERE username=?", token, username)
	if err != nil {
		panic(err)
	}
}

func (s *DB) AddUser(username string, passwordHash string) {
	_, err := s.sql.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, passwordHash)
	if err != nil {
		panic(err)
	}
}
