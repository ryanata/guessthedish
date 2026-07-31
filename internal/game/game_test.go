package game

import (
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"guessthedish/internal/content"
)

func testPuzzles() []content.Puzzle {
	return []content.Puzzle{
		{ID: "dish-one", Name: "Dish One", Family: "soup", Cuisine: "Thai", Clues: []string{"a", "b", "c", "d"}},
		{ID: "dish-two", Name: "Dish Two", Family: "soup", Cuisine: "Thai", Clues: []string{"a", "b", "c", "d"}},
		{ID: "dish-three", Name: "Dish Three", Family: "soup", Cuisine: "Thai", Clues: []string{"a", "b", "c", "d"}},
		{ID: "dish-four", Name: "Dish Four", Family: "soup", Cuisine: "Thai", Clues: []string{"a", "b", "c", "d"}},
		{ID: "dish-five", Name: "Dish Five", Family: "soup", Cuisine: "Thai", Clues: []string{"a", "b", "c", "d"}},
	}
}

func newTestStore(t *testing.T) (*Store, *time.Time, string, string, string) {
	t.Helper()
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	s := NewStore(testPuzzles())
	s.now = func() time.Time { return now }
	first, err := s.Join("Cook")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Join("Chef")
	if err != nil {
		t.Fatal(err)
	}
	return s, &now, first.ID, first.Token, second.Token
}

func TestJoinWaitsThenStartsWithSecondHuman(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	s := NewStore(testPuzzles())
	s.now = func() time.Time { return now }
	first, err := s.Join("Cook")
	if err != nil {
		t.Fatal(err)
	}
	if first.Phase != "waiting" || first.Round != 0 || len(first.Clues) != 0 || first.Token == "" {
		t.Fatalf("unexpected waiting snapshot: %+v", first)
	}
	second, err := s.Join("Chef")
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID || second.Token == first.Token || second.Player.Name != "Chef" || second.Opponent.Name != "Cook" {
		t.Fatalf("second player did not claim waiting match: first=%+v second=%+v", first, second)
	}
	firstView, err := s.Get(first.ID, first.Token)
	if err != nil {
		t.Fatal(err)
	}
	if firstView.Phase != "playing" || firstView.Player.Name != "Cook" || firstView.Opponent.Name != "Chef" {
		t.Fatalf("first perspective is wrong: %+v", firstView)
	}
}

func TestRoundRevealsWithoutLeakingFutureClues(t *testing.T) {
	s, now, id, token, _ := newTestStore(t)
	first, err := s.Get(id, token)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Clues) != 1 || first.Answer != nil || first.NextRevealAt == nil {
		t.Fatalf("unsafe initial snapshot: %+v", first)
	}
	*now = first.NextRevealAt.Add(time.Nanosecond)
	second, err := s.Get(id, token)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Clues) != 2 || second.Answer != nil {
		t.Fatalf("expected exactly two revealed clues, got %+v", second)
	}
}

func TestRevealTimingSlowsAfterEarlyClues(t *testing.T) {
	if got := revealWait(2, 8); got != 5*time.Second {
		t.Fatalf("early clue wait = %v, want 5s", got)
	}
	if got := revealWait(3, 8); got != 7500*time.Millisecond {
		t.Fatalf("later clue wait = %v, want 7.5s", got)
	}
	start := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if got, want := roundDeadline(start, 4), start.Add(27500*time.Millisecond); !got.Equal(want) {
		t.Fatalf("round deadline = %v, want %v", got, want)
	}
}

func TestPlayersHaveIndependentGuessLocks(t *testing.T) {
	s, now, id, firstToken, secondToken := newTestStore(t)
	m := s.matches[id]
	answer := s.currentPuzzle(m).ID
	wrong := "dish-one"
	if wrong == answer {
		wrong = "dish-two"
	}
	first, err := s.Guess(id, firstToken, wrong)
	if err != nil {
		t.Fatal(err)
	}
	if first.LockUntil == nil || first.Player.LatestGuess == "" {
		t.Fatalf("wrong guess did not lock first player: %+v", first)
	}
	second, err := s.Guess(id, secondToken, wrong)
	if err != nil {
		t.Fatalf("first player's lock blocked second: %v", err)
	}
	if second.LockUntil == nil || second.Opponent.LatestGuess == "" {
		t.Fatalf("shared guesses or second lock missing: %+v", second)
	}
	if _, err := s.Guess(id, firstToken, answer); !errors.Is(err, ErrLocked) {
		t.Fatalf("expected first player's lock error, got %v", err)
	}
	*now = first.LockUntil.Add(time.Nanosecond)
	result, err := s.Guess(id, firstToken, answer)
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.Get(id, secondToken)
	if err != nil {
		t.Fatal(err)
	}
	if result.RoundWinner != "player" || other.RoundWinner != "opponent" || other.Opponent.Score != 1 {
		t.Fatalf("winner was not perspective-oriented: first=%+v second=%+v", result, other)
	}
}

func TestAuthorizationRequired(t *testing.T) {
	s, _, id, token, _ := newTestStore(t)
	if _, err := s.Get(id, ""); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("missing token error = %v", err)
	}
	if _, err := s.Get(id, token+"wrong"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("wrong token error = %v", err)
	}
}

func TestWaitingGetActivatesBotAfterThreeSeconds(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	s := NewStore(testPuzzles())
	s.now = func() time.Time { return now }
	created, err := s.Join("Cook")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(3*time.Second - time.Nanosecond)
	stillWaiting, err := s.Get(created.ID, created.Token)
	if err != nil {
		t.Fatal(err)
	}
	if stillWaiting.Phase != "waiting" {
		t.Fatalf("fallback activated early: %+v", stillWaiting)
	}
	now = now.Add(time.Nanosecond)
	active, err := s.Get(created.ID, created.Token)
	if err != nil {
		t.Fatal(err)
	}
	if active.Phase != "playing" || !active.Opponent.IsBot || len(active.Clues) != 1 {
		t.Fatalf("bot fallback did not activate: %+v", active)
	}
}

func TestUnknownDishRejected(t *testing.T) {
	s, _, id, token, _ := newTestStore(t)
	if _, err := s.Guess(id, token, "not-in-catalog"); !errors.Is(err, ErrUnknownDish) {
		t.Fatalf("expected unknown dish error, got %v", err)
	}
}

func TestBotActsWhenItsRollFallsUnderCumulativeChance(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	s := NewStore(testPuzzles())
	s.now = func() time.Time { return now }
	created, err := s.Join("Cook")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(waitDuration)
	if _, err := s.Get(created.ID, created.Token); err != nil {
		t.Fatal(err)
	}
	m := s.matches[created.ID]
	m.botRoll = 0.01
	m.botActionAt = time.Time{}
	now = m.nextReveal
	snapshot, err := s.Get(created.ID, created.Token)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Phase != "playing" || snapshot.Opponent.LatestGuess != "" {
		t.Fatalf("bot acted before reaction time: %+v", snapshot)
	}
	actionAt := m.botActionAt
	if actionAt.IsZero() {
		t.Fatal("bot action was not scheduled from revealed clues")
	}
	now = actionAt
	result, err := s.Get(created.ID, created.Token)
	if err != nil {
		t.Fatal(err)
	}
	if result.Phase != "result" || result.Opponent.Score != 1 || result.RoundWinner != "opponent" {
		t.Fatalf("bot did not complete its clue-aware action: %+v", result)
	}
}

func TestBotCumulativeChance(t *testing.T) {
	tests := []struct {
		name     string
		revealed int
		total    int
		want     float64
	}{
		{name: "before halfway", revealed: 3, total: 8, want: 0},
		{name: "halfway", revealed: 4, total: 8, want: 0.02},
		{name: "three quarters", revealed: 6, total: 8, want: 0.09875},
		{name: "final clue", revealed: 8, total: 8, want: 0.65},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := botCumulativeChance(tt.revealed, tt.total); math.Abs(got-tt.want) > 0.000001 {
				t.Fatalf("chance = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNextRoundUsesDifferentPuzzle(t *testing.T) {
	s, now, id, token, _ := newTestStore(t)
	m := s.matches[id]
	first := s.currentPuzzle(m).ID
	result, err := s.Guess(id, token, first)
	if err != nil {
		t.Fatal(err)
	}
	*now = m.resultUntil
	next, err := s.Get(id, token)
	if err != nil {
		t.Fatal(err)
	}
	if next.Round != 2 || next.Phase != "playing" || s.currentPuzzle(m).ID == first || result.Answer == nil {
		t.Fatalf("round did not advance to a new puzzle: %+v", next)
	}
}

func TestProgressiveAssistanceUnlocksOnSchedule(t *testing.T) {
	// Four clues: family at ceil(4/2)=2, cuisine at ceil(4*0.75)=3.
	for _, tt := range []struct {
		revealed        int
		family, cuisine string
	}{
		{1, "", ""},
		{2, "soup", ""},
		{3, "soup", "Thai"},
		{4, "soup", "Thai"},
	} {
		s := NewStore(testPuzzles())
		now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		s.now = func() time.Time { return now }
		first, err := s.Join("Cook")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.Join("Rival"); err != nil {
			t.Fatal(err)
		}
		m := s.matches[first.ID]
		m.revealed = tt.revealed
		got := s.snapshot(m, 0, now)
		if len(got.Clues) != tt.revealed {
			t.Fatalf("revealed %d: got %d clues", tt.revealed, len(got.Clues))
		}
		if got.Family != tt.family {
			t.Errorf("revealed %d: family = %q, want %q", tt.revealed, got.Family, tt.family)
		}
		if got.Cuisine != tt.cuisine {
			t.Errorf("revealed %d: cuisine = %q, want %q", tt.revealed, got.Cuisine, tt.cuisine)
		}
	}
}

func TestProgressiveAssistanceFullyRevealedAfterRound(t *testing.T) {
	s, _, id, token, _ := newTestStore(t)
	m := s.matches[id]
	m.revealed = 1
	if _, err := s.Guess(id, token, s.currentPuzzle(m).ID); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(id, token)
	if err != nil {
		t.Fatal(err)
	}
	if got.Phase == "playing" {
		t.Fatal("expected the round to have ended")
	}
	// The answer is on screen, so withholding the hints would serve no purpose.
	if got.Family != "soup" || got.Cuisine != "Thai" {
		t.Fatalf("family = %q, cuisine = %q, want both populated", got.Family, got.Cuisine)
	}
}

func mixedPuzzles(ordinary, challenging int) []content.Puzzle {
	var puzzles []content.Puzzle
	add := func(n int, difficulty, prefix string) {
		for i := 0; i < n; i++ {
			puzzles = append(puzzles, content.Puzzle{
				ID: fmt.Sprintf("%s-%d", prefix, i), Name: fmt.Sprintf("%s %d", prefix, i),
				Family: "soup", Cuisine: "Thai", Difficulty: difficulty,
				Clues: []string{"a", "b", "c", "d"},
			})
		}
	}
	add(ordinary, "familiar", "easy")
	add(challenging, difficultyChallenge, "hard")
	return puzzles
}

func countChallenges(s *Store, deck []int) int {
	n := 0
	for _, i := range deck {
		if s.puzzles[i].Difficulty == difficultyChallenge {
			n++
		}
	}
	return n
}

func TestBuildDeckDealsAtMostOneChallenge(t *testing.T) {
	const ordinary, challenging = 20, 20
	s := NewStore(mixedPuzzles(ordinary, challenging))
	positions := map[int]bool{}
	for trial := 0; trial < 500; trial++ {
		deck := s.buildDeck()

		// Still a permutation: every puzzle exactly once, so a long match
		// neither repeats a dish nor runs out of cards.
		if len(deck) != ordinary+challenging {
			t.Fatalf("deck has %d cards, want %d", len(deck), ordinary+challenging)
		}
		seen := make(map[int]bool, len(deck))
		for _, i := range deck {
			if seen[i] {
				t.Fatalf("puzzle %d dealt twice", i)
			}
			seen[i] = true
		}

		// Exactly one challenge dish before the ordinary pool is exhausted.
		if got := countChallenges(s, deck[:ordinary+1]); got != 1 {
			t.Fatalf("%d challenge dishes in the playable prefix, want 1", got)
		}
		for at, i := range deck {
			if s.puzzles[i].Difficulty == difficultyChallenge {
				positions[at] = true
				break
			}
		}
	}
	// It must not always land on the same round.
	if len(positions) < 2 {
		t.Fatalf("challenge dish always dealt at %v", positions)
	}
	for at := range positions {
		if at > challengeWindow {
			t.Fatalf("challenge dealt at round %d, past the window of %d", at+1, challengeWindow)
		}
	}
}

func TestBuildDeckHandlesLopsidedCatalogs(t *testing.T) {
	for _, tt := range []struct {
		name                  string
		ordinary, challenging int
	}{
		{"no challenge dishes", 10, 0},
		{"every dish a challenge", 0, 10},
		{"single challenge dish", 9, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore(mixedPuzzles(tt.ordinary, tt.challenging))
			for trial := 0; trial < 50; trial++ {
				deck := s.buildDeck()
				if len(deck) != tt.ordinary+tt.challenging {
					t.Fatalf("deck has %d cards, want %d", len(deck), tt.ordinary+tt.challenging)
				}
				if countChallenges(s, deck) != tt.challenging {
					t.Fatal("deck lost or invented a challenge dish")
				}
				want := min(1, tt.challenging)
				if got := countChallenges(s, deck[:tt.ordinary+want]); got != want {
					t.Fatalf("prefix holds %d challenge dishes, want %d", got, want)
				}
			}
		})
	}
}
