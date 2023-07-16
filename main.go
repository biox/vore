package main

import (
	"log"
	"net/http"
)

func main() {
	s := New()
	mux := http.NewServeMux()

	// rootHandler handles /, /<username>, and 404
	mux.HandleFunc("/", s.rootHandler)
	mux.HandleFunc("/discover", s.discoverHandler)
	mux.HandleFunc("/feeds", s.settingsHandler)
	mux.HandleFunc("/feeds/submit", s.feedsSubmitHandler)
	mux.HandleFunc("/login", s.loginHandler)
	mux.HandleFunc("/logout", s.logoutHandler)
	mux.HandleFunc("/register", s.registerHandler)

	log.Println("main: listening on http://localhost:5544")
	log.Fatal(http.ListenAndServe(":5544", mux))
}
