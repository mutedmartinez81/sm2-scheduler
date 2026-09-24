package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// DeckCard is the state persisted for one card between runs: its SM-2
// scheduling state plus the date it's next due.
type DeckCard struct {
	Card
	Due time.Time
}

// Deck is a set of cards keyed by name, persisted as a single JSON file so
// a wrapper script can track many cards across runs without a database.
type Deck map[string]DeckCard

// deckCardJSON is the on-disk shape of a DeckCard. It exists separately
// from DeckCard because Due needs to round-trip as a plain date string,
// not Go's default RFC 3339 timestamp.
type deckCardJSON struct {
	Interval int     `json:"interval"`
	Ease     float64 `json:"ease"`
	Reps     int     `json:"reps"`
	Due      string  `json:"due"`
}

// LoadDeck reads a deck from path. A missing file is not an error; it
// returns an empty deck so a new deck can be built up one review at a
// time, starting from a wrapper script that hasn't created the file yet.
func LoadDeck(path string) (Deck, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Deck{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading deck %s: %w", path, err)
	}

	var raw map[string]deckCardJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing deck %s: %w", path, err)
	}

	deck := make(Deck, len(raw))
	for name, c := range raw {
		due, err := time.Parse(dateLayout, c.Due)
		if err != nil {
			return nil, fmt.Errorf("parsing due date for card %q: %w", name, err)
		}
		deck[name] = DeckCard{
			Card: Card{Interval: c.Interval, Ease: c.Ease, Reps: c.Reps},
			Due:  due,
		}
	}
	return deck, nil
}

// Save writes the deck to path as indented JSON, sorted by card name so
// the file diffs cleanly if it's kept under version control.
func (d Deck) Save(path string) error {
	names := d.Names()
	// encoding/json sorts map keys itself, but building an explicit,
	// pre-sorted intermediate keeps the on-disk shape obvious from this
	// function alone rather than relying on that implementation detail.
	out := make(map[string]deckCardJSON, len(names))
	for _, name := range names {
		c := d[name]
		out[name] = deckCardJSON{
			Interval: c.Interval,
			Ease:     c.Ease,
			Reps:     c.Reps,
			Due:      c.Due.Format(dateLayout),
		}
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding deck: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing deck %s: %w", path, err)
	}
	return nil
}

// Names returns the deck's card names in sorted order.
func (d Deck) Names() []string {
	names := make([]string, 0, len(d))
	for name := range d {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
