package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"git.j3s.sh/feeds.gay/sqlite"
)

type Site struct {
	db *sql.DB
}

// New returns a fully populated & ready for action Site
func New() *Site {
	s := Site{
		db: sqlite.SetupAndOpen("feeds.gay.db"),
	}
	return &s
}

func (s *Site) Start(addr string, mux *http.ServeMux) {
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *Site) rootHandler(w http.ResponseWriter, r *http.Request) {
	if !methodAllowed(w, r, "GET") {
		return
	}
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "feeds.gay is dope & you should like it\n")
}

func (s *Site) loginHandler(w http.ResponseWriter, r *http.Request) {
	if !methodAllowed(w, r, "GET", "POST") {
		return
	}
	if r.Method == "GET" {
		// if logged out:
		fmt.Fprintf(w, "display login forms\n")
		// if logged in:
		fmt.Fprintf(w, "you are already logged in :D\n")
	}
	if r.Method == "POST" {
		fmt.Fprintf(w, "cmon POST\n")
	}
}

func (s *Site) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if !methodAllowed(w, r, "POST") {
		return
	}
	// TODO: delete session cookie
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Site) registerHandler(w http.ResponseWriter, r *http.Request) {
	if !methodAllowed(w, r, "POST") {
		return
	}
	// TODO: create user in database
	// TODO: add session cookie
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
