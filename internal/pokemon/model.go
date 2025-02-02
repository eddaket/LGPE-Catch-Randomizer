package pokemon

type PokemonID int

type Pokemon struct {
	ID             PokemonID   `json:"id"`
	Name           string      `json:"name"`
	Obtainable     bool        `json:"obtainable"`
	OnePct         bool        `json:"onePct"`
	Excludes       []PokemonID `json:"excludes"`
	OnePctRequires []PokemonID `json:"onePctRequires"`
	Requires       []PokemonID `json:"requires"`
	PreBrock       bool        `json:"preBrock"`
}

type PokemonMap map[PokemonID]*Pokemon

type PokemonSet map[PokemonID]bool
