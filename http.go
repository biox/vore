package main

import (
	"net/http"
	"strings"
)

func internalServerError(w http.ResponseWriter, details string) {
	status := "oopsie woopsie, uwu\n"
	status += "we made a fucky wucky!!\n\n"
	status += "500 internal server error: " + details
	http.Error(w, status, http.StatusInternalServerError)
}

// methodAllowed takes an http w/r, and returns true if the
// http requests method is in teh allowedMethods list.
// if methodNotAllowed returns false, it has already
// written a request & it's on the caller to close it.
func methodAllowed(w http.ResponseWriter, r *http.Request, allowedMethods ...string) bool {
	allowed := false
	for _, m := range allowedMethods {
		if m == r.Method {
			allowed = true
		}
	}
	if allowed == false {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
	}
	return allowed
}
