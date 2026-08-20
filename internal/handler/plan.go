package handler

import (
	"net/http"

	"github.com/jb843051627/paleomag-lab/internal/service"
)

func (a *API) plans(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.services.Plans.ListBySpecimen(r.Context(), r.URL.Query().Get("specimen_id"))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var input service.CreatePlanInput
		if err := decode(r, &input); err != nil {
			writeError(w, err)
			return
		}
		item, err := a.services.Plans.Create(r.Context(), input)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) planAction(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path, "/api/plans/")
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		item, err := a.services.Plans.Get(r.Context(), parts[0])
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
	var item any
	var err error
	switch parts[1] {
	case "step":
		var input struct {
			Field       float64 `json:"field"`
			Temperature float64 `json:"temperature"`
			Note        string  `json:"note"`
		}
		if err = decode(r, &input); err == nil {
			item, err = a.services.Plans.AddStep(r.Context(), id, input.Field, input.Temperature, input.Note, actor(r))
		}
	case "prepare":
		item, err = a.services.Plans.Prepare(r.Context(), id, actor(r))
	case "start":
		item, err = a.services.Plans.Start(r.Context(), id, actor(r))
	case "complete":
		item, err = a.services.Plans.Complete(r.Context(), id, actor(r))
	case "reject":
		var input struct {
			Reason string `json:"reason"`
		}
		if err = decode(r, &input); err == nil {
			item, err = a.services.Plans.Reject(r.Context(), id, input.Reason, actor(r))
		}
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
