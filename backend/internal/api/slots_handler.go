package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

func newSlotsHandler(deps Dependencies) http.HandlerFunc {
	service := newJobService(deps)

	return func(w http.ResponseWriter, r *http.Request) {
		jobID := r.PathValue("job_id")
		slotID := r.PathValue("slot_id")

		switch {
		case r.Method == http.MethodGet && slotID == "":
			slots, err := service.ListSlots(r.Context(), jobID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"job_id": jobID,
				"slots":  slots,
			})
		case r.Method == http.MethodGet && slotID != "":
			slot, err := service.GetSlot(r.Context(), jobID, slotID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, slot)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/reject"):
			var payload struct {
				Note string `json:"note"`
			}
			if r.Body != nil {
				_ = json.NewDecoder(r.Body).Decode(&payload)
			}

			slot, err := service.RejectSlot(r.Context(), jobID, slotID, payload.Note)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"job_id":      jobID,
				"slot_id":     slot.ID,
				"slot_status": slot.Status,
				"message":     "slot rejected",
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/re-pick"):
			job, err := service.RequestRepick(r.Context(), jobID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"job_id":        job.ID,
				"status":        job.Status,
				"current_stage": job.CurrentStage,
				"message":       "re-pick requested",
			})
		default:
			writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "endpoint not implemented in this phase", map[string]any{
				"method": r.Method,
				"path":   r.URL.Path,
			})
		}
	}
}
