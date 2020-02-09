package proxy

import (
	"log"
	"net/http"
)

// LogRequest can be used as a middleware chain to log every request before proxying the request
func LogRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[FROM]\t%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler(w, r)
	}
}
