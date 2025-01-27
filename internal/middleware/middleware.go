package middleware

import "net/http"

type Middleware func(http.HandlerFunc) http.HandlerFunc

func ChainMiddleware(m ...Middleware) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		chained := h
		for _, x := range m {
			chained = x(chained)
		}
		return chained
	}
}
