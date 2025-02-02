package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/logic"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/pokemon"
)

func validateRandomizeParams(r *http.Request) (int, error) {
	query := r.URL.Query()
	allowedOnePctStr := query.Get("allowedOnePct")

	var err error
	var allowedOnePct int
	if allowedOnePctStr == "unlimited" {
		allowedOnePct = 50
	} else {
		allowedOnePct, err = strconv.Atoi(allowedOnePctStr)
		if err != nil || allowedOnePct < 0 {
			return 0, fmt.Errorf("invalid allowedOnePct: %s", allowedOnePctStr)
		}
	}

	return allowedOnePct, nil
}

func performRandomization(seed int64, allowedOnePct int) (*logic.Generation, error) {
	gen, err := logic.Randomize(seed, allowedOnePct, pokemon.AllPokemon)
	if err != nil {
		return nil, fmt.Errorf("error during randomization: %w", err)
	}

	return gen, nil
}

func getBaseURL(r *http.Request) string {
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s", protocol, r.Host)
}
