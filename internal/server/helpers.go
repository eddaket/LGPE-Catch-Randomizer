package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/logic"
)

func validateRandomizeParams(r *http.Request) (allowedOnePct int, allowedRareSpawn int, silphGifts bool, err error) {
	query := r.URL.Query()
	allowedOnePctStr := query.Get("allowedOnePct")
	allowedRareSpawnStr := query.Get("allowedRareSpawn")
	silphGiftsStr := query.Get("silphGifts")

	if allowedOnePctStr == "unlimited" {
		allowedOnePct = 50
	} else {
		allowedOnePct, err = strconv.Atoi(allowedOnePctStr)
		if err != nil || allowedOnePct < 0 {
			return 0, 0, false, fmt.Errorf("invalid allowedOnePct: %s", allowedOnePctStr)
		}
	}

	if allowedRareSpawnStr == "unlimited" {
		allowedRareSpawn = 50
	} else {
		allowedRareSpawn, err = strconv.Atoi(allowedRareSpawnStr)
		if err != nil || allowedRareSpawn < 0 {
			return 0, 0, false, fmt.Errorf("invalid allowedRareSpawn: %s", allowedRareSpawnStr)
		}
	}

	silphGifts, err = strconv.ParseBool(silphGiftsStr)
	if err != nil {
		return 0, 0, false, fmt.Errorf("invalid silphGifts: %s", silphGiftsStr)
	}

	return
}

func performRandomization(config *logic.Config) (*logic.Generation, error) {
	gen, err := logic.Randomize(config)
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
