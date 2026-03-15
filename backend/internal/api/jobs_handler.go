package api

import (
	"net/http"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
)

func newJobsHandler(deps Dependencies) http.HandlerFunc {
	service := newJobService(deps)

	return func(w http.ResponseWriter, r *http.Request) {
		jobID := r.PathValue("job_id")

		switch r.URL.Path {
		case "/api/jobs/" + jobID:
			job, err := service.Get(r.Context(), jobID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, job)
		case "/api/jobs/" + jobID + "/logs":
			logs, err := service.ListLogs(r.Context(), jobID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"job_id": jobID,
				"logs":   logs,
			})
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", map[string]any{
				"method": r.Method,
			})
		}
	}
}

func newJobService(deps Dependencies) *services.JobService {
	frameExtractor := deps.AnchorFrameExtractor
	if frameExtractor == nil {
		frameExtractor = services.NewFFmpegAnchorFrameExtractor(deps.Config.ArtifactsDir)
	}
	blobClient := deps.BlobClient
	if blobClient == nil {
		blobClient = services.NewNoopBlobStorageClient(deps.Logger)
	}
	renderClient := deps.RenderClient
	if renderClient == nil {
		renderClient = services.NewNoopRenderClient(deps.Logger)
	}

	service := services.NewJobService(
		deps.DB,
		db.NewJobsRepository(deps.DB),
		db.NewCampaignsRepository(deps.DB),
		db.NewProductsRepository(deps.DB),
		db.NewJobLogsRepository(deps.DB),
		db.NewScenesRepository(deps.DB),
		db.NewSlotsRepository(deps.DB),
		db.NewPreviewsRepository(deps.DB),
		deps.AnalysisClient,
		deps.OpenAIClient,
		deps.MLClient,
		frameExtractor,
		services.NewLocalStorageService(),
		blobClient,
		renderClient,
		deps.Config.PreviewsDir,
		deps.Config.CacheDir,
	)
	service.SetAuditLogger(deps.AuditLogger)
	return service
}
