package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDeckMissingFile(t *testing.T) {
	deck, err := LoadDeck(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("LoadDeck() returned unexpected error: %v", err)
	}
	if len(deck) != 0 {
		t.Fatalf("LoadDeck() = %v, want empty deck", deck)
	}
}

func TestDeckSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deck.json")

	want := Deck{
		"capital-of-france": {
			Card: Card{Interval: 6, Ease: 2.5, Reps: 2},
			Due:  day(2026, 1, 8),
		},
		"capital-of-peru": {
			Card: Card{Interval: 1, Ease: 2.36, Reps: 1},
			Due:  day(2026, 1, 2),
		},
	}

	if err := want.Save(path); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	got, err := LoadDeck(path)
	if err != nil {
		t.Fatalf("LoadDeck() returned unexpected error: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("LoadDeck() returned %d cards, want %d", len(got), len(want))
	}
	for name, wantCard := range want {
		gotCard, ok := got[name]
		if !ok {
			t.Fatalf("LoadDeck() missing card %q", name)
		}
		if gotCard.Card != wantCard.Card {
			t.Errorf("card %q = %+v, want %+v", name, gotCard.Card, wantCard.Card)
		}
		if !gotCard.Due.Equal(wantCard.Due) {
			t.Errorf("card %q due = %v, want %v", name, gotCard.Due, wantCard.Due)
		}
	}
}

func TestDeckNamesSorted(t *testing.T) {
	deck := Deck{
		"zebra": {Card: Card{Ease: DefaultEase}, Due: day(2026, 1, 1)},
		"apple": {Card: Card{Ease: DefaultEase}, Due: day(2026, 1, 1)},
		"mango": {Card: Card{Ease: DefaultEase}, Due: day(2026, 1, 1)},
	}

	want := []string{"apple", "mango", "zebra"}
	got := deck.Names()
	if len(got) != len(want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Names() = %v, want %v", got, want)
		}
	}
}

func TestLoadDeckBadDueDate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deck.json")
	body := `{"card-a":{"interval":1,"ease":2.5,"reps":1,"due":"not-a-date"}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile() returned unexpected error: %v", err)
	}

	if _, err := LoadDeck(path); err == nil {
		t.Fatal("LoadDeck() expected an error for a malformed due date, got nil")
	}
}

func TestLoadDeckBadJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deck.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned unexpected error: %v", err)
	}

	if _, err := LoadDeck(path); err == nil {
		t.Fatal("LoadDeck() expected an error for malformed JSON, got nil")
	}
}
