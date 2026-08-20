package handler

import (
	"net/http"
	"strconv"
	"time"
)

func (a *API) dashboard(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	item, err := a.services.Insights.Dashboard(r.Context(), limit, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *API) search(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.services.Insights.Search(r.Context(), r.URL.Query().Get("q"), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}
