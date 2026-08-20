package handler

import (
	"net/http"
	"strconv"

	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/service"
)

func (a *API) specimens(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		filter := model.SpecimenFilter{Status: model.SpecimenStatus(r.URL.Query().Get("status")), Site: r.URL.Query().Get("site"), Lithology: r.URL.Query().Get("lithology"), Search: r.URL.Query().Get("search")}
		filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
		filter.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
		items, err := a.services.Specimens.List(r.Context(), filter)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var input service.CreateSpecimenInput
		if err := decode(r, &input); err != nil {
			writeError(w, err)
			return
		}
		item, err := a.services.Specimens.Create(r.Context(), input)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) specimenAction(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path, "/api/specimens/")
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		item, err := a.services.Specimens.Get(r.Context(), id)
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
	case "orientation":
		var input model.Orientation
		if err := decode(r, &input); err != nil {
			writeError(w, err)
			return
		}
		item, err := a.services.Specimens.Orient(r.Context(), id, input, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case "start":
		item, err := a.services.Specimens.StartWork(r.Context(), id, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	case "archive":
		item, err := a.services.Specimens.Archive(r.Context(), id, actor(r))
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
		item, err := a.services.Specimens.UpdateNotes(r.Context(), id, input.Notes, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	default:
		http.NotFound(w, r)
	}
}
