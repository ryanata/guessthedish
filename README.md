# Guess the Dish

The Go backend is a standard-library-only, server-authoritative local match implementation. It loads the private puzzle JSON at runtime; the public catalog endpoint never includes clues.

## Run locally

```sh
go run ./cmd/server
```

The default address is `127.0.0.1:8080`, and local development content defaults to `../guessthedish-data/data/puzzles.json`. Override these with `ADDR`, `CONTENT_PATH`, and `DIST_PATH`. If `dist/index.html` exists, the server serves static assets with an SPA fallback; API and health endpoints work without it.

## HTTP API

- `GET /healthz` and `GET /readyz`
- `GET /api/catalog` returns `{ "dishes": [{ "id", "name", "aliases" }] }`.
- `POST /api/matches` with `{ "name": "Guest" }` joins the oldest waiting human or creates a waiting match. The response includes that seat's opaque `token`.
- `POST /api/rooms` with `{ "name": "Host" }` creates a private waiting room and returns its snapshot with a 96-bit URL-safe `roomCode` and the host seat's opaque `token`.
- `POST /api/rooms/{roomCode}/join` with `{ "name": "Guest" }` claims the second seat and starts round one. Unknown codes return `404`; occupied rooms return `409`.
- `GET /api/matches/{id}` advances lazy timers and returns the current snapshot. If the first player has waited at least three seconds, this authenticated request fills the opponent seat with a bot and starts round one.
- `POST /api/matches/{id}/guesses` with `{ "dishId": "pad-thai" }` submits a catalog selection.
- `DELETE /api/matches/{id}` cancels a waiting match or deletes an active match for both seats.

Private rooms remain waiting indefinitely while retained, never receive a bot, and are excluded from Quick Play matchmaking. After joining, both room seats use the normal match-ID endpoints above. All match-specific requests require `Authorization: Bearer <token>`; missing or invalid tokens return `401`. Tokens are returned only by the successful creation/join response. Snapshots are oriented to the token's seat and use `player` and `opponent`; `opponent.isBot` identifies fallback play. `phase` is `waiting`, `playing`, `result`, or `finished`. Winners are `player`, `opponent`, or `none` from each seat's perspective. Other fields include scores/names/latest visible guesses, stable non-conflicting `avatar` descriptors, revealed `clues`, `totalClueCount`, and RFC3339Nano absolute `nextRevealAt`, `deadlineAt`, and seat-specific `lockUntil` values where applicable. `answer`, all clues, and `roundWinner` appear only after a round ends; `matchWinner` appears when a player reaches three. API failures use `{ "error": { "message": "..." } }`.

Run tests with `go test ./...`.
