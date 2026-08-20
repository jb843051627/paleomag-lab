package handler

import (
	"net/http"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

func (a *API) validateProtocol(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Protocol     model.DemagProtocol `json:"protocol"`
		InstrumentID string              `json:"instrument_id"`
	}
	if err := decode(r, &input); err != nil {
		writeError(w, err)
		return
	}
	if err := a.services.Protocols.ValidateForInstrument(r.Context(), input.Protocol, input.InstrumentID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Valid bool `json:"valid"`
		Steps int  `json:"steps"`
	}{true, input.Protocol.StepCount()})
}
