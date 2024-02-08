package main

import (
	"log"
	"net/http"
)

func main() {
	s := New()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", s.indexHandler)
	mux.HandleFunc("GET /{username}", s.userHandler)
	mux.HandleFunc("GET /static/{file}", s.staticHandler)
	mux.HandleFunc("GET /discover", s.discoverHandler)
	mux.HandleFunc("GET /settings", s.settingsHandler)
	mux.HandleFunc("POST /settings/submit", s.settingsSubmitHandler)
	mux.HandleFunc("GET /login", s.loginHandler)
	mux.HandleFunc("POST /login", s.loginHandler)
	mux.HandleFunc("GET /logout", s.logoutHandler)
	mux.HandleFunc("POST /logout", s.logoutHandler)
	mux.HandleFunc("POST /register", s.registerHandler)
	mux.HandleFunc("GET /feeds/{url}", s.feedDetailsHandler)

	// left in-place for backwards compat
	mux.HandleFunc("GET /feeds", s.settingsHandler)
	mux.HandleFunc("POST /feeds/submit", s.settingsSubmitHandler)

	log.Println("main: listening on http://localhost:5544")
	log.Fatal(http.ListenAndServe(":5544", mux))
}
