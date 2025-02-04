package logic

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/pokemon"
)

func GetComputedSeed() int64 {
	return time.Now().UnixMilli()
}

const maxIncluded = 50

type Generation struct {
	pokemonMap     pokemon.PokemonMap
	included       pokemon.PokemonSet
	banned         pokemon.PokemonSet
	requireBucket  pokemon.PokemonSet
	onePct         pokemon.PokemonSet
	onePctCount    int
	onePctBucket   pokemon.PokemonSet
	rareSpawn      pokemon.PokemonSet
	rareSpawnCount int
	randomization  *rand.Rand

	Seed             int64
	AllowedOnePct    int
	AllowedRareSpawn int
	SilphGifts       bool
	Pikachu          pokemon.PokemonSet
	Eevee            pokemon.PokemonSet
}

func Randomize(config *Config) (*Generation, error) {
	g := &Generation{
		pokemonMap:     config.PokemonMap,
		included:       make(pokemon.PokemonSet),
		banned:         make(pokemon.PokemonSet),
		requireBucket:  make(pokemon.PokemonSet),
		onePct:         make(pokemon.PokemonSet),
		onePctCount:    0,
		onePctBucket:   make(pokemon.PokemonSet),
		rareSpawn:      make(pokemon.PokemonSet),
		rareSpawnCount: 0,
		randomization:  rand.New(rand.NewSource(config.Seed)),

		Seed:             config.Seed,
		AllowedOnePct:    config.AllowedOnePct,
		AllowedRareSpawn: config.AllowedRareSpawn,
		SilphGifts:       config.SilphGifts,
	}

	err := g.generate()
	if err != nil {
		return nil, err
	}

	g.populateVersions()

	return g, nil
}

func (g *Generation) generate() error {
	// Always start with Pikachu
	g.attemptAdd(PikachuID)

	// Fill most of the way, to allow for breathing room in Pre-Brock logic
	g.fillToLimit(40)

	err := g.ensureGrassForBrock()
	if err != nil {
		return err
	}

	g.fillToLimit(50)
	return nil
}

func (g *Generation) fillToLimit(maxFill int) {
	for len(g.included) < maxFill {
		pokemonID := pokemon.PokemonID(g.randomization.Intn(len(g.pokemonMap)) + 1)
		g.attemptAdd(pokemonID)
	}
}

func (g *Generation) attemptAdd(id pokemon.PokemonID) bool {
	// We hit the cap, just return out
	if len(g.included) >= maxIncluded {
		return false
	}

	// This pokemon has already been processed
	// When we attempt to re-add a Pokemon from a bucket, we first delete them from the bucket
	if g.included[id] || g.banned[id] || g.requireBucket[id] || g.onePctBucket[id] {
		return false
	}

	pokemon := g.pokemonMap[id]
	if !pokemon.Obtainable {
		return false
	}

	// Pokemon has outstanding requirements, put them in the bucket and move on
	if !g.isFulfilled(pokemon.Requires) {
		g.requireBucket[id] = true
		return false
	}

	// Pokemon is a 1% but we don't have room, put them in the bucket and move on
	if g.handleOnePct(pokemon) {
		g.onePctBucket[id] = true
		return false
	}

	// Pokemon is a rare spawn but we don't have room. No bucket, just move on
	if g.handleRareSpawn(pokemon) {
		return false
	}

	// Include this Pokemon. Ban anything that needs to be banned
	g.included[id] = true
	for _, banId := range pokemon.Excludes {
		g.banned[banId] = true
	}

	// Re-evaluate the 1%s. If their other requirements are met, their slot can be freed
	for onePctId := range g.onePct {
		if g.isFulfilled(g.pokemonMap[onePctId].OnePctRequires) {
			delete(g.onePct, onePctId)
			g.onePctCount -= 1
			if onePctId == DragonairID {
				g.onePctCount -= 1
			}
		}
	}

	// Now that this Pokemon is added, we can re-evaluate the OnePct bucket and the overall Bucket
	g.processBucket(g.requireBucket)
	g.processBucket(g.onePctBucket)

	return true
}

func (g *Generation) handleOnePct(pokemon *pokemon.Pokemon) bool {
	// Pokemon's 1% requirements aren't met, so we treat it like a 1%
	if pokemon.OnePct && !g.isFulfilled(pokemon.OnePctRequires) {
		count := 1 + g.onePctCount

		// Special case for Dragonair
		if pokemon.ID == DragonairID {
			count += 1
		}

		// Too many 1%s
		if count > g.AllowedOnePct {
			return true
		}

		// Otherwise we're counting it
		g.onePctCount = count
		g.onePct[pokemon.ID] = true
	}
	return false
}

func (g *Generation) handleRareSpawn(pokemon *pokemon.Pokemon) bool {
	// Pokemon's a rare spawn and either not a Silph Gift, or Silph Gifts are included
	if pokemon.RareSpawn && (!pokemon.SilphGift || g.SilphGifts) {
		count := 1 + g.rareSpawnCount

		// Too many rare spawns
		if count > g.AllowedRareSpawn {
			return true
		}

		// Otherwise, we're counting it
		g.rareSpawnCount = count
		g.rareSpawn[pokemon.ID] = true
	}
	return false
}

func (g *Generation) processBucket(bucket pokemon.PokemonSet) {
	for {
		anyAdded := false
		for i := 1; i < 151; i++ {
			id := pokemon.PokemonID(i)
			if !bucket[id] {
				continue
			}

			delete(bucket, id)
			if g.attemptAdd(id) {
				anyAdded = true
			}
		}

		if !anyAdded {
			break
		}
	}
}

func (g *Generation) isFulfilled(requirements []pokemon.PokemonID) bool {
	fulfilled := true
	for _, reqId := range requirements {
		if !g.included[reqId] {
			fulfilled = false
			break
		}
	}
	return fulfilled
}

func (g *Generation) ensureGrassForBrock() error {
	chainable := false
	for id := range g.included {
		pokemon := g.pokemonMap[id]
		if pokemon.PreBrock {
			chainable = true
		}
	}

	if !g.preBrockSatisfied(chainable) {
		prospects := []pokemon.PokemonID{OddishID}
		if chainable && g.rareSpawnCount < g.AllowedRareSpawn {
			prospects = append(prospects, BulbasaurID)
		}

		i := g.randomization.Intn(len(prospects))
		g.attemptAdd(prospects[i])

		// If it's still unsatisfied, something went terribly wrong
		if !g.preBrockSatisfied(chainable) {
			return fmt.Errorf("unable to add grass before brock")
		}
	}

	return nil
}

func (g *Generation) preBrockSatisfied(chainable bool) bool {
	return g.included[OddishID] || (g.included[BulbasaurID] && chainable)
}

var ConvertMap = map[pokemon.PokemonID]pokemon.PokemonID{
	25:  133, // Pikachu to Eevee
	26:  135, // Raichu to Jolteon
	27:  23,  // Sandshrew to Ekans
	28:  24,
	43:  69, // Oddish to Bellsprout
	44:  70,
	45:  71,
	53:  59, // Persian to Arcanine
	56:  37, // Mankey to Vulpix
	57:  38,
	58:  52, // Growlithe to Meowth
	59:  53,
	88:  109, // Grimer to Koffing
	89:  110,
	123: 127, // Scyther to Pinsir
	133: 25,  // Eevee to Pikachu
	135: 26,  // Jolteon to Raichu
}

func (g *Generation) populateVersions() {
	g.Pikachu = g.included
	g.Eevee = make(pokemon.PokemonSet, len(g.Pikachu))
	for pikaId := range g.Pikachu {
		if eeveeId, ok := ConvertMap[pikaId]; ok {
			g.Eevee[eeveeId] = true
		} else {
			g.Eevee[pikaId] = true
		}
	}
}

const (
	PikachuID   pokemon.PokemonID = 25
	BulbasaurID pokemon.PokemonID = 1
	OddishID    pokemon.PokemonID = 43
	DragonairID pokemon.PokemonID = 148
)
