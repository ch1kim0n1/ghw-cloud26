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
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/select"):
			job, slot, err := service.SelectSlot(r.Context(), jobID, slotID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"job_id":                 job.ID,
				"slot_id":                slot.ID,
				"status":                 job.Status,
				"current_stage":          job.CurrentStage,
				"slot_status":            slot.Status,
				"suggested_product_line": slot.SuggestedProductLine,
				"message":                "slot selected and product line prepared",
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
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/generate"):
			var payload struct {
				ProductLineMode   string `json:"product_line_mode"`
				CustomProductLine string `json:"custom_product_line"`
			}
			if r.Body != nil {
				_ = json.NewDecoder(r.Body).Decode(&payload)
			}

			job, slot, err := service.StartGeneration(r.Context(), jobID, slotID, payload.ProductLineMode, payload.CustomProductLine)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"job_id":        job.ID,
				"slot_id":       slot.ID,
				"status":        job.Status,
				"current_stage": job.CurrentStage,
				"slot_status":   slot.Status,
				"message":       "cafai generation started",
			})
		default:
			writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "endpoint not implemented in this phase", map[string]any{
				"method": r.Method,
				"path":   r.URL.Path,
			})
		}
	}
}
