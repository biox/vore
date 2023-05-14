package main

import (
	"fmt"
	"net/http"
)

func main() {
	s := New()
	mux := http.NewServeMux()
	// since "/" is a wildcard, this anonymous function acts
	// as a router for patterns that can't be registered at
	// start time. e.g. /j3s or /j3s/feeds

	// handles /, /<username>, and 404
	mux.HandleFunc("/", s.rootHandler)
	mux.HandleFunc("/feeds", s.feedsHandler)
	mux.HandleFunc("/feeds/submit", s.feedsSubmitHandler)
	mux.HandleFunc("/login", s.loginHandler)
	mux.HandleFunc("/logout", s.logoutHandler)
	mux.HandleFunc("/register", s.registerHandler)

	fmt.Println("main: listening on http://localhost:5544")
	panic(http.ListenAndServe(":5544", mux))
}
