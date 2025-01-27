package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/database"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/middleware"
)

type Server struct {
	port        int
	db          database.Service
	rateLimiter *middleware.RateLimiterMiddleware
}

func NewServer() *http.Server {
	port := 8080
	db := database.New()
	rateLimiter := middleware.NewRateLimiterMiddleware(1, 5)

	NewServer := &Server{
		port:        port,
		db:          db,
		rateLimiter: rateLimiter,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
