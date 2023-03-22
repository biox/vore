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
                url TEXT UNIQUE NOT NULL,
                fetch_error TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`)
	if err != nil {
		panic(err)
	}
	// subscribe
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS subscribe (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id INTEGER NOT NULL,
                feed_id INTEGER NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`)
	if err != nil {
		panic(err)
	}

	return &DB{sql: db}
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

func (s *DB) AddUser(username string, passwordHash string) error {
	_, err := s.sql.Exec("INSERT INTO user (username, password) VALUES (?, ?)", username, passwordHash)
	return err
}

func (s *DB) Subscribe(username string, feedURL string) {
	uid := s.GetUserID(username)
	fid := s.GetFeedID(feedURL)
	var id int
	err := s.sql.QueryRow("SELECT id FROM subscribe WHERE user_id=? AND feed_id=?", uid, fid).Scan(&id)
	if err == sql.ErrNoRows {
		_, err := s.sql.Exec("INSERT INTO subscribe (user_id, feed_id) VALUES (?, ?)", uid, fid)
		if err != nil {
			panic(err)
		}
		return
	}
	if err != nil {
		panic(err)
	}
}

func (s *DB) UnsubscribeAll(username string) {
	_, err := s.sql.Exec("DELETE FROM subscribe WHERE user_id=?", s.GetUserID(username))
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

func (s *DB) GetAllFeedURLs() []string {
	// TODO: BAD SELECT STATEMENT!! SORRY :( --wesley
	rows, err := s.sql.Query("SELECT url FROM feed")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		err = rows.Scan(&url)
		if err != nil {
			panic(err)
		}
		urls = append(urls, url)
	}
	return urls
}

func (s *DB) GetUserFeedURLs(username string) []string {
	uid := s.GetUserID(username)

	// this query returns sql rows representing the list of
	// rss feed urls the user is subscribed to
	rows, err := s.sql.Query(`
		SELECT f.url
		FROM feed f
		JOIN subscribe s ON f.id = s.feed_id
		JOIN user u ON s.user_id = u.id
		WHERE u.id = ?`, uid)
	if err == sql.ErrNoRows {
		return []string{}
	}
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		err = rows.Scan(&url)
		if err != nil {
			panic(err)
		}
		urls = append(urls, url)
	}
	return urls
}

func (s *DB) GetUserID(username string) int {
	var uid int
	err := s.sql.QueryRow("SELECT id FROM user WHERE username=?", username).Scan(&uid)
	if err != nil {
		panic(err)
	}
	return uid
}

func (s *DB) GetFeedID(feedURL string) int {
	var fid int
	err := s.sql.QueryRow("SELECT id FROM feed WHERE url=?", feedURL).Scan(&fid)
	if err != nil {
		panic(err)
	}
	return fid
}

// WriteFeed writes an rss feed to the database for permanent storage
// if the given feed already exists, WriteFeed does nothing.
func (s *DB) WriteFeed(url string) {
	_, err := s.sql.Exec(`INSERT INTO feed(url) VALUES(?)
				ON CONFLICT(url) DO NOTHING`, url)
	if err != nil {
		panic(err)
	}
}

// WriteFeed writes an rss feed to the database for permanent storage
// if the given feed already exists, WriteFeed does nothing.
func (s *DB) SetFeedFetchError(url string, fetchErr string) error {
	_, err := s.sql.Exec("UPDATE feed SET fetch_error=? WHERE url=?", fetchErr, url)
	return err
}

// WriteFeed writes an rss feed to the database for permanent storage
// if the given feed already exists, WriteFeed does nothing.
func (s *DB) GetFeedFetchError(url string) (string, error) {
	var result sql.NullString
	err := s.sql.QueryRow("SELECT fetch_error FROM feed WHERE url=?", url).Scan(&result)
	if err != nil {
		return "", err
	}
	if result.Valid {
		return result.String, nil
	}
	return "", nil
}
