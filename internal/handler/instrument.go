package handler

import (
	"net/http"

	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/service"
)

func (a *API) instruments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.services.Instruments.List(r.Context(), model.InstrumentStatus(r.URL.Query().Get("status")))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var input service.RegisterInstrumentInput
		if err := decode(r, &input); err != nil {
			writeError(w, err)
			return
		}
		item, err := a.services.Instruments.Register(r.Context(), input)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) instrumentAction(w http.ResponseWriter, r *http.Request) {
	parts := pathParts(r.URL.Path, "/api/instruments/")
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		item, err := a.services.Instruments.Get(r.Context(), parts[0])
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
	case "calibration":
		var input struct {
			Reference string `json:"reference"`
		}
		if err := decode(r, &input); err != nil {
			writeError(w, err)
			return
		}
		item, err := a.services.Instruments.BeginCalibration(r.Context(), id, input.Reference, actor(r))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		http.NotFound(w, r)
	}
}
