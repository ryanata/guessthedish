package content

import (
	"strings"
	"testing"
)

const validContent = `{
  "version": 1,
  "puzzles": [
    {"id":"dish-one","name":"Dish One","aliases":["One"],"family":"soup","cuisine":"American","difficulty":"familiar","clues":["a","b","c","d"]},
    {"id":"dish-two","name":"Dish Two","aliases":["Two"],"family":"soup","cuisine":"American","difficulty":"familiar","clues":["a","b","c","d"]},
    {"id":"dish-three","name":"Dish Three","aliases":["Three"],"family":"soup","cuisine":"American","difficulty":"familiar","clues":["a","b","c","d"]},
    {"id":"dish-four","name":"Dish Four","aliases":["Four"],"family":"soup","cuisine":"American","difficulty":"familiar","clues":["a","b","c","d"]},
    {"id":"dish-five","name":"Dish Five","aliases":["Five"],"family":"soup","cuisine":"American","difficulty":"familiar","clues":["a","b","c","d"]}
  ]
}`

func TestDecodeValid(t *testing.T) {
	bundle, err := Decode(strings.NewReader(validContent))
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Puzzles) != 5 {
		t.Fatalf("got %d puzzles", len(bundle.Puzzles))
	}
}

func TestDecodeRejectsUnknownField(t *testing.T) {
	input := strings.Replace(validContent, `"version": 1`, `"version": 1, "extra": true`, 1)
	if _, err := Decode(strings.NewReader(input)); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestValidateRejectsDuplicateAnswerAndClue(t *testing.T) {
	bundle, err := Decode(strings.NewReader(validContent))
	if err != nil {
		t.Fatal(err)
	}
	bundle.Puzzles[1].Aliases[0] = "dish ONE"
	if err := Validate(bundle); err == nil || !strings.Contains(err.Error(), "duplicates answer") {
		t.Fatalf("expected duplicate answer error, got %v", err)
	}
	bundle.Puzzles[1].Aliases[0] = "Two"
	bundle.Puzzles[0].Clues[3] = "A"
	if err := Validate(bundle); err == nil || !strings.Contains(err.Error(), "duplicates a clue") {
		t.Fatalf("expected duplicate clue error, got %v", err)
	}
}

func TestValidateRejectsContractViolations(t *testing.T) {
	bundle, err := Decode(strings.NewReader(validContent))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*Bundle)
	}{
		{"version", func(b *Bundle) { b.Version = 2 }},
		{"slug", func(b *Bundle) { b.Puzzles[0].ID = "Bad ID" }},
		{"trimmed clue", func(b *Bundle) { b.Puzzles[0].Clues[0] = " a" }},
		{"missing family", func(b *Bundle) { b.Puzzles[0].Family = "" }},
		{"missing cuisine", func(b *Bundle) { b.Puzzles[0].Cuisine = "" }},
		{"missing difficulty", func(b *Bundle) { b.Puzzles[0].Difficulty = "" }},
		{"untrimmed family", func(b *Bundle) { b.Puzzles[0].Family = " soup" }},
		{"minimum puzzles", func(b *Bundle) { b.Puzzles = b.Puzzles[:4] }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copy := bundle
			copy.Puzzles = append([]Puzzle(nil), bundle.Puzzles...)
			tt.mutate(&copy)
			if err := Validate(copy); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
