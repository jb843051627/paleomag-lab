package handler

import (
	"context"
	"net/http"

	"github.com/jb843051627/paleomag-lab/internal/service"
	"github.com/jb843051627/paleomag-lab/internal/store"
	"github.com/jb843051627/paleomag-lab/internal/web"
	"github.com/jb843051627/paleomag-lab/internal/worker"
)

type API struct {
	services *service.Services
	workers  *worker.Manager
	db       *store.DB
}

func New(services *service.Services, workers *worker.Manager, db *store.DB) *API {
	return &API{services: services, workers: workers, db: db}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("/", a.index)
	mux.HandleFunc("GET /static/", a.static)
	mux.HandleFunc("/api/specimens", a.specimens)
	mux.HandleFunc("/api/specimens/", a.specimenAction)
	mux.HandleFunc("/api/plans", a.plans)
	mux.HandleFunc("/api/plans/", a.planAction)
	mux.HandleFunc("/api/instruments", a.instruments)
	mux.HandleFunc("/api/instruments/", a.instrumentAction)
	mux.HandleFunc("/api/runs", a.runs)
	mux.HandleFunc("/api/runs/", a.runAction)
	mux.HandleFunc("/api/interpretations", a.interpretations)
	mux.HandleFunc("/api/interpretations/", a.interpretationAction)
	mux.HandleFunc("/api/reviews", a.reviews)
	mux.HandleFunc("/api/exports", a.exports)
	mux.HandleFunc("/api/exports/", a.exportAction)
	mux.HandleFunc("GET /api/dashboard", a.dashboard)
	mux.HandleFunc("GET /api/search", a.search)
	mux.HandleFunc("GET /api/reports/", a.report)
	mux.HandleFunc("POST /api/protocols/validate", a.validateProtocol)
	return logging(mux)
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.db.HealthCheck(r.Context()))
}

func (a *API) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := web.FS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "page unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (a *API) static(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]
	data, err := web.FS.ReadFile(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	contentType := "text/plain; charset=utf-8"
	if len(name) >= 4 && name[len(name)-4:] == ".css" {
		contentType = "text/css; charset=utf-8"
	}
	if len(name) >= 3 && name[len(name)-3:] == ".js" {
		contentType = "application/javascript; charset=utf-8"
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), requestKey{}, r.Method+" "+r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type requestKey struct{}
