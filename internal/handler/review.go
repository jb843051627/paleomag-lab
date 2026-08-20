package handler

import (
	"net/http"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

func (a *API) reviews(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.services.Reviews.List(r.Context(), r.URL.Query().Get("interpretation_id"))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var input struct {
			InterpretationID string               `json:"interpretation_id"`
			Reviewer         string               `json:"reviewer"`
			Decision         model.ReviewDecision `json:"decision"`
			Comment          string               `json:"comment"`
		}
		if err := decode(r, &input); err != nil {
			writeError(w, err)
			return
		}
		item, interpretation, err := a.services.Reviews.Decide(r.Context(), input.InterpretationID, input.Reviewer, input.Decision, input.Comment)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, struct {
			Review         model.Review         `json:"review"`
			Interpretation model.Interpretation `json:"interpretation"`
		}{item, interpretation})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
