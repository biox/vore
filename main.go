package main

import (
	"log"
	"net/http"
)

func main() {
	s := New()
	log.Println("listening on http://localhost:5544")
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.indexHandler)
	mux.HandleFunc("/login", s.loginHandler)
	mux.HandleFunc("/logout", s.logoutHandler)
	mux.HandleFunc("/register", s.registerHandler)
	s.Start(":5544", mux)
}
