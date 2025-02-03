package server

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/database"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/logic"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/output"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/pokemon"
)

const defaultTimeout = 30 * time.Second

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("[ERROR] indexHandler: Error parsing template %v", err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

func (s *Server) randomizeHandler(w http.ResponseWriter, r *http.Request) {
	allowedOnePct, allowedRareSpawn, silphGifts, err := validateRandomizeParams(r)
	if err != nil {
		log.Printf("[WARN] randomizeHandler: Invalid parameters %v", err)
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	id := uuid.NewString()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	seed := logic.GetComputedSeed()

	doneChan := make(chan bool, 1)
	errorChan := make(chan error, 1)
	go func() {
		config := logic.Config{
			Seed:             seed,
			AllowedOnePct:    allowedOnePct,
			AllowedRareSpawn: allowedRareSpawn,
			SilphGifts:       silphGifts,
			PokemonMap:       pokemon.AllPokemon,
		}
		data, err := performRandomization(&config)
		if err != nil {
			errorChan <- err
			return
		}

		generation := database.Generation{
			ID:               id,
			Seed:             seed,
			AllowedOnePct:    allowedOnePct,
			AllowedRareSpawn: allowedRareSpawn,
			SilphGifts:       silphGifts,
			PikachuData:      data.Pikachu,
			EeveeData:        data.Eevee,
			CreatedAt:        time.Now(),
		}

		err = s.db.InsertGeneration(&generation)
		if err != nil {
			errorChan <- err
			return
		}

		doneChan <- true
	}()

	select {
	case <-ctx.Done():
		log.Printf("[ERROR] randomizeHandler: Timeout")
		http.Error(w, "Randomization timed out", http.StatusGatewayTimeout)
		return
	case err := <-errorChan:
		log.Printf("[ERROR] randomizeHandler: Error encountered %v", err)
		http.Error(w, "Error during randomization", http.StatusInternalServerError)
		return
	case <-doneChan:
		http.Redirect(w, r, fmt.Sprintf("/seed/%s", id), http.StatusSeeOther)
		log.Printf("[INFO] randomizeHandler: Success - ID=%s, Seed=%d", id, seed)
	}
}

func (s *Server) seedHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	generation, err := s.db.GetGenerationById(id)
	if err != nil {
		log.Printf("[ERROR] seedHandler: Error getting generation %v", err)
		http.Error(w, "Error retrieving generation", http.StatusInternalServerError)
		return
	}

	if generation == nil {
		http.Error(w, "Generation not found", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/seed.html")
	if err != nil {
		log.Printf("[ERROR] seedHandler: Error parsing template %v", err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}

	type SeedData struct {
		Settings struct {
			AllowedOnePct    string
			AllowedRareSpawn string
			SilphGifts       string
		}
		DownloadURL string
		SeedID      string
		SeedURL     string
		TimeStamp   string
	}

	settings := struct {
		AllowedOnePct    string
		AllowedRareSpawn string
		SilphGifts       string
	}{}

	if generation.AllowedOnePct > 3 {
		settings.AllowedOnePct = "Unlimited"
	} else {
		settings.AllowedOnePct = fmt.Sprint(generation.AllowedOnePct)
	}

	if generation.AllowedRareSpawn > 3 {
		settings.AllowedRareSpawn = "Unlimited"
	} else {
		settings.AllowedRareSpawn = fmt.Sprint(generation.AllowedRareSpawn)
	}

	if generation.SilphGifts {
		settings.SilphGifts = "Yes"
	} else {
		settings.SilphGifts = "No"
	}

	baseURL := getBaseURL(r)
	data := SeedData{
		Settings:    settings,
		DownloadURL: fmt.Sprintf("/seed/%s/tracker", id),
		SeedID:      id,
		SeedURL:     fmt.Sprintf("%s/seed/%s", baseURL, id),
		TimeStamp:   generation.CreatedAt.Format(time.RFC1123),
	}
	tmpl.Execute(w, data)
}

func (s *Server) seedTrackerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	generation, err := s.db.GetGenerationById(id)
	if err != nil {
		log.Printf("[ERROR] seedTrackerHandler: Error getting generation %v", err)
		http.Error(w, "Error retrieving generation", http.StatusInternalServerError)
		return
	}

	if generation == nil {
		http.Error(w, "Generation not found", http.StatusNotFound)
		return
	}

	query := r.URL.Query()
	version := query.Get("version")
	if version != "pikachu" && version != "eevee" {
		http.Error(w, "Unsupported version", http.StatusBadRequest)
		return
	}

	input, err := output.DecodeSpider(r.Body)
	if err != nil {
		log.Printf("[WARN] seedTrackerHandler: Unable to read body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	lg := logic.Generation{
		Seed:          generation.Seed,
		AllowedOnePct: generation.AllowedOnePct,
		Pikachu:       generation.PikachuData,
		Eevee:         generation.EeveeData,
	}
	data := output.GenerateSpider(&lg, version, input)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Printf("[ERROR] seedTrackerHandler: Error writing JSON %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}
