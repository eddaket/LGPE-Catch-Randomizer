package pokemon

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func LoadPokemonData(filename string) (map[PokemonID]*Pokemon, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var pokemonList []Pokemon
	if err := json.Unmarshal(bytes, &pokemonList); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	pokemonMap := make(map[PokemonID]*Pokemon)
	for _, pokemon := range pokemonList {
		pokemonMap[pokemon.ID] = &pokemon
	}

	return pokemonMap, nil
}
