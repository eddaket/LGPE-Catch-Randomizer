package output

import (
	"encoding/json"
	"fmt"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/logic"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/pokemon"
)

type SpiderTracker struct {
	Name            string        `json:"name"`
	TextColor       string        `json:"textColor"`
	BackgroundColor string        `json:"backgroundColor"`
	BonusColor      string        `json:"bonusColor"`
	PlannedColor    string        `json:"plannedColor"`
	MarkedColor     string        `json:"markedColor"`
	PokesPerLine    int           `json:"pokesPerLine"`
	TrackerStyle    string        `json:"trackerStyle"`
	Pokes           []SpiderPokes `json:"pokes"`
}

type SpiderPokes struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

func GenerateSpider(gen *logic.Generation, version string) []byte {
	set := gen.Pikachu
	if version == "Eevee" {
		set = gen.Eevee
	}

	// TODO: Allow customization of these fields. For now you all get my settings
	out := SpiderTracker{
		Name:            fmt.Sprintf("Catch Rando %d", gen.Seed),
		TextColor:       "#ffffff",
		BackgroundColor: "#000000",
		BonusColor:      "#27272a",
		PlannedColor:    "#60a5fa",
		MarkedColor:     "#fbbf24",
		PokesPerLine:    15,
		TrackerStyle:    "bobchao",
		Pokes:           []SpiderPokes{},
	}

	for i := 1; i < 151; i++ {
		id := i
		status := "unmarked"
		if !set[pokemon.PokemonID(i)] {
			status = "hidden"
		}
		out.Pokes = append(out.Pokes, SpiderPokes{ID: id, Status: status})
	}

	outJson, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		panic(err)
	}

	return outJson
}
