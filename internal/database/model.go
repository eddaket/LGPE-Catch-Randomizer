package database

import (
	"time"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/logic"
)

type Generation struct {
	ID            string           `bson:"_id"`
	Seed          int64            `bson:"seed"`
	AllowedOnePct int              `bson:"allowedOnePct"`
	PikachuData   logic.PokemonSet `bson:"pikachuData"`
	EeveeData     logic.PokemonSet `bson:"eeveeData"`
	CreatedAt     time.Time        `bson:"createdAt"`
}
