# sm2-scheduler

A command-line implementation of the SM-2 spaced repetition algorithm
(the scheduler behind the original SuperMemo and, with tweaks, most
flashcard apps that came after it).

It does one thing: given a card's current scheduling state and how well
you recalled it just now, it tells you the new state and the date the
card is next due. It does not store decks, cards, or history — that part
is your wrapper script's job. This tool is the scheduling math, callable
from a shell script, a cron job, or whatever else is managing your cards.

## Why

Most spaced repetition tools bury the scheduling algorithm inside a GUI
app with its own file format and sync service. Sometimes you just want
the calculation: "I got this one mostly right, when do I see it again?"
That's what this is for.

## Usage

A new card, reviewed for the first time, graded "good" (4 on the 0-5
scale). `-today` fixes the review date so the example is reproducible;
leave it off to schedule against the actual current date:

```
$ sm2 -interval 0 -ease 2.5 -reps 0 -grade 4 -today 2026-01-01
interval: 1
ease:     2.50
reps:     1
due:      2026-01-02
```

The same card, reviewed again a day later and graded "good" again — feed
the previous output back in as input:

```
$ sm2 -interval 1 -ease 2.5 -reps 1 -grade 4 -today 2026-01-02
interval: 6
ease:     2.50
reps:     2
due:      2026-01-08
```

A lapse resets the interval and repetition count, even after a long
streak:

```
$ sm2 -interval 60 -ease 2.6 -reps 6 -grade 1 -today 2026-01-08
interval: 1
ease:     2.06
reps:     0
due:      2026-01-09
```

Pass `-json` to get machine-readable output for a wrapper script instead
of the plain-text form above:

```
$ sm2 -interval 15 -ease 2.5 -reps 3 -grade 5 -today 2026-01-23 -json
{
  "interval": 38,
  "ease": 2.6,
  "reps": 4,
  "due": "2026-03-02"
}
```

## The grade scale

SM-2 uses a 0-5 quality score for how the review went:

| grade | meaning                                  |
|-------|-------------------------------------------|
| 0     | total blackout, no recall at all           |
| 1     | wrong, but the answer felt familiar        |
| 2     | wrong, but it was close                    |
| 3     | correct, but it took real effort           |
| 4     | correct, after a moment's hesitation       |
| 5     | correct, immediate and effortless          |

Grades 0-2 count as a failed review: the interval collapses back to one
day and the repetition streak resets to zero. Grades 3-5 count as a
pass: the repetition streak advances and the interval grows.

## Install

```
go build -o sm2 .
```

## License

MIT, see LICENSE.
