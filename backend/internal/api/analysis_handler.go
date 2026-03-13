package api

import "net/http"

func newAnalysisHandler(deps Dependencies) http.HandlerFunc {
	service := newJobService(deps)

	return func(w http.ResponseWriter, r *http.Request) {
		jobID := r.PathValue("job_id")

		job, err := service.StartAnalysis(r.Context(), jobID)
		if err != nil {
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"job_id":        job.ID,
			"status":        job.Status,
			"current_stage": job.CurrentStage,
			"message":       "analysis started",
		})
	}
}
