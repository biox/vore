package main

import (
	"log"
	"net/http"
)

func main() {
	s := New()
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.indexHandler(w, r)
			return
		}
		if r.URL.Path == "/" {
			s.indexHandler(w, r)
			return
		}
	})
	mux.HandleFunc("/login", s.loginHandler)
	mux.HandleFunc("/logout", s.logoutHandler)
	mux.HandleFunc("/register", s.registerHandler)

	log.Println("listening on http://localhost:5544")
	log.Fatal(http.ListenAndServe(":5544", mux))
}
