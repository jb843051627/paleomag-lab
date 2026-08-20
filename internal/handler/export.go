package handler

import "net/http"

func (a *API) exports(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.services.Exports.List(r.Context(), r.URL.Query().Get("specimen_id"))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var input struct {
			SpecimenID  string `json:"specimen_id"`
			Format      string `json:"format"`
			RequestedBy string `json:"requested_by"`
		}
		if err := decode(r, &input); err != nil {
			writeError(w, err)
			return
		}
		item, err := a.services.Exports.Request(r.Context(), input.SpecimenID, input.Format, input.RequestedBy)
		if err != nil {
			writeError(w, err)
			return
		}
		if err := a.workers.EnqueueExport(item.ID); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, item)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) exportAction(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path, "/api/exports/")
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		item, err := a.services.Exports.Get(r.Context(), parts[0])
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	if len(parts) == 2 && parts[1] == "run" && r.Method == http.MethodPost {
		item, err := a.services.Exports.Run(r.Context(), parts[0])
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
