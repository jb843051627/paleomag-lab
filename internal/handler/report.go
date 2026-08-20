package handler

import "net/http"

func (a *API) report(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path, "/api/reports/")
	if len(parts) != 1 || r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	snapshot, err := a.services.Reports.Snapshot(r.Context(), parts[0])
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}
