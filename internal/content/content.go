package content

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// DefaultPath assumes guessthedish-data is checked out beside this repo.
// Override it with CONTENT_PATH.
const DefaultPath = "../guessthedish-data/data/puzzles.json"

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Bundle struct {
	Version int      `json:"version"`
	Puzzles []Puzzle `json:"puzzles"`
}

type Puzzle struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	// Family, Cuisine and Difficulty drive progressive assistance and match
	// balancing. They are required but deliberately not checked against a
	// vocabulary here: guessthedish-data owns that list and gates it in its own
	// validator. Duplicating the enum would make the server refuse to boot
	// whenever the catalog adds a cuisine, purely because of deploy order.
	Family     string   `json:"family"`
	Cuisine    string   `json:"cuisine"`
	Difficulty string   `json:"difficulty"`
	Clues      []string `json:"clues"`
}

// Dish is the projection sent to the browser. Family, Cuisine and Difficulty
// are omitted on purpose: the catalog lists every answer, so exposing them
// would let a player narrow the current round by filtering it.
type Dish struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}

func Load(path string) (Bundle, error) {
	f, err := os.Open(path)
	if err != nil {
		return Bundle{}, fmt.Errorf("open content: %w", err)
	}
	defer f.Close()
	return Decode(f)
}

func Decode(r io.Reader) (Bundle, error) {
	var bundle Bundle
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&bundle); err != nil {
		return Bundle{}, fmt.Errorf("decode content: %w", err)
	}
	if err := ensureEOF(dec); err != nil {
		return Bundle{}, err
	}
	if err := Validate(bundle); err != nil {
		return Bundle{}, err
	}
	return bundle, nil
}

func ensureEOF(dec *json.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode content: multiple JSON values")
		}
		return fmt.Errorf("decode content: %w", err)
	}
	return nil
}

func Validate(bundle Bundle) error {
	if bundle.Version != 1 {
		return fmt.Errorf("content version must be 1")
	}
	if len(bundle.Puzzles) < 5 {
		return fmt.Errorf("content must contain at least 5 puzzles")
	}

	ids := make(map[string]struct{}, len(bundle.Puzzles))
	names := make(map[string]string)
	for i, puzzle := range bundle.Puzzles {
		label := fmt.Sprintf("puzzles[%d]", i)
		if !slugPattern.MatchString(puzzle.ID) {
			return fmt.Errorf("%s.id must be a lowercase slug", label)
		}
		if _, exists := ids[puzzle.ID]; exists {
			return fmt.Errorf("%s.id %q is duplicated", label, puzzle.ID)
		}
		ids[puzzle.ID] = struct{}{}
		if err := addAnswerName(names, puzzle.Name, label+".name"); err != nil {
			return err
		}
		for j, alias := range puzzle.Aliases {
			if err := addAnswerName(names, alias, fmt.Sprintf("%s.aliases[%d]", label, j)); err != nil {
				return err
			}
		}
		for _, field := range []struct{ name, value string }{
			{"family", puzzle.Family},
			{"cuisine", puzzle.Cuisine},
			{"difficulty", puzzle.Difficulty},
		} {
			if field.value == "" || field.value != strings.TrimSpace(field.value) {
				return fmt.Errorf("%s.%s must be non-empty and trimmed", label, field.name)
			}
		}
		if len(puzzle.Clues) < 4 || len(puzzle.Clues) > 12 {
			return fmt.Errorf("%s.clues must contain 4-12 entries", label)
		}
		clues := make(map[string]struct{}, len(puzzle.Clues))
		for j, clue := range puzzle.Clues {
			if clue == "" || clue != strings.TrimSpace(clue) {
				return fmt.Errorf("%s.clues[%d] must be non-empty and trimmed", label, j)
			}
			key := strings.ToLower(clue)
			if _, exists := clues[key]; exists {
				return fmt.Errorf("%s.clues[%d] duplicates a clue", label, j)
			}
			clues[key] = struct{}{}
		}
	}
	return nil
}

func addAnswerName(names map[string]string, value, field string) error {
	if value == "" || value != strings.TrimSpace(value) {
		return fmt.Errorf("%s must be non-empty and trimmed", field)
	}
	key := strings.ToLower(value)
	if previous, exists := names[key]; exists {
		return fmt.Errorf("%s duplicates answer name in %s (case-insensitive)", field, previous)
	}
	names[key] = field
	return nil
}

func Catalog(bundle Bundle) []Dish {
	dishes := make([]Dish, len(bundle.Puzzles))
	for i, puzzle := range bundle.Puzzles {
		dishes[i] = Dish{ID: puzzle.ID, Name: puzzle.Name, Aliases: puzzle.Aliases}
	}
	return dishes
}
