package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/server"
)

func gracefulShutdown(server *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("[INFO] Shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Server forced to shutdown with error: %v", err)
	}

	log.Println("[INFO] Server exiting")
	done <- true
}

func main() {
	server := server.NewServer()
	done := make(chan bool, 1)
	go gracefulShutdown(server, done)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("[ERROR] http server error: %v", err)
	}

	<-done
	log.Println("[INFO] Graceful shutdown complete")
}
