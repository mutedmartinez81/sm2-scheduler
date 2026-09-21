// Command sm2 computes the next SM-2 spaced repetition schedule for a
// single flashcard given its current state and a review grade. It does
// not manage a deck or a database; it is meant to be called from a
// wrapper script that stores card state (e.g. one line per card in a
// text file or a small JSON blob) and shells out here to advance it.
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
	interval := flag.Int("interval", 0, "days since the card was last due (0 for a new card)")
	ease := flag.Float64("ease", DefaultEase, "current ease factor (2.5 for a new card)")
	reps := flag.Int("reps", 0, "consecutive successful reviews so far (0 for a new card)")
	grade := flag.Int("grade", -1, "review grade, 0-5 (required; 0 = blackout, 3 = pass, 5 = perfect)")
	today := flag.String("today", "", "date the review happened, YYYY-MM-DD (default: today)")
	asJSON := flag.Bool("json", false, "print the result as JSON instead of plain text")
	flag.Parse()

	if *grade < 0 {
		fmt.Fprintln(os.Stderr, "error: -grade is required (0-5)")
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

	result, err := Review(Card{Interval: *interval, Ease: *ease, Reps: *reps}, *grade, reviewedOn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
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
