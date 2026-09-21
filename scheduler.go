// Package main implements the SM-2 spaced repetition scheduling algorithm
// as described in P.A. Wozniak's original SuperMemo 2 paper. This file
// holds the pure algorithm; main.go wires it to the command line.
package main

import (
	"fmt"
	"math"
	"time"
)

// Card holds the scheduling state carried between reviews. Reps counts
// consecutive reviews graded 3 or higher; a single grade below 3 resets it.
type Card struct {
	Interval int     // days until the review that just happened was due
	Ease     float64 // ease factor going into this review
	Reps     int     // consecutive successful reps going into this review
}

// DefaultEase is the ease factor assigned to a card that has never been
// reviewed. SM-2 starts every card here.
const DefaultEase = 2.5

// MinEase is the floor SM-2 places on the ease factor. Without it a run of
// poor grades can drive the ease so low that intervals collapse to nothing
// and never recover.
const MinEase = 1.3

// Result is the state to carry forward after a review, plus the date on
// which the card is next due.
type Result struct {
	Card Card
	Due  time.Time
}

// Review applies one graded review to a card and returns its next
// scheduling state. grade must be in [0,5] using the SM-2 quality scale,
// where 0 is a total blackout and 5 is a perfect, effortless recall;
// anything below 3 counts as a failed review.
func Review(prev Card, grade int, reviewedOn time.Time) (Result, error) {
	if grade < 0 || grade > 5 {
		return Result{}, fmt.Errorf("grade %d out of range: must be 0-5", grade)
	}
	if prev.Ease <= 0 {
		return Result{}, fmt.Errorf("ease %.2f out of range: must be > 0", prev.Ease)
	}
	if prev.Interval < 0 {
		return Result{}, fmt.Errorf("interval %d out of range: must be >= 0", prev.Interval)
	}
	if prev.Reps < 0 {
		return Result{}, fmt.Errorf("reps %d out of range: must be >= 0", prev.Reps)
	}

	var next Card

	if grade < 3 {
		// A failed review always restarts the graduation ladder, even if
		// the card had a long interval before this lapse.
		next.Reps = 0
		next.Interval = 1
	} else {
		next.Reps = prev.Reps + 1
		switch prev.Reps {
		case 0:
			next.Interval = 1
		case 1:
			next.Interval = 6
		default:
			// The interval multiplier uses the ease factor as it stood
			// going into this review, not the value updated below.
			next.Interval = int(math.Round(float64(prev.Interval) * prev.Ease))
			if next.Interval < 1 {
				next.Interval = 1
			}
		}
	}

	// The ease update runs for every grade, pass or fail, per the original
	// SM-2 formula: EF' = EF + (0.1 - (5-q)*(0.08 + (5-q)*0.02)).
	q := float64(grade)
	next.Ease = prev.Ease + (0.1 - (5-q)*(0.08+(5-q)*0.02))
	if next.Ease < MinEase {
		next.Ease = MinEase
	}

	return Result{
		Card: next,
		Due:  reviewedOn.AddDate(0, 0, next.Interval),
	}, nil
}
