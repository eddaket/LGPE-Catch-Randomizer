package pokemon

type PokemonID int

type Pokemon struct {
	ID             PokemonID
	Name           string
	Obtainable     bool
	Requires       []PokemonID
	Excludes       []PokemonID
	OnePct         bool
	OnePctRequires []PokemonID
	RareSpawn      bool
	SilphGift      bool
	PreBrock       bool
}

type PokemonMap map[PokemonID]*Pokemon

type PokemonSet map[PokemonID]bool
