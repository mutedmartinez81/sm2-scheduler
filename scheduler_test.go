package main

import (
	"math"
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestReview(t *testing.T) {
	tests := []struct {
		name       string
		prev       Card
		grade      int
		reviewedOn time.Time
		wantCard   Card
		wantDue    time.Time
	}{
		{
			// A brand new card's first review always gets a 1-day interval,
			// regardless of the ease factor it starts with.
			name:       "first review passing",
			prev:       Card{Interval: 0, Ease: DefaultEase, Reps: 0},
			grade:      4,
			reviewedOn: day(2026, 1, 1),
			wantCard:   Card{Interval: 1, Ease: 2.5, Reps: 1},
			wantDue:    day(2026, 1, 2),
		},
		{
			// The second successful review is fixed at 6 days by SM-2,
			// again independent of the ease factor.
			name:       "second review passing",
			prev:       Card{Interval: 1, Ease: 2.5, Reps: 1},
			grade:      4,
			reviewedOn: day(2026, 1, 2),
			wantCard:   Card{Interval: 6, Ease: 2.5, Reps: 2},
			wantDue:    day(2026, 1, 8),
		},
		{
			// From the third review on, interval = round(prevInterval * ease).
			name:       "third review multiplies by ease",
			prev:       Card{Interval: 6, Ease: 2.5, Reps: 2},
			grade:      4,
			reviewedOn: day(2026, 1, 8),
			wantCard:   Card{Interval: 15, Ease: 2.5, Reps: 3},
			wantDue:    day(2026, 1, 23),
		},
		{
			// A grade of 3 is the lowest passing grade: reps still advance,
			// but the ease factor takes a noticeable hit.
			name:       "boundary passing grade 3 still advances reps",
			prev:       Card{Interval: 6, Ease: 2.5, Reps: 2},
			grade:      3,
			reviewedOn: day(2026, 1, 8),
			wantCard:   Card{Interval: 15, Ease: 2.36, Reps: 3},
			wantDue:    day(2026, 1, 23),
		},
		{
			// A grade of 2 is the highest failing grade: reps and interval
			// reset even though the recall was close to correct.
			name:       "boundary failing grade 2 resets reps",
			prev:       Card{Interval: 15, Ease: 2.5, Reps: 3},
			grade:      2,
			reviewedOn: day(2026, 1, 23),
			wantCard:   Card{Interval: 1, Ease: 2.18, Reps: 0},
			wantDue:    day(2026, 1, 24),
		},
		{
			// A total blackout drives the ease factor down hard; it must
			// clamp at MinEase rather than go lower or negative.
			name:       "total blackout floors ease",
			prev:       Card{Interval: 1, Ease: 1.35, Reps: 0},
			grade:      0,
			reviewedOn: day(2026, 1, 1),
			wantCard:   Card{Interval: 1, Ease: 1.3, Reps: 0},
			wantDue:    day(2026, 1, 2),
		},
		{
			// A perfect recall raises the ease factor by exactly 0.1.
			name:       "perfect recall raises ease",
			prev:       Card{Interval: 15, Ease: 2.5, Reps: 3},
			grade:      5,
			reviewedOn: day(2026, 1, 23),
			wantCard:   Card{Interval: 38, Ease: 2.6, Reps: 4},
			wantDue:    day(2026, 3, 2),
		},
		{
			// A lapse after a long streak drops all the accumulated
			// interval, not just a fraction of it.
			name:       "lapse after long streak resets interval",
			prev:       Card{Interval: 50, Ease: 2.7, Reps: 5},
			grade:      1,
			reviewedOn: day(2026, 1, 1),
			wantCard:   Card{Interval: 1, Ease: 2.16, Reps: 0},
			wantDue:    day(2026, 1, 2),
		},
		{
			// interval * ease landing exactly on a half-integer must round
			// away from zero (8.5 -> 9), not truncate down to 8.
			name:       "half integer interval rounds up",
			prev:       Card{Interval: 4, Ease: 2.125, Reps: 2},
			grade:      3,
			reviewedOn: day(2026, 1, 1),
			wantCard:   Card{Interval: 9, Ease: 1.985, Reps: 3},
			wantDue:    day(2026, 1, 10),
		},
		{
			// Due dates must roll over a year boundary correctly.
			name:       "due date crosses year boundary",
			prev:       Card{Interval: 1, Ease: 2.5, Reps: 1},
			grade:      4,
			reviewedOn: day(2025, 12, 30),
			wantCard:   Card{Interval: 6, Ease: 2.5, Reps: 2},
			wantDue:    day(2026, 1, 5),
		},
		{
			// Due dates must roll over a leap-year February correctly.
			name:       "due date crosses leap day",
			prev:       Card{Interval: 6, Ease: 2.5, Reps: 2},
			grade:      4,
			reviewedOn: day(2028, 2, 25),
			wantCard:   Card{Interval: 15, Ease: 2.5, Reps: 3},
			wantDue:    day(2028, 3, 11),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Review(tt.prev, tt.grade, tt.reviewedOn)
			if err != nil {
				t.Fatalf("Review() returned unexpected error: %v", err)
			}
			if got.Card.Interval != tt.wantCard.Interval {
				t.Errorf("Interval = %d, want %d", got.Card.Interval, tt.wantCard.Interval)
			}
			if !almostEqual(got.Card.Ease, tt.wantCard.Ease) {
				t.Errorf("Ease = %.4f, want %.4f", got.Card.Ease, tt.wantCard.Ease)
			}
			if got.Card.Reps != tt.wantCard.Reps {
				t.Errorf("Reps = %d, want %d", got.Card.Reps, tt.wantCard.Reps)
			}
			if !got.Due.Equal(tt.wantDue) {
				t.Errorf("Due = %v, want %v", got.Due, tt.wantDue)
			}
		})
	}
}

func TestReviewInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		prev Card
		grade int
	}{
		{"grade below range", Card{Ease: DefaultEase}, -1},
		{"grade above range", Card{Ease: DefaultEase}, 6},
		{"zero ease", Card{Ease: 0}, 4},
		{"negative ease", Card{Ease: -1.3}, 4},
		{"negative interval", Card{Ease: DefaultEase, Interval: -1}, 4},
		{"negative reps", Card{Ease: DefaultEase, Reps: -1}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Review(tt.prev, tt.grade, day(2026, 1, 1))
			if err == nil {
				t.Fatal("Review() expected an error, got nil")
			}
		})
	}
}
