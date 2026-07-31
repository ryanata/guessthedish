package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"guessthedish/internal/content"
	"guessthedish/internal/game"
)

const maxBodyBytes = 4096

type API struct {
	store   *game.Store
	catalog []content.Dish
}

func New(store *game.Store, catalog []content.Dish, distPath string) http.Handler {
	api := &API{store: store, catalog: catalog}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", api.health)
	mux.HandleFunc("/readyz", api.health)
	mux.HandleFunc("/api/catalog", api.catalogHandler)
	mux.HandleFunc("/api/rooms", api.rooms)
	mux.HandleFunc("/api/rooms/", api.room)
	mux.HandleFunc("/api/matches", api.matches)
	mux.HandleFunc("/api/matches/", api.match)
	mux.Handle("/", staticHandler(distPath))
	return mux
}

func (a *API) rooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	name, ok := decodeName(w, r)
	if !ok {
		return
	}
	snapshot, err := a.store.CreateRoom(name)
	a.writeGameResult(w, snapshot, err, http.StatusCreated)
}

func (a *API) room(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/rooms/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "join" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	name, ok := decodeName(w, r)
	if !ok {
		return
	}
	snapshot, err := a.store.JoinRoom(parts[0], name)
	a.writeGameResult(w, snapshot, err, http.StatusOK)
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) catalogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"dishes": a.catalog})
}

func (a *API) matches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	name, ok := decodeName(w, r)
	if !ok {
		return
	}
	snapshot, err := a.store.Join(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create match")
		return
	}
	writeJSON(w, http.StatusCreated, snapshot)
}

func decodeName(w http.ResponseWriter, r *http.Request) (string, bool) {
	var input struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return "", false
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len([]rune(input.Name)) > 40 {
		writeError(w, http.StatusBadRequest, "name must contain 1-40 characters")
		return "", false
	}
	return input.Name, true
}

func (a *API) match(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/matches/")
	parts := strings.Split(path, "/")
	if parts[0] == "" || len(parts) > 2 || (len(parts) == 2 && parts[1] != "guesses") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	id := parts[0]
	token, err := bearerToken(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if len(parts) == 2 {
		a.guess(w, r, id, token)
		return
	}
	switch r.Method {
	case http.MethodGet:
		snapshot, err := a.store.Get(id, token)
		a.writeGameResult(w, snapshot, err, http.StatusOK)
	case http.MethodDelete:
		if err := a.store.Delete(id, token); err != nil {
			a.writeGameError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(w, http.MethodGet+", "+http.MethodDelete)
	}
}

func (a *API) guess(w http.ResponseWriter, r *http.Request, id, token string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	var input struct {
		DishID string `json:"dishId"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.DishID == "" {
		writeError(w, http.StatusBadRequest, "dishId is required")
		return
	}
	snapshot, err := a.store.Guess(id, token, input.DishID)
	a.writeGameResult(w, snapshot, err, http.StatusOK)
}

func (a *API) writeGameResult(w http.ResponseWriter, snapshot game.Snapshot, err error, status int) {
	if err != nil {
		a.writeGameError(w, err)
		return
	}
	writeJSON(w, status, snapshot)
}

func (a *API) writeGameError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, game.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, game.ErrRoomNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, game.ErrRoomFull):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, game.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, game.ErrUnknownDish), errors.Is(err, game.ErrNotPlaying):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, game.ErrLocked):
		writeError(w, http.StatusTooManyRequests, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func bearerToken(r *http.Request) (string, error) {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", game.ErrUnauthorized
	}
	return parts[1], nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	if contentType := r.Header.Get("Content-Type"); contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		return errors.New("Content-Type must be application/json")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return errors.New("invalid JSON body: " + err.Error())
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("body must contain one JSON object")
	}
	return nil
}

func staticHandler(distPath string) http.Handler {
	info, err := os.Stat(filepath.Join(distPath, "index.html"))
	if err != nil || info.IsDir() {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeError(w, http.StatusNotFound, "not found")
		})
	}
	files := http.Dir(distPath)
	server := http.FileServer(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			methodNotAllowed(w, http.MethodGet+", "+http.MethodHead)
			return
		}
		name := strings.TrimPrefix(filepath.Clean(r.URL.Path), string(filepath.Separator))
		if name != "." {
			if _, err := os.Stat(filepath.Join(distPath, name)); err == nil {
				server.ServeHTTP(w, r)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
	})
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"message": message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
