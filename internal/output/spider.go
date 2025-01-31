package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/eddaket/LGPE-Catch-Randomizer/internal/logic"
	"github.com/eddaket/LGPE-Catch-Randomizer/internal/pokemon"
)

type SpiderTracker struct {
	Name            string        `json:"name,omitempty"`
	TextColor       string        `json:"textColor,omitempty"`
	BackgroundColor string        `json:"backgroundColor,omitempty"`
	BonusColor      string        `json:"bonusColor,omitempty"`
	PlannedColor    string        `json:"plannedColor,omitempty"`
	MarkedColor     string        `json:"markedColor,omitempty"`
	PokesPerLine    int           `json:"pokesPerLine,omitempty"`
	TrackerStyle    string        `json:"trackerStyle,omitempty"`
	Pokes           []SpiderPokes `json:"pokes"`
}

type SpiderPokes struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

func DecodeSpider(r io.Reader) (*SpiderTracker, error) {
	// 1<<20 == 1MB
	lr := io.LimitReader(r, 1<<20)

	var inputSpider SpiderTracker
	err := json.NewDecoder(lr).Decode(&inputSpider)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	return &inputSpider, nil
}

func GenerateSpider(gen *logic.Generation, version string, input *SpiderTracker) []byte {
	set := gen.Pikachu
	if version == "eevee" {
		set = gen.Eevee
	}

	out := SpiderTracker{
		Name:            input.Name,
		TextColor:       input.TextColor,
		BackgroundColor: input.BackgroundColor,
		BonusColor:      input.BonusColor,
		PlannedColor:    input.PlannedColor,
		MarkedColor:     input.MarkedColor,
		PokesPerLine:    input.PokesPerLine,
		TrackerStyle:    input.TrackerStyle,
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
