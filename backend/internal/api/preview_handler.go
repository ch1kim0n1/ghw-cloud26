package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

func newPreviewHandler(deps Dependencies) http.HandlerFunc {
	service := newJobService(deps)

	return func(w http.ResponseWriter, r *http.Request) {
		jobID := r.PathValue("job_id")

		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/preview/render"):
			var payload struct {
				SlotID string `json:"slot_id"`
			}
			if r.Body != nil {
				_ = json.NewDecoder(r.Body).Decode(&payload)
			}

			job, preview, err := service.StartPreviewRender(r.Context(), jobID, payload.SlotID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"job_id":        job.ID,
				"slot_id":       preview.SlotID,
				"status":        job.Status,
				"current_stage": job.CurrentStage,
				"message":       "preview render started",
			})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/preview/stream"):
			preview, err := service.GetPreview(r.Context(), jobID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			if !ensurePreviewReady(w, preview, jobID) {
				return
			}

			w.Header().Set("Content-Type", "video/mp4")
			w.Header().Set("Content-Disposition", `inline; filename="`+filepath.Base(preview.OutputVideoPath)+`"`)
			http.ServeFile(w, r, preview.OutputVideoPath)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/preview/download"):
			preview, err := service.GetPreview(r.Context(), jobID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			if !ensurePreviewReady(w, preview, jobID) {
				return
			}

			w.Header().Set("Content-Type", "video/mp4")
			w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(preview.OutputVideoPath)+`"`)
			http.ServeFile(w, r, preview.OutputVideoPath)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/preview"):
			preview, err := service.GetPreview(r.Context(), jobID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, preview)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", map[string]any{
				"method": r.Method,
			})
		}
	}
}

func ensurePreviewReady(w http.ResponseWriter, preview models.Preview, jobID string) bool {
	if preview.Status != "completed" || strings.TrimSpace(preview.OutputVideoPath) == "" {
		writeError(w, http.StatusConflict, constants.ErrorCodeInvalidRequest, "preview is not ready", map[string]any{
			"job_id":         jobID,
			"preview_status": preview.Status,
		})
		return false
	}
	if _, err := os.Stat(preview.OutputVideoPath); err != nil {
		writeError(w, http.StatusNotFound, constants.ErrorCodeResourceNotFound, "preview output file not found", map[string]any{
			"job_id": jobID,
			"path":   preview.OutputVideoPath,
		})
		return false
	}
	return true
}
