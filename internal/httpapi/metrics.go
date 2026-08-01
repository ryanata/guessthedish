package httpapi

import (
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type requestMetric struct {
	method string
	route  string
	status int
}

type metricValue struct {
	count    uint64
	duration time.Duration
}

type metrics struct {
	mu       sync.Mutex
	started  time.Time
	requests map[requestMetric]metricValue
}

func newMetrics() *metrics {
	return &metrics{started: time.Now(), requests: make(map[requestMetric]metricValue)}
}

func (m *metrics) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		started := time.Now()
		writer := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(writer, r)
		key := requestMetric{method: metricMethod(r.Method), route: metricRoute(r.URL.Path), status: writer.status}
		m.mu.Lock()
		value := m.requests[key]
		value.count++
		value.duration += time.Since(started)
		m.requests[key] = value
		m.mu.Unlock()
	})
}

func (m *metrics) serveHTTP(api *API, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}

	m.mu.Lock()
	requests := make(map[requestMetric]metricValue, len(m.requests))
	for key, value := range m.requests {
		requests[key] = value
	}
	m.mu.Unlock()

	keys := make([]requestMetric, 0, len(requests))
	for key := range requests {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := keys[i]
		right := keys[j]
		if left.route != right.route {
			return left.route < right.route
		}
		if left.method != right.method {
			return left.method < right.method
		}
		return left.status < right.status
	})

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintln(w, "# HELP guessthedish_http_requests_total HTTP requests handled by route, method, and status.")
	fmt.Fprintln(w, "# TYPE guessthedish_http_requests_total counter")
	for _, key := range keys {
		value := requests[key]
		fmt.Fprintf(w, "guessthedish_http_requests_total{method=%s,route=%s,status=%s} %d\n", quote(key.method), quote(key.route), quote(strconv.Itoa(key.status)), value.count)
	}
	fmt.Fprintln(w, "# HELP guessthedish_http_request_duration_seconds HTTP request duration by route and method.")
	fmt.Fprintln(w, "# TYPE guessthedish_http_request_duration_seconds summary")
	for _, key := range keys {
		value := requests[key]
		labels := fmt.Sprintf("method=%s,route=%s,status=%s", quote(key.method), quote(key.route), quote(strconv.Itoa(key.status)))
		fmt.Fprintf(w, "guessthedish_http_request_duration_seconds_sum{%s} %g\n", labels, value.duration.Seconds())
		fmt.Fprintf(w, "guessthedish_http_request_duration_seconds_count{%s} %d\n", labels, value.count)
	}

	stats := api.store.Stats()
	fmt.Fprintln(w, "# HELP guessthedish_matches Current matches by phase.")
	fmt.Fprintln(w, "# TYPE guessthedish_matches gauge")
	for phase, value := range map[string]int{"waiting": stats.Waiting, "playing": stats.Playing, "result": stats.Result, "finished": stats.Finished} {
		fmt.Fprintf(w, "guessthedish_matches{phase=%s} %d\n", quote(phase), value)
	}
	writeGauge(w, "guessthedish_matches_total", "Current retained matches.", stats.Matches)
	writeGauge(w, "guessthedish_rooms", "Current private rooms.", stats.Rooms)
	writeGauge(w, "guessthedish_bot_matches", "Current matches containing a bot.", stats.BotMatches)
	writeGauge(w, "guessthedish_quick_play_queue", "Players waiting in Quick Play.", stats.QuickPlayQueue)
	writeGauge(w, "guessthedish_process_goroutines", "Current Go goroutines.", runtime.NumGoroutine())
	fmt.Fprintln(w, "# HELP guessthedish_process_uptime_seconds Process uptime in seconds.")
	fmt.Fprintln(w, "# TYPE guessthedish_process_uptime_seconds gauge")
	fmt.Fprintf(w, "guessthedish_process_uptime_seconds %g\n", time.Since(m.started).Seconds())
}

func metricRoute(path string) string {
	switch path {
	case "/healthz", "/readyz", "/api/catalog", "/api/rooms", "/api/matches":
		return path
	}
	if strings.HasPrefix(path, "/api/rooms/") {
		return "/api/rooms/{code}/join"
	}
	if strings.HasPrefix(path, "/api/matches/") {
		if strings.HasSuffix(path, "/guesses") {
			return "/api/matches/{id}/guesses"
		}
		return "/api/matches/{id}"
	}
	return "static"
}

func metricMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodDelete, http.MethodOptions:
		return method
	default:
		return "OTHER"
	}
}

func writeGauge(w http.ResponseWriter, name, help string, value int) {
	fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s gauge\n%s %d\n", name, help, name, name, value)
}

func quote(value string) string {
	return strconv.Quote(value)
}

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *statusWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
