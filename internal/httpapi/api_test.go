package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"guessthedish/internal/content"
	"guessthedish/internal/game"
)

func testServer(t *testing.T) (*httptest.Server, *http.Client) {
	t.Helper()
	puzzles := []content.Puzzle{
		{ID: "one", Name: "One", Aliases: []string{"First"}, Clues: []string{"a", "b", "c", "d"}},
		{ID: "two", Name: "Two", Clues: []string{"a", "b", "c", "d"}},
		{ID: "three", Name: "Three", Clues: []string{"a", "b", "c", "d"}},
		{ID: "four", Name: "Four", Clues: []string{"a", "b", "c", "d"}},
		{ID: "five", Name: "Five", Clues: []string{"a", "b", "c", "d"}},
	}
	server := httptest.NewServer(New(game.NewStore(puzzles), content.Catalog(content.Bundle{Puzzles: puzzles}), t.TempDir()))
	t.Cleanup(server.Close)
	return server, server.Client()
}

func createMatch(t *testing.T, client *http.Client, baseURL, name string) game.Snapshot {
	t.Helper()
	res, err := client.Post(baseURL+"/api/matches", "application/json", strings.NewReader(`{"name":`+strconvQuote(name)+`}`))
	if err != nil {
		t.Fatal(err)
	}
	var snapshot game.Snapshot
	decodeBody(t, res, &snapshot)
	if res.StatusCode != http.StatusCreated || snapshot.Token == "" {
		t.Fatalf("create failed: %d %+v", res.StatusCode, snapshot)
	}
	return snapshot
}

func postName(t *testing.T, client *http.Client, url, name string) (*http.Response, game.Snapshot) {
	t.Helper()
	res, err := client.Post(url, "application/json", strings.NewReader(`{"name":`+strconvQuote(name)+`}`))
	if err != nil {
		t.Fatal(err)
	}
	var snapshot game.Snapshot
	decodeBody(t, res, &snapshot)
	return res, snapshot
}

func authorizedRequest(t *testing.T, method, url, token string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func TestQuickPlayJoinsOldestWaitingMatchWithSeparateTokens(t *testing.T) {
	server, client := testServer(t)
	first := createMatch(t, client, server.URL, "Cook")
	if first.Phase != "waiting" || first.Round != 0 || len(first.Clues) != 0 {
		t.Fatalf("first player did not wait: %+v", first)
	}
	second := createMatch(t, client, server.URL, "Chef")
	if second.ID != first.ID || second.Token == first.Token || second.Phase != "playing" {
		t.Fatalf("second player did not join: first=%+v second=%+v", first, second)
	}
	if second.Player.Name != "Chef" || second.Opponent.Name != "Cook" {
		t.Fatalf("join response has wrong perspective: %+v", second)
	}
	res, err := client.Do(authorizedRequest(t, http.MethodGet, server.URL+"/api/matches/"+first.ID, first.Token, nil))
	if err != nil {
		t.Fatal(err)
	}
	var firstView game.Snapshot
	decodeBody(t, res, &firstView)
	if firstView.Player.Name != "Cook" || firstView.Opponent.Name != "Chef" || firstView.Token != "" {
		t.Fatalf("GET leaked token or returned wrong perspective: %+v", firstView)
	}
}

func TestMatchEndpointsRequireCorrectToken(t *testing.T) {
	server, client := testServer(t)
	first := createMatch(t, client, server.URL, "Cook")
	_ = createMatch(t, client, server.URL, "Chef")

	for _, token := range []string{"", "wrong"} {
		req, err := http.NewRequest(http.MethodGet, server.URL+"/api/matches/"+first.ID, nil)
		if err != nil {
			t.Fatal(err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("token %q got status %d", token, res.StatusCode)
		}
	}
}

func TestEitherSeatGuessesAreSharedAndPerspectiveOriented(t *testing.T) {
	server, client := testServer(t)
	first := createMatch(t, client, server.URL, "Cook")
	second := createMatch(t, client, server.URL, "Chef")
	body := []byte(`{"dishId":"one"}`)
	res, err := client.Do(authorizedRequest(t, http.MethodPost, server.URL+"/api/matches/"+first.ID+"/guesses", second.Token, bytes.NewReader(body)))
	if err != nil {
		t.Fatal(err)
	}
	var secondView game.Snapshot
	decodeBody(t, res, &secondView)
	if res.StatusCode != http.StatusOK || secondView.Player.LatestGuess != "One" {
		t.Fatalf("second seat guess failed: %d %+v", res.StatusCode, secondView)
	}
	res, err = client.Do(authorizedRequest(t, http.MethodGet, server.URL+"/api/matches/"+first.ID, first.Token, nil))
	if err != nil {
		t.Fatal(err)
	}
	var firstView game.Snapshot
	decodeBody(t, res, &firstView)
	if firstView.Opponent.LatestGuess != "One" {
		t.Fatalf("opponent guess was not shared: %+v", firstView)
	}
}

func TestDuplicateNamesCanMatchEachOther(t *testing.T) {
	server, client := testServer(t)
	first := createMatch(t, client, server.URL, "Cook")
	second := createMatch(t, client, server.URL, "Cook")
	if first.ID != second.ID || first.Token == second.Token || second.Player.Name != "Cook" || second.Opponent.Name != "Cook" {
		t.Fatalf("duplicate names were not safely matched: first=%+v second=%+v", first, second)
	}
}

func TestAuthenticatedGetActivatesBotFallback(t *testing.T) {
	server, client := testServer(t)
	created := createMatch(t, client, server.URL, "Cook")
	time.Sleep(3 * time.Second)
	res, err := client.Do(authorizedRequest(t, http.MethodGet, server.URL+"/api/matches/"+created.ID, created.Token, nil))
	if err != nil {
		t.Fatal(err)
	}
	var active game.Snapshot
	decodeBody(t, res, &active)
	if res.StatusCode != http.StatusOK || active.Phase != "playing" || !active.Opponent.IsBot || len(active.Clues) != 1 {
		t.Fatalf("bot fallback did not activate: %d %+v", res.StatusCode, active)
	}
}

func TestAuthenticatedDeleteRemovesWholeMatch(t *testing.T) {
	server, client := testServer(t)
	first := createMatch(t, client, server.URL, "Cook")
	second := createMatch(t, client, server.URL, "Chef")
	res, err := client.Do(authorizedRequest(t, http.MethodDelete, server.URL+"/api/matches/"+first.ID, first.Token, nil))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete failed: %d", res.StatusCode)
	}
	res, err = client.Do(authorizedRequest(t, http.MethodGet, server.URL+"/api/matches/"+first.ID, second.Token, nil))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("deleted match remained for second seat: %d", res.StatusCode)
	}
}

func TestPrivateRoomCreateWaitJoinAndQuickPlayIsolation(t *testing.T) {
	server, client := testServer(t)
	res, host := postName(t, client, server.URL+"/api/rooms", "Host")
	if res.StatusCode != http.StatusCreated || host.Phase != "waiting" || host.Token == "" || host.RoomCode == "" {
		t.Fatalf("room create failed: %d %+v", res.StatusCode, host)
	}
	if len(host.RoomCode) != 16 {
		t.Fatalf("room code does not encode 96 bits: %q", host.RoomCode)
	}
	time.Sleep(3 * time.Second)
	waitRes, err := client.Do(authorizedRequest(t, http.MethodGet, server.URL+"/api/matches/"+host.ID, host.Token, nil))
	if err != nil {
		t.Fatal(err)
	}
	var waiting game.Snapshot
	decodeBody(t, waitRes, &waiting)
	if waitRes.StatusCode != http.StatusOK || waiting.Phase != "waiting" || waiting.Opponent.IsBot || waiting.Token != "" {
		t.Fatalf("private room did not remain waiting: %d %+v", waitRes.StatusCode, waiting)
	}

	quick := createMatch(t, client, server.URL, "Quick One")
	if quick.ID == host.ID {
		t.Fatalf("Quick Play claimed private room: room=%+v quick=%+v", host, quick)
	}

	res, guest := postName(t, client, server.URL+"/api/rooms/"+host.RoomCode+"/join", "Guest")
	if res.StatusCode != http.StatusOK || guest.ID != host.ID || guest.RoomCode != host.RoomCode || guest.Phase != "playing" || guest.Token == "" || guest.Token == host.Token {
		t.Fatalf("room join failed: %d host=%+v guest=%+v", res.StatusCode, host, guest)
	}
	if guest.Player.Name != "Guest" || guest.Opponent.Name != "Host" || guest.Opponent.IsBot {
		t.Fatalf("room join perspective is wrong: %+v", guest)
	}

	quickTwo := createMatch(t, client, server.URL, "Quick Two")
	if quickTwo.ID != quick.ID || quickTwo.ID == host.ID {
		t.Fatalf("Quick Play queue was not isolated: first=%+v second=%+v", quick, quickTwo)
	}

	guessBody := bytes.NewBufferString(`{"dishId":"one"}`)
	guessRes, err := client.Do(authorizedRequest(t, http.MethodPost, server.URL+"/api/matches/"+host.ID+"/guesses", guest.Token, guessBody))
	if err != nil {
		t.Fatal(err)
	}
	var guestView game.Snapshot
	decodeBody(t, guessRes, &guestView)
	if guessRes.StatusCode != http.StatusOK || guestView.Player.LatestGuess != "One" {
		t.Fatalf("guest could not use match API: %d %+v", guessRes.StatusCode, guestView)
	}
	hostRes, err := client.Do(authorizedRequest(t, http.MethodGet, server.URL+"/api/matches/"+host.ID, host.Token, nil))
	if err != nil {
		t.Fatal(err)
	}
	var hostView game.Snapshot
	decodeBody(t, hostRes, &hostView)
	if hostRes.StatusCode != http.StatusOK || hostView.Opponent.LatestGuess != "One" || hostView.Token != "" {
		t.Fatalf("host did not see shared match state or GET leaked token: %d %+v", hostRes.StatusCode, hostView)
	}
	deleteRes, err := client.Do(authorizedRequest(t, http.MethodDelete, server.URL+"/api/matches/"+host.ID, guest.Token, nil))
	if err != nil {
		t.Fatal(err)
	}
	_ = deleteRes.Body.Close()
	if deleteRes.StatusCode != http.StatusNoContent {
		t.Fatalf("guest delete failed: %d", deleteRes.StatusCode)
	}
	hostRes, err = client.Do(authorizedRequest(t, http.MethodGet, server.URL+"/api/matches/"+host.ID, host.Token, nil))
	if err != nil {
		t.Fatal(err)
	}
	_ = hostRes.Body.Close()
	if hostRes.StatusCode != http.StatusNotFound {
		t.Fatalf("guest delete did not remove host match: %d", hostRes.StatusCode)
	}
}

func TestPrivateRoomUnknownAndFull(t *testing.T) {
	server, client := testServer(t)
	res, err := client.Post(server.URL+"/api/rooms/not-a-room/join", "application/json", strings.NewReader(`{"name":"Guest"}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown room status = %d, want 404", res.StatusCode)
	}

	_, host := postName(t, client, server.URL+"/api/rooms", "Host")
	res, _ = postName(t, client, server.URL+"/api/rooms/"+host.RoomCode+"/join", "Guest")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("first join status = %d", res.StatusCode)
	}
	res, err = client.Post(server.URL+"/api/rooms/"+host.RoomCode+"/join", "application/json", strings.NewReader(`{"name":"Third"}`))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("full room status = %d, want 409", res.StatusCode)
	}
}

func TestAPIRejectsUnknownJSONFields(t *testing.T) {
	handler := New(game.NewStore(nil), nil, t.TempDir())
	req := httptest.NewRequest(http.MethodPost, "/api/matches", strings.NewReader(`{"name":"Cook","admin":true}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest || !strings.Contains(res.Body.String(), "unknown field") {
		t.Fatalf("expected strict JSON error: %d %s", res.Code, res.Body.String())
	}
}

func TestMetricsExposeBoundedRoutesAndGameState(t *testing.T) {
	server, client := testServer(t)
	created := createMatch(t, client, server.URL, "Cook")
	res, err := client.Do(authorizedRequest(t, http.MethodGet, server.URL+"/api/matches/"+created.ID, created.Token, nil))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()

	res, err = client.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	metrics := string(body)
	for _, expected := range []string{
		`route="/api/matches/{id}"`,
		`guessthedish_matches{phase="waiting"} 1`,
		`guessthedish_quick_play_queue 1`,
		`guessthedish_process_uptime_seconds`,
	} {
		if !strings.Contains(metrics, expected) {
			t.Fatalf("metrics missing %q:\n%s", expected, metrics)
		}
	}
	if strings.Contains(metrics, created.ID) {
		t.Fatalf("metrics leaked match ID %q", created.ID)
	}
}

func TestMetricsNormalizeUnknownMethods(t *testing.T) {
	handler := New(game.NewStore(nil), nil, t.TempDir())
	for _, method := range []string{"CUSTOM-ONE", "CUSTOM-TWO"} {
		req := httptest.NewRequest(method, "/api/catalog", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := res.Body.String()
	if strings.Count(body, `method="OTHER",route="/api/catalog"`) != 3 {
		t.Fatalf("unknown methods were not normalized:\n%s", body)
	}
	if strings.Contains(body, "CUSTOM-") {
		t.Fatalf("metrics retained an attacker-controlled method:\n%s", body)
	}
}

func strconvQuote(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func decodeBody(t *testing.T, response *http.Response, target any) {
	t.Helper()
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatal(err)
	}
}
