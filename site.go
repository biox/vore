package main

import (
	"fmt"
	"net/http"

	"git.j3s.sh/feeds.gay/auth"
	"git.j3s.sh/feeds.gay/sqlite"
	"golang.org/x/crypto/bcrypt"
)

type Site struct {
	// title of the website
	title string

	// site database handle
	db *sqlite.DB
}

// New returns a fully populated & ready for action Site
func New() *Site {
	title := "feeds.gay"
	s := Site{
		title: title,
		db:    sqlite.New(title + ".db"),
	}
	return &s
}

func (s *Site) indexHandler(w http.ResponseWriter, r *http.Request) {
	if !methodAllowed(w, r, "GET") {
		return
	}
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if s.loggedIn(r) {
		fmt.Fprintf(w, `<h1>index</h1>
			<small>logged in as %s
			(<a href="/logout">logout</a>)
			</small>`, s.username(r))
	} else {
		fmt.Fprintf(w, `<h1>index</h1>
			<a href="/login">login</a>`)
	}
}

func (s *Site) loginHandler(w http.ResponseWriter, r *http.Request) {
	if !methodAllowed(w, r, "GET", "POST") {
		return
	}
	if r.Method == "GET" {
		if s.loggedIn(r) {
			fmt.Fprintf(w, "you are already logged in :3\n")
		} else {
			fmt.Fprintf(w, `<h1>login</h1>
				<form method="POST" action="/login">
				<label for="username">username:</label>
				<input type="text" name="username" required><br>
				<label for="password">password:</label>
				<input type="password" name="password" required><br>
				<input type="submit" value="login">
				</form>
				<p>if you want to register a new account, click the tree:
				<a href="/register">🌳</a>`)
		}
	}
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := s.login(w, username, password)
		if err != nil {
			fmt.Fprintf(w, `<h1>incorrect username/password</h1>
				<p>if you want to register a new account, click the tree:
				<a href="/register">🌳</a>`)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// TODO: make this take a POST only in accordance w/ some spec
func (s *Site) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if !methodAllowed(w, r, "GET", "POST") {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: "",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Site) registerHandler(w http.ResponseWriter, r *http.Request) {
	if !methodAllowed(w, r, "GET", "POST") {
		return
	}

	if r.Method == "GET" {
		fmt.Fprintf(w, `<h1>register</h1>
			<form method="POST" action="/register">
			<label for="username">username:</label>
			<input type="text" name="username" required><br>
			<label for="password">password:</label>
			<input type="password" name="password" required><br>
			<input type="submit" value="login">
			</form>`)
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		err := s.register(username, password)
		if err != nil {
			internalServerError(w, "failed to register user")
			return
		}
		err = s.login(w, username, password)
		if err != nil {
			internalServerError(w, "extremely weird login error")
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// username fetches a client's username based
// on the sessionToken that user has set. username
// will return "" if there is no sessionToken.
func (s *Site) username(r *http.Request) string {
	sessionToken, err := r.Cookie("session_token")
	if err != nil {
		return ""
	}
	username := s.db.GetUsernameBySessionToken(sessionToken.Value)
	return username
}

func (s *Site) loggedIn(r *http.Request) bool {
	if s.username(r) == "" {
		return false
	}
	return true
}

// login compares the sqlite password field against the user supplied password and
// sets a session token against the supplied writer.
func (s *Site) login(w http.ResponseWriter, username string, password string) error {
	storedPassword := s.db.GetPassword(username)
	if storedPassword == "" {
		return fmt.Errorf("blank stored password")
	}
	if username == "" {
		return fmt.Errorf("username cannot be nil")
	}
	if password == "" {
		return fmt.Errorf("password cannot be nil")
	}
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password")
	}
	sessionToken := auth.GenerateSessionToken()
	s.db.SetSessionToken(username, sessionToken)
	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	})
	return nil
}

func (s *Site) register(username string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	s.db.AddUser(username, string(hashedPassword))
	return nil
}
