package main

import (
	"log"
	"net/http"
)

func main() {
	s := New()
	mux := http.NewServeMux()
	// handles /, /<username>, and 404
	mux.HandleFunc("/", s.rootHandler)
	mux.HandleFunc("/feeds", s.feedsHandler)
	mux.HandleFunc("/feeds/submit", s.feedsSubmitHandler)
	mux.HandleFunc("/login", s.loginHandler)
	mux.HandleFunc("/logout", s.logoutHandler)
	mux.HandleFunc("/register", s.registerHandler)

	log.Println("main: listening on http://localhost:5544")
	log.Fatal(http.ListenAndServe(":5544", mux))
}
