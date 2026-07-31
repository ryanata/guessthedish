package game

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"math"
	mathrand "math/rand"
	"sync"
	"time"

	"guessthedish/internal/content"
)

var (
	ErrNotFound     = errors.New("match not found")
	ErrUnauthorized = errors.New("invalid or missing match token")
	ErrNotPlaying   = errors.New("match is not accepting guesses")
	ErrLocked       = errors.New("player is temporarily locked")
	ErrUnknownDish  = errors.New("unknown dish ID")
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomFull     = errors.New("room is full")
)

const (
	wrongLock      = 300 * time.Millisecond
	resultDuration = 3 * time.Second
	waitDuration   = 3 * time.Second
	noWinner       = -1

	// difficultyChallenge is the catalog's hardest tier, capped at one per match.
	difficultyChallenge = "challenge"
	// challengeWindow is how many opening rounds the challenge dish may land in,
	// chosen to cover a long first-to-3 match.
	challengeWindow = 6
)

type Store struct {
	mu      sync.Mutex
	puzzles []content.Puzzle
	dishes  map[string]content.Dish
	matches map[string]*match
	rooms   map[string]string
	now     func() time.Time
	random  *mathrand.Rand
}

type match struct {
	id          string
	roomCode    string
	players     [2]player
	tokenHashes [2][sha256.Size]byte
	occupied    [2]bool
	bot         bool
	phase       string
	waitingAt   time.Time
	round       int
	puzzleOrder []int
	puzzlePos   int
	revealed    int
	nextReveal  time.Time
	deadline    time.Time
	resultUntil time.Time
	winner      int
	botRoll     float64
	botActionAt time.Time
	lastAccess  time.Time
}

type player struct {
	name       string
	avatar     Avatar
	score      int
	guess      string
	guessUntil time.Time
	lockUntil  time.Time
}

type Snapshot struct {
	ID             string     `json:"id"`
	Token          string     `json:"token,omitempty"`
	RoomCode       string     `json:"roomCode,omitempty"`
	Phase          string     `json:"phase"`
	Round          int        `json:"round"`
	TotalClueCount int        `json:"totalClueCount"`
	Clues          []string   `json:"clues"`
	Family         string     `json:"family,omitempty"`
	Cuisine        string     `json:"cuisine,omitempty"`
	Player         PlayerView `json:"player"`
	Opponent       PlayerView `json:"opponent"`
	NextRevealAt   *time.Time `json:"nextRevealAt,omitempty"`
	DeadlineAt     *time.Time `json:"deadlineAt,omitempty"`
	LockUntil      *time.Time `json:"lockUntil,omitempty"`
	Answer         *Answer    `json:"answer,omitempty"`
	RoundWinner    string     `json:"roundWinner,omitempty"`
	MatchWinner    string     `json:"matchWinner,omitempty"`
}

type PlayerView struct {
	Name        string `json:"name"`
	Avatar      Avatar `json:"avatar"`
	Score       int    `json:"score"`
	LatestGuess string `json:"latestGuess,omitempty"`
	IsBot       bool   `json:"isBot,omitempty"`
}

type Avatar struct {
	Color string `json:"color"`
	Style int    `json:"style"`
}

type Answer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var botNames = []string{"PrepCook", "SauceBoss", "TableSeven", "MiseEnPlace", "PanSlinger", "DailySpecial"}
var avatarColors = []string{"paprika", "herb", "mustard", "aubergine", "blue", "rose"}

func NewStore(puzzles []content.Puzzle) *Store {
	dishes := make(map[string]content.Dish, len(puzzles))
	for _, puzzle := range puzzles {
		dishes[puzzle.ID] = content.Dish{ID: puzzle.ID, Name: puzzle.Name, Aliases: puzzle.Aliases}
	}
	return &Store{
		puzzles: puzzles,
		dishes:  dishes,
		matches: make(map[string]*match),
		rooms:   make(map[string]string),
		now:     time.Now,
		random:  mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
	}
}

// Join places a player in the oldest non-expired waiting match, or creates one.
func (s *Store) Join(name string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	s.cleanup(now)

	var waiting *match
	for _, candidate := range s.matches {
		if candidate.roomCode == "" && candidate.phase == "waiting" && now.Sub(candidate.waitingAt) < waitDuration &&
			(waiting == nil || candidate.waitingAt.Before(waiting.waitingAt)) {
			waiting = candidate
		}
	}

	token, hash, err := randomToken()
	if err != nil {
		return Snapshot{}, err
	}
	if waiting != nil {
		waiting.players[1].name = name
		waiting.tokenHashes[1] = hash
		waiting.occupied[1] = true
		waiting.lastAccess = now
		s.startRound(waiting, now)
		view := s.snapshot(waiting, 1, now)
		view.Token = token
		return view, nil
	}

	id, err := randomOpaqueValue(18)
	if err != nil {
		return Snapshot{}, err
	}
	firstAvatar, secondAvatar := s.avatarPair()
	m := &match{
		id:          id,
		players:     [2]player{{name: name, avatar: firstAvatar}, {avatar: secondAvatar}},
		tokenHashes: [2][sha256.Size]byte{hash, {}},
		occupied:    [2]bool{true, false},
		phase:       "waiting",
		waitingAt:   now,
		winner:      noWinner,
		puzzleOrder: s.buildDeck(),
		lastAccess:  now,
	}
	s.matches[id] = m
	view := s.snapshot(m, 0, now)
	view.Token = token
	return view, nil
}

// CreateRoom creates a private waiting match that is excluded from Quick Play.
func (s *Store) CreateRoom(name string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	s.cleanup(now)

	token, hash, err := randomToken()
	if err != nil {
		return Snapshot{}, err
	}
	id, err := randomOpaqueValue(18)
	if err != nil {
		return Snapshot{}, err
	}
	var code string
	for {
		code, err = randomOpaqueValue(12)
		if err != nil {
			return Snapshot{}, err
		}
		if _, exists := s.rooms[code]; !exists {
			break
		}
	}
	firstAvatar, secondAvatar := s.avatarPair()
	m := &match{
		id:          id,
		roomCode:    code,
		players:     [2]player{{name: name, avatar: firstAvatar}, {avatar: secondAvatar}},
		tokenHashes: [2][sha256.Size]byte{hash, {}},
		occupied:    [2]bool{true, false},
		phase:       "waiting",
		waitingAt:   now,
		winner:      noWinner,
		puzzleOrder: s.buildDeck(),
		lastAccess:  now,
	}
	s.matches[id] = m
	s.rooms[code] = id
	view := s.snapshot(m, 0, now)
	view.Token = token
	return view, nil
}

// JoinRoom claims the second seat in a private room and starts its first round.
func (s *Store) JoinRoom(code, name string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	s.cleanup(now)
	id, ok := s.rooms[code]
	if !ok {
		return Snapshot{}, ErrRoomNotFound
	}
	m, ok := s.matches[id]
	if !ok {
		delete(s.rooms, code)
		return Snapshot{}, ErrRoomNotFound
	}
	if m.occupied[1] {
		return Snapshot{}, ErrRoomFull
	}
	token, hash, err := randomToken()
	if err != nil {
		return Snapshot{}, err
	}
	m.players[1].name = name
	m.tokenHashes[1] = hash
	m.occupied[1] = true
	m.lastAccess = now
	s.startRound(m, now)
	view := s.snapshot(m, 1, now)
	view.Token = token
	return view, nil
}

func (s *Store) Get(id, token string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	m, seat, err := s.authorize(id, token)
	if err != nil {
		return Snapshot{}, err
	}
	if m.roomCode == "" && m.phase == "waiting" && !now.Before(m.waitingAt.Add(waitDuration)) {
		m.players[1].name = botNames[s.random.Intn(len(botNames))]
		m.occupied[1] = true
		m.bot = true
		s.startRound(m, now)
	}
	s.advance(m, now)
	m.lastAccess = now
	return s.snapshot(m, seat, now), nil
}

func (s *Store) Guess(id, token, dishID string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	m, seat, err := s.authorize(id, token)
	if err != nil {
		return Snapshot{}, err
	}
	s.advance(m, now)
	if _, ok := s.dishes[dishID]; !ok {
		return Snapshot{}, ErrUnknownDish
	}
	if m.phase != "playing" {
		return Snapshot{}, ErrNotPlaying
	}
	p := &m.players[seat]
	if now.Before(p.lockUntil) {
		return Snapshot{}, ErrLocked
	}
	p.guess = s.dishes[dishID].Name
	p.guessUntil = now.Add(2 * time.Second)
	if dishID == s.currentPuzzle(m).ID {
		s.endRound(m, now, seat)
	} else {
		p.lockUntil = now.Add(wrongLock)
	}
	m.lastAccess = now
	return s.snapshot(m, seat, now), nil
}

// Delete removes the entire match, including an active match, for both seats.
func (s *Store) Delete(id, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, _, err := s.authorize(id, token)
	if err != nil {
		return err
	}
	s.deleteMatch(id, m)
	return nil
}

func (s *Store) authorize(id, token string) (*match, int, error) {
	m, ok := s.matches[id]
	if !ok {
		return nil, 0, ErrNotFound
	}
	hash := sha256.Sum256([]byte(token))
	for seat := range m.players {
		if m.occupied[seat] && subtle.ConstantTimeCompare(hash[:], m.tokenHashes[seat][:]) == 1 {
			return m, seat, nil
		}
	}
	return nil, 0, ErrUnauthorized
}

func (s *Store) startRound(m *match, at time.Time) {
	m.phase = "playing"
	m.round++
	m.revealed = 1
	m.winner = noWinner
	for i := range m.players {
		m.players[i].lockUntil = time.Time{}
		m.players[i].guess = ""
	}
	puzzle := s.currentPuzzle(m)
	m.nextReveal = at.Add(revealWait(1, len(puzzle.Clues)))
	m.deadline = roundDeadline(at, len(puzzle.Clues))
	m.botRoll = s.random.Float64()
	m.botActionAt = time.Time{}
	s.scheduleBot(m, at)
}

func (s *Store) advance(m *match, now time.Time) {
	for {
		if m.phase == "result" {
			if now.Before(m.resultUntil) {
				return
			}
			if m.players[0].score == 3 || m.players[1].score == 3 || m.puzzlePos+1 >= len(m.puzzleOrder) {
				m.phase = "finished"
				return
			}
			m.puzzlePos++
			s.startRound(m, m.resultUntil)
			continue
		}
		if m.phase != "playing" {
			return
		}

		eventAt := m.deadline
		event := "deadline"
		if !m.nextReveal.IsZero() && m.nextReveal.Before(eventAt) {
			eventAt, event = m.nextReveal, "reveal"
		}
		if !m.botActionAt.IsZero() && m.botActionAt.Before(eventAt) {
			eventAt, event = m.botActionAt, "bot"
		}
		if now.Before(eventAt) {
			return
		}
		switch event {
		case "reveal":
			m.revealed++
			puzzle := s.currentPuzzle(m)
			if m.revealed < len(puzzle.Clues) {
				m.nextReveal = eventAt.Add(revealWait(m.revealed, len(puzzle.Clues)))
			} else {
				m.nextReveal = time.Time{}
			}
			s.scheduleBot(m, eventAt)
		case "bot":
			m.botActionAt = time.Time{}
			m.players[1].guess = s.currentPuzzle(m).Name
			m.players[1].guessUntil = eventAt.Add(2 * time.Second)
			s.endRound(m, eventAt, 1)
		case "deadline":
			s.endRound(m, eventAt, noWinner)
		}
	}
}

func (s *Store) scheduleBot(m *match, at time.Time) {
	if !m.bot || !m.botActionAt.IsZero() {
		return
	}
	if m.botRoll < botCumulativeChance(m.revealed, len(s.currentPuzzle(m).Clues)) {
		m.botActionAt = at.Add(time.Duration(450+s.random.Intn(751)) * time.Millisecond)
	}
}

func (s *Store) endRound(m *match, at time.Time, winner int) {
	m.phase = "result"
	m.winner = winner
	m.resultUntil = at.Add(resultDuration)
	m.nextReveal = time.Time{}
	m.botActionAt = time.Time{}
	if winner != noWinner {
		m.players[winner].score++
	}
}

func (s *Store) snapshot(m *match, seat int, now time.Time) Snapshot {
	opponent := 1 - seat
	view := Snapshot{
		ID:       m.id,
		RoomCode: m.roomCode,
		Phase:    m.phase,
		Round:    m.round,
		Clues:    []string{},
		Player:   playerSnapshot(m.players[seat], now, m.phase != "playing" && m.winner == seat),
		Opponent: playerSnapshot(m.players[opponent], now, m.phase != "playing" && m.winner == opponent),
	}
	view.Opponent.IsBot = m.bot && opponent == 1
	if m.phase == "waiting" {
		return view
	}

	puzzle := s.currentPuzzle(m)
	clueCount := m.revealed
	if m.phase != "playing" {
		clueCount = len(puzzle.Clues)
	}
	view.TotalClueCount = len(puzzle.Clues)
	view.Clues = append([]string(nil), puzzle.Clues[:clueCount]...)
	// Progressive assistance. Gated here rather than on the client so a hint
	// the player has not earned never leaves the server. clueCount is the full
	// list once the round ends, so both are visible with the answer.
	if clueCount >= familyRevealAt(len(puzzle.Clues)) {
		view.Family = puzzle.Family
	}
	if clueCount >= cuisineRevealAt(len(puzzle.Clues)) {
		view.Cuisine = puzzle.Cuisine
	}
	if m.phase == "playing" {
		if !m.nextReveal.IsZero() {
			t := m.nextReveal
			view.NextRevealAt = &t
		}
		deadline := m.deadline
		view.DeadlineAt = &deadline
		if now.Before(m.players[seat].lockUntil) {
			lockUntil := m.players[seat].lockUntil
			view.LockUntil = &lockUntil
		}
	} else {
		view.Answer = &Answer{ID: puzzle.ID, Name: puzzle.Name}
		view.RoundWinner = perspectiveWinner(m.winner, seat)
		if m.phase == "finished" {
			if m.players[seat].score == 3 {
				view.MatchWinner = "player"
			} else if m.players[opponent].score == 3 {
				view.MatchWinner = "opponent"
			} else {
				view.MatchWinner = "none"
			}
		}
	}
	return view
}

func perspectiveWinner(winner, seat int) string {
	if winner == noWinner {
		return "none"
	}
	if winner == seat {
		return "player"
	}
	return "opponent"
}

func playerSnapshot(p player, now time.Time, preserveGuess bool) PlayerView {
	view := PlayerView{Name: p.name, Avatar: p.avatar, Score: p.score}
	if preserveGuess || now.Before(p.guessUntil) {
		view.LatestGuess = p.guess
	}
	return view
}

func (s *Store) avatarPair() (Avatar, Avatar) {
	const styles = 4
	total := len(avatarColors) * styles
	first := s.random.Intn(total)
	second := s.random.Intn(total - 1)
	if second >= first {
		second++
	}
	avatar := func(index int) Avatar {
		return Avatar{Color: avatarColors[index/styles], Style: index % styles}
	}
	return avatar(first), avatar(second)
}

func (s *Store) currentPuzzle(m *match) content.Puzzle {
	return s.puzzles[m.puzzleOrder[m.puzzlePos]]
}

// buildDeck orders one match's puzzles so a player meets at most one challenge
// dish. Roughly a quarter of the catalog is rated challenge, so an unweighted
// shuffle deals two or more hard dishes to about half of all five-round
// matches, which is what makes the game feel arbitrary rather than difficult.
//
// One challenge dish is dealt into the opening rounds and the rest are moved
// behind every other puzzle. The result is still a permutation of the whole
// catalog, so a long match never runs out of cards and never repeats one.
func (s *Store) buildDeck() []int {
	var ordinary, challenging []int
	for _, i := range s.random.Perm(len(s.puzzles)) {
		if s.puzzles[i].Difficulty == difficultyChallenge {
			challenging = append(challenging, i)
		} else {
			ordinary = append(ordinary, i)
		}
	}
	if len(challenging) == 0 {
		return ordinary
	}
	// Land it anywhere in a typical match rather than always on the same round.
	at := s.random.Intn(min(challengeWindow, len(ordinary)) + 1)
	deck := make([]int, 0, len(s.puzzles))
	deck = append(deck, ordinary[:at]...)
	deck = append(deck, challenging[0])
	deck = append(deck, ordinary[at:]...)
	return append(deck, challenging[1:]...)
}

// familyRevealAt and cuisineRevealAt are the clue counts at which progressive
// assistance unlocks: the dish family at half the clues, the cuisine at three
// quarters. Assistance is deliberately limited to these two; first-letter and
// word-length hints turn recall into a spelling exercise.
func familyRevealAt(total int) int {
	return int(math.Ceil(float64(total) / 2))
}

func cuisineRevealAt(total int) int {
	return int(math.Ceil(float64(total) * 0.75))
}

func botCumulativeChance(revealed, total int) float64 {
	halfway := int(math.Ceil(float64(total) / 2))
	if revealed < halfway {
		return 0
	}
	progress := float64(revealed-halfway) / float64(total-halfway)
	return 0.02 + 0.63*math.Pow(progress, 3)
}

func revealWait(revealed, total int) time.Duration {
	k := int(math.Ceil(float64(total) * 0.25))
	if revealed <= k {
		return 5 * time.Second
	}
	return 7500 * time.Millisecond
}

func roundDeadline(start time.Time, total int) time.Time {
	at := start
	for revealed := 1; revealed <= total; revealed++ {
		at = at.Add(revealWait(revealed, total))
	}
	return at
}

func randomToken() (string, [sha256.Size]byte, error) {
	token, err := randomOpaqueValue(32)
	if err != nil {
		return "", [sha256.Size]byte{}, err
	}
	return token, sha256.Sum256([]byte(token)), nil
}

func randomOpaqueValue(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func (s *Store) cleanup(now time.Time) {
	for id, m := range s.matches {
		maxAge := time.Hour
		if m.phase == "finished" {
			maxAge = 15 * time.Minute
		}
		if now.Sub(m.lastAccess) > maxAge {
			s.deleteMatch(id, m)
		}
	}
}

func (s *Store) deleteMatch(id string, m *match) {
	delete(s.matches, id)
	if m.roomCode != "" {
		delete(s.rooms, m.roomCode)
	}
}
