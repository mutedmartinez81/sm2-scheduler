package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// runDue implements the "due" subcommand: it lists cards from a deck file
// that are due for review on or before a given date, most overdue first.
func runDue(args []string) {
	fs := flag.NewFlagSet("due", flag.ExitOnError)
	deckPath := fs.String("deck", "", "deck file to read card state from (required)")
	today := fs.String("today", "", "date to check cards against, YYYY-MM-DD (default: today)")
	asJSON := fs.Bool("json", false, "print the result as JSON instead of plain text")
	fs.Parse(args)

	if *deckPath == "" {
		fmt.Fprintln(os.Stderr, "error: -deck is required")
		fs.Usage()
		os.Exit(2)
	}

	asOf := time.Now().UTC()
	if *today != "" {
		parsed, err := time.Parse(dateLayout, *today)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: -today must look like YYYY-MM-DD: %v\n", err)
			os.Exit(2)
		}
		asOf = parsed
	}

	deck, err := LoadDeck(*deckPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	due := deck.Due(asOf)

	if *asJSON {
		type dueCard struct {
			Name string `json:"name"`
			Due  string `json:"due"`
		}
		out := make([]dueCard, len(due))
		for i, name := range due {
			out[i] = dueCard{Name: name, Due: deck[name].Due.Format(dateLayout)}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(due) == 0 {
		fmt.Println("no cards due")
		return
	}
	for _, name := range due {
		fmt.Printf("%s  due %s\n", name, deck[name].Due.Format(dateLayout))
	}
}
