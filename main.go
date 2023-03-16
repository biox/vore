package main

import (
	"log"
	"net/http"
)

func main() {
	s := New()
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.indexHandler)
	mux.HandleFunc("/login", s.loginHandler)
	mux.HandleFunc("/logout", s.logoutHandler)
	mux.HandleFunc("/register", s.registerHandler)

	log.Println("listening on http://localhost:5544")
	log.Fatal(http.ListenAndServe(":5544", mux))
}
