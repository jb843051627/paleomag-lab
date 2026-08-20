package handler

import (
	"net/http"

	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/service"
)

func (a *API) runs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var input service.StartRunInput
		if err := decode(r, &input); err != nil {
			writeError(w, err)
			return
		}
		item, err := a.services.Measurements.StartRun(r.Context(), input, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) runAction(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path, "/api/runs/")
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}
	runID := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		item, err := a.services.Measurements.GetRun(r.Context(), runID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	if len(parts) == 2 && parts[1] == "measurements" {
		switch r.Method {
		case http.MethodGet:
			items, err := a.services.Measurements.List(r.Context(), model.MeasurementFilter{RunID: runID})
			if err != nil {
				writeError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, items)
		case http.MethodPost:
			var item model.Measurement
			if err := decode(r, &item); err != nil {
				writeError(w, err)
				return
			}
			item.RunID = runID
			if err := a.services.Measurements.Record(r.Context(), item, actor(r)); err != nil {
				writeError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, item)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}
	if len(parts) == 2 && parts[1] == "quality" && r.Method == http.MethodGet {
		item, err := a.services.Insights.Quality(r.Context(), runID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	if len(parts) != 2 || r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	switch parts[1] {
	case "complete":
		item, summary, err := a.services.Measurements.CompleteRun(r.Context(), runID, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, struct {
			Run     model.MeasurementRun     `json:"run"`
			Summary model.MeasurementSummary `json:"summary"`
		}{item, summary})
	case "accept":
		item, err := a.services.Measurements.AcceptRun(r.Context(), runID, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case "rework":
		item, err := a.services.Measurements.MarkRework(r.Context(), runID, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	default:
		http.NotFound(w, r)
	}
}
