package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/logic"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/middleware"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/output"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/pokemon"
)

func main() {
	rateLimiter := middleware.NewRateLimiterMiddleware(1, 5)
	mux := registerRoutes(rateLimiter)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: mux,
	}

	log.Printf("[INFO] Starting server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("[ERROR] Server failed: %v", err)
	}
}

func registerRoutes(rateLimiter *middleware.RateLimiterMiddleware) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", rateLimiter.Middleware(http.HandlerFunc(indexHandler)))
	mux.Handle("/randomize", rateLimiter.Middleware(http.HandlerFunc(randomizeHandler)))
	return mux
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] indexHandler: Serving index.html")

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("[ERROR] indexHandler: Error parsing template %v", err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

func randomizeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] randomizeHandler: Randomization initiated")

	seed, version, allowedOnePct, err := validateRandomizeParams(r)
	if err != nil {
		log.Printf("[WARN] randomizeHandler: Invalid parameters %v", err)
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	dataChan := make(chan []byte)
	errorChan := make(chan error)
	go func() {
		data, err := performRandomization(seed, version, allowedOnePct)
		if err != nil {
			errorChan <- err
			return
		}

		dataChan <- data
	}()

	select {
	case <-ctx.Done():
		log.Printf("[ERROR] randomizeHandler: Timeout")
		http.Error(w, "Randomization timed out", http.StatusGatewayTimeout)
		return
	case err := <-errorChan:
		log.Printf("[WARN] randomizeHandler: Error encountered %v", err)
		http.Error(w, "Error during randomization", http.StatusInternalServerError)
		return
	case data := <-dataChan:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s_%d.json", "Catch_Rando", version, seed))
		w.Write(data)

		log.Printf("[INFO] randomizeHandler: Randomization success")
	}
}

func validateRandomizeParams(r *http.Request) (int64, string, int, error) {
	query := r.URL.Query()
	seedStr := query.Get("seed")
	version := query.Get("version")
	allowedOnePctStr := query.Get("allowedOnePct")

	var err error

	var seed int64
	if seedStr == "" {
		seed = logic.GetComputedSeed()
	} else {
		seed, err = strconv.ParseInt(seedStr, 10, 64)
		if err != nil || seed < 0 {
			return 0, "", 0, fmt.Errorf("invalid seed: %s", seedStr)
		}
	}

	if version != "Pikachu" && version != "Eevee" {
		return 0, "", 0, fmt.Errorf("invalid version: %s", version)
	}

	var allowedOnePct int
	if allowedOnePctStr == "unlimited" {
		allowedOnePct = 50
	} else {
		allowedOnePct, err = strconv.Atoi(allowedOnePctStr)
		if err != nil || allowedOnePct < 0 {
			return 0, "", 0, fmt.Errorf("invalid allowedOnePct: %s", allowedOnePctStr)
		}
	}

	return seed, version, allowedOnePct, nil
}

func performRandomization(seed int64, version string, allowedOnePct int) ([]byte, error) {
	pokemonMap, err := pokemon.LoadPokemonData("data/pokemon_pikachu.json")
	if err != nil {
		return nil, fmt.Errorf("error loading Pokemon data: %w", err)
	}

	gen, err := logic.Randomize(seed, allowedOnePct, pokemonMap)
	if err != nil {
		return nil, fmt.Errorf("error during randomization: %w", err)
	}

	return output.GenerateSpider(gen, version), nil
}

const defaultTimeout = 30 * time.Second
