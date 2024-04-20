package main

import (
	"log"
	"net/http"

	_ "net/http/pprof"
)

func main() {
	s := New()

	http.HandleFunc("GET /{$}", s.indexHandler)
	http.HandleFunc("GET /{username}", s.userHandler)
	http.HandleFunc("GET /static/{file}", s.staticHandler)
	http.HandleFunc("GET /discover", s.discoverHandler)
	http.HandleFunc("GET /settings", s.settingsHandler)
	http.HandleFunc("POST /settings/submit", s.settingsSubmitHandler)
	http.HandleFunc("GET /login", s.loginHandler)
	http.HandleFunc("POST /login", s.loginHandler)
	http.HandleFunc("GET /logout", s.logoutHandler)
	http.HandleFunc("POST /logout", s.logoutHandler)
	http.HandleFunc("POST /register", s.registerHandler)
	http.HandleFunc("GET /feeds/{url}", s.feedDetailsHandler)

	// left in-place for backwards compat
	http.HandleFunc("GET /feeds", s.settingsHandler)
	http.HandleFunc("POST /feeds/submit", s.settingsSubmitHandler)

	log.Println("main: listening on http://localhost:5544")
	log.Fatal(http.ListenAndServe(":5544", nil))
}
