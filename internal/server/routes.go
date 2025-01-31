package server

import (
	"net/http"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	commonMiddleware := middleware.ChainMiddleware(
		s.rateLimiter.Middleware,
		middleware.LogMiddleware,
	)

	registerRoute := func(path string, handler http.HandlerFunc) {
		mux.HandleFunc(path, commonMiddleware(handler))
	}

	registerRoute("/", s.indexHandler)
	registerRoute("/randomize", s.randomizeHandler)
	registerRoute("/seed/{id}", s.seedHandler)
	registerRoute("/download/{id}", s.seedTrackerHandler)

	return mux
}
