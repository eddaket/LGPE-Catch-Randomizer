package main

import (
	"log"
	"time"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/logic"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/pokemon"
)

func main() {
	allowedOnePcts := []int{0, 1, 2, 3, 50}
	allowedRareSpawns := []int{0, 1, 2, 3, 50}
	silphGifts := []bool{true, false}

	for i := 1; i <= 1000; i++ {
		for _, aop := range allowedOnePcts {
			for _, ars := range allowedRareSpawns {
				for _, sg := range silphGifts {
					seed := logic.GetComputedSeed()
					config := logic.Config{
						Seed:             seed,
						AllowedOnePct:    aop,
						AllowedRareSpawn: ars,
						SilphGifts:       sg,
						PokemonMap:       pokemon.AllPokemon,
					}
					_, err := logic.Randomize(&config)
					if err != nil {
						log.Printf("[ERROR] Error encountered: %v %v", err, config)
					}

					// Sleep to make sure we get a new seed on the next loop
					time.Sleep(time.Millisecond)
				}
			}
		}

		if i%100 == 0 {
			log.Printf("[INFO] Iteration %d complete", i)
		}
	}

	log.Printf("[INFO] Stress complete")
}
