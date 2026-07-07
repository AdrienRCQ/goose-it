package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/AdrienRCQ/goose-it/internal/contracts"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// type API contient les dépendances nécessaires aux handlers HTTP
type API struct {
	logger    *slog.Logger
	version   string
	startedAt time.Time
}

// Construction des routes HTTP :
func NewRouter(logger *slog.Logger, applicationVersion string, startedAt time.Time) http.Handler {
	api := &API{
		logger:    logger,
		version:   applicationVersion,
		startedAt: startedAt,
	}

	router := chi.NewRouter()
	// Attribution d'un uid à chaque requête
	router.Use(middleware.RequestID)

	// Interception des panic pour éviter l'arret complet du serveur
	router.Use(middleware.Recoverer)

	// Log les requetes
	router.Use(api.requestLogger)

	router.Get("/healthz", api.health)

	return router
}

// retourne l'état du server
func (a *API) health(w http.ResponseWriter, _ *http.Request) {
	response := contracts.HealthResponse{
		Status:        "ok",
		Service:       "goose-server",
		Version:       a.version,
		ServerTime:    time.Now().UTC(),
		UptimeSeconds: int64(time.Since(a.startedAt).Seconds()),
	}
	writeJSON(w, http.StatusOK, response)
}

// inscription de la response http in json format
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error(
			"failed to encode HTTP response",
			"error", err,
		)
	}
}

// Mémorisation du code http return
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(body)
}

// journalisation des requests HTTP
func (a *API) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()

			recorder := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(recorder, r)

			status := recorder.status

			if status == 0 {
				status = http.StatusOK
			}

			a.logger.Info(
				"HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration_ms",
				time.Since(startedAt).Milliseconds(),
				"remote_address", r.RemoteAddr,
				"request_id",
				middleware.GetReqID(r.Context()),
			)
		},
	)
}
