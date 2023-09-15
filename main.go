package main

import (
	"log"
	"net/http"

	"github.com/jba/muxpatterns"
)

func main() {
	s := New()
	mux := muxpatterns.NewServeMux()

	mux.HandleFunc("GET /{$}", s.indexHandler)
	mux.HandleFunc("GET /{username}", s.userHandler)
	mux.HandleFunc("GET /static/{file}", s.staticHandler)
	mux.HandleFunc("GET /discover", s.discoverHandler)
	mux.HandleFunc("GET /feeds", s.settingsHandler)
	mux.HandleFunc("POST /feeds/submit", s.feedsSubmitHandler)
	mux.HandleFunc("GET /login", s.loginHandler)
	mux.HandleFunc("POST /login", s.loginHandler)
	mux.HandleFunc("GET /logout", s.logoutHandler)
	mux.HandleFunc("POST /logout", s.logoutHandler)
	mux.HandleFunc("POST /register", s.registerHandler)

	log.Println("main: listening on http://localhost:5544")
	log.Fatal(http.ListenAndServe(":5544", mux))
}
