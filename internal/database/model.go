package database

import (
	"time"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/pokemon"
)

type Generation struct {
	ID            string             `bson:"_id"`
	Seed          int64              `bson:"seed"`
	AllowedOnePct int                `bson:"allowedOnePct"`
	PikachuData   pokemon.PokemonSet `bson:"pikachuData"`
	EeveeData     pokemon.PokemonSet `bson:"eeveeData"`
	CreatedAt     time.Time          `bson:"createdAt"`
}
