package logic

import "github.com/eddaket/LGPE-Catch-Randomizer/internal/pokemon"

type Config struct {
	Seed             int64
	AllowedOnePct    int
	AllowedRareSpawn int
	SilphGifts       bool
	PokemonMap       pokemon.PokemonMap
}
