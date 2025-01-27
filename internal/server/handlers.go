package server

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/database"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/logic"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/output"
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
	allowedOnePct, err := validateRandomizeParams(r)
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
		data, err := performRandomization(seed, allowedOnePct)
		if err != nil {
			errorChan <- err
			return
		}

		generation := database.Generation{
			ID:            id,
			Seed:          seed,
			AllowedOnePct: allowedOnePct,
			PikachuData:   data.Pikachu,
			EeveeData:     data.Eevee,
			CreatedAt:     time.Now(),
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
		DownloadURL string
		SeedURL     string
		TimeStamp   string
	}

	baseURL := getBaseURL(r)
	data := SeedData{
		DownloadURL: fmt.Sprintf("/download/%s", id),
		SeedURL:     fmt.Sprintf("%s/seed/%s", baseURL, id),
		TimeStamp:   generation.CreatedAt.Format(time.RFC1123),
	}
	tmpl.Execute(w, data)
}

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	version := query.Get("version")
	if version != "pikachu" && version != "eevee" {
		http.Error(w, "Unsupported version", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	generation, err := s.db.GetGenerationById(id)
	if err != nil {
		log.Printf("[ERROR] generationHandler: Error getting generation %v", err)
		http.Error(w, "Error retrieving generation", http.StatusInternalServerError)
		return
	}

	if generation == nil {
		http.Error(w, "Generation not found", http.StatusNotFound)
		return
	}

	lg := logic.Generation{
		Seed:          generation.Seed,
		AllowedOnePct: generation.AllowedOnePct,
		Pikachu:       generation.PikachuData,
		Eevee:         generation.EeveeData,
	}
	data := output.GenerateSpider(&lg, version)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s_%d.json", "Catch_Rando", cases.Title(language.English).String(version), generation.Seed))
	_, err = w.Write(data)
	if err != nil {
		log.Printf("[ERROR] generationHandler: Error writing JSON %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}
