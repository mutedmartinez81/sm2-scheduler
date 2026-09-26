// Command sm2 computes the next SM-2 spaced repetition schedule for a
// single flashcard given its current state and a review grade. State for
// one card can be passed directly on the command line, or persisted
// across runs in a deck file (see deck.go) so a wrapper script only has
// to remember a card's name. The "due" subcommand (see due.go) lists
// which cards in a deck are ready for review.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

const dateLayout = "2006-01-02"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "due" {
		runDue(os.Args[2:])
		return
	}

	interval := flag.Int("interval", 0, "days since the card was last due (0 for a new card); ignored if -deck is set")
	ease := flag.Float64("ease", DefaultEase, "current ease factor (2.5 for a new card); ignored if -deck is set")
	reps := flag.Int("reps", 0, "consecutive successful reviews so far (0 for a new card); ignored if -deck is set")
	grade := flag.Int("grade", -1, "review grade, 0-5 (required; 0 = blackout, 3 = pass, 5 = perfect)")
	today := flag.String("today", "", "date the review happened, YYYY-MM-DD (default: today)")
	asJSON := flag.Bool("json", false, "print the result as JSON instead of plain text")
	deckPath := flag.String("deck", "", "deck file to load the card's state from and save its new state to")
	cardName := flag.String("card", "", "name of the card within -deck (required if -deck is set)")
	flag.Parse()

	if *grade < 0 {
		fmt.Fprintln(os.Stderr, "error: -grade is required (0-5)")
		flag.Usage()
		os.Exit(2)
	}
	if *deckPath != "" && *cardName == "" {
		fmt.Fprintln(os.Stderr, "error: -card is required when -deck is set")
		flag.Usage()
		os.Exit(2)
	}

	reviewedOn := time.Now().UTC()
	if *today != "" {
		parsed, err := time.Parse(dateLayout, *today)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: -today must look like YYYY-MM-DD: %v\n", err)
			os.Exit(2)
		}
		reviewedOn = parsed
	}

	prev := Card{Interval: *interval, Ease: *ease, Reps: *reps}

	var deck Deck
	if *deckPath != "" {
		var err error
		deck, err = LoadDeck(*deckPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if existing, ok := deck[*cardName]; ok {
			prev = existing.Card
		} else {
			prev = Card{Interval: 0, Ease: DefaultEase, Reps: 0}
		}
	}

	result, err := Review(prev, *grade, reviewedOn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if deck != nil {
		deck[*cardName] = DeckCard{Card: result.Card, Due: result.Due}
		if err := deck.Save(*deckPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	if *asJSON {
		out := struct {
			Interval int     `json:"interval"`
			Ease     float64 `json:"ease"`
			Reps     int     `json:"reps"`
			Due      string  `json:"due"`
		}{
			Interval: result.Card.Interval,
			Ease:     result.Card.Ease,
			Reps:     result.Card.Reps,
			Due:      result.Due.Format(dateLayout),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("interval: %d\n", result.Card.Interval)
	fmt.Printf("ease:     %.2f\n", result.Card.Ease)
	fmt.Printf("reps:     %d\n", result.Card.Reps)
	fmt.Printf("due:      %s\n", result.Due.Format(dateLayout))
}
