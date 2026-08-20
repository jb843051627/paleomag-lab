package handler

import (
	"net/http"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

func (a *API) interpretations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.services.Interpretations.ListBySpecimen(r.Context(), r.URL.Query().Get("specimen_id"), model.InterpretationStatus(r.URL.Query().Get("status")))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var input struct {
			RunID string `json:"run_id"`
		}
		if err := decode(r, &input); err != nil {
			writeError(w, err)
			return
		}
		item, err := a.services.Interpretations.Generate(r.Context(), input.RunID, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) interpretationAction(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path, "/api/interpretations/")
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		item, err := a.services.Interpretations.Get(r.Context(), parts[0])
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
	id := parts[0]
	switch parts[1] {
	case "submit":
		item, err := a.services.Interpretations.Submit(r.Context(), id, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case "notes":
		var input struct {
			Notes string `json:"notes"`
		}
		if err := decode(r, &input); err != nil {
			writeError(w, err)
			return
		}
		item, err := a.services.Interpretations.UpdateNotes(r.Context(), id, input.Notes, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	default:
		http.NotFound(w, r)
	}
}
