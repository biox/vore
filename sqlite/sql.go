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
	// user
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS user (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT UNIQUE NOT NULL,
                password TEXT NOT NULL,
                session_token TEXT UNIQUE,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`)
	if err != nil {
		panic(err)
	}
	// feed
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS feed (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                url TEXT NOT NULL,
                fetch_error TEXT,
		created_by TEXT NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`)
	if err != nil {
		panic(err)
	}
	// subscribe
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS subscribe (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id TEXT NOT NULL,
                feed_id TEXT NOT NULL,
		created_by TEXT NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`)
	if err != nil {
		panic(err)
	}

	wrapper := DB{
		sql: db,
	}

	return &wrapper
}

// TODO: think more about errors

func (s *DB) GetUsernameBySessionToken(token string) string {
	var username string
	err := s.sql.QueryRow("SELECT username FROM user WHERE session_token=?", token).Scan(&username)
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
	err := s.sql.QueryRow("SELECT password FROM user WHERE username=?", username).Scan(&password)
	if err == sql.ErrNoRows {
		return ""
	}
	if err != nil {
		panic(err)
	}
	return password
}

func (s *DB) SetSessionToken(username string, token string) {
	_, err := s.sql.Exec("UPDATE user SET session_token=? WHERE username=?", token, username)
	if err != nil {
		panic(err)
	}
}

func (s *DB) AddUser(username string, passwordHash string) {
	_, err := s.sql.Exec("INSERT INTO user (username, password) VALUES (?, ?)", username, passwordHash)
	if err != nil {
		panic(err)
	}
}

func (s *DB) UserExists(username string) bool {
	var result string
	err := s.sql.QueryRow("SELECT username FROM user WHERE username=?", username).Scan(&result)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		panic(err)
	}
	return true
}

func (s *DB) GetFeeds(username string) bool {
	var result string
	err := s.sql.QueryRow("SELECT username FROM user WHERE username=?", username).Scan(&result)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		panic(err)
	}
	return true
}
