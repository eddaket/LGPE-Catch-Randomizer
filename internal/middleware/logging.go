package middleware

import (
	"log"
	"net/http"
)

func LogMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] Request received: %s %s", r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}
