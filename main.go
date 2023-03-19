package main

import (
	"fmt"
	"net/http"
	"strings"
)

func main() {
	s := New()
	mux := http.NewServeMux()

	// since "/" is a wildcard, this anonymous function acts
	// as a router for patterns that can't be registered at
	// start time. e.g. /j3s or /j3s/feeds
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.indexHandler(w, r)
			return
		}
		// handles /<username>
		if s.db.UserExists(strings.TrimPrefix(r.URL.Path, "/")) {
			s.userHandler(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/feeds", s.feedsHandler)
	mux.HandleFunc("/login", s.loginHandler)
	mux.HandleFunc("/logout", s.logoutHandler)
	mux.HandleFunc("/register", s.registerHandler)

	fmt.Printf("vore: listening on http://localhost:5544")
	panic(http.ListenAndServe(":5544", mux))
}
