package api

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
)

type Dependencies struct {
	Config               config.Config
	Logger               *slog.Logger
	DB                   *sql.DB
	AnalysisClient       services.AnalysisClient
	OpenAIClient         services.OpenAIClient
	MLClient             services.MLClient
	AnchorFrameExtractor services.AnchorFrameExtractor
	SpeechClient         services.SpeechClient
	BlobClient           services.BlobStorageClient
	RenderClient         services.RenderClient
	CafaiGenerator       services.CafaiGenerator
	AuditLogger          services.JobAuditLogger
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", newHealthHandler(deps.Config.Version, deps.Config.ProviderProfile, deps.AuditLogger))
	registerRoutes(mux, deps)

	handler := corsMiddleware(deps.Config.AllowedOrigins, mux)
	handler = loggingMiddleware(deps.Logger, handler)
	handler = recoverMiddleware(deps.Logger, handler)
	return handler
}

func registerRoutes(mux *http.ServeMux, deps Dependencies) {
	products := newProductsHandler(deps)
	campaigns := newCampaignsHandler(deps)
	jobs := newJobsHandler(deps)
	analysis := newAnalysisHandler(deps)
	slots := newSlotsHandler(deps)
	preview := newPreviewHandler(deps)

	mux.HandleFunc("POST /api/products", products)
	mux.HandleFunc("GET /api/products", products)
	mux.HandleFunc("POST /api/campaigns", campaigns)
	mux.HandleFunc("GET /api/campaigns/{campaign_id}", campaigns)
	mux.HandleFunc("GET /api/jobs/{job_id}", jobs)
	mux.HandleFunc("GET /api/jobs/{job_id}/logs", jobs)
	mux.HandleFunc("POST /api/jobs/{job_id}/start-analysis", analysis)
	mux.HandleFunc("GET /api/jobs/{job_id}/slots", slots)
	mux.HandleFunc("GET /api/jobs/{job_id}/slots/{slot_id}", slots)
	mux.HandleFunc("POST /api/jobs/{job_id}/slots/manual-select", slots)
	mux.HandleFunc("POST /api/jobs/{job_id}/slots/manual-import", slots)
	mux.HandleFunc("POST /api/jobs/{job_id}/slots/{slot_id}/select", slots)
	mux.HandleFunc("POST /api/jobs/{job_id}/slots/{slot_id}/reject", slots)
	mux.HandleFunc("POST /api/jobs/{job_id}/slots/re-pick", slots)
	mux.HandleFunc("POST /api/jobs/{job_id}/slots/{slot_id}/generate", slots)
	mux.HandleFunc("POST /api/jobs/{job_id}/preview/render", preview)
	mux.HandleFunc("GET /api/jobs/{job_id}/preview", preview)
	mux.HandleFunc("GET /api/jobs/{job_id}/preview/stream", preview)
	mux.HandleFunc("GET /api/jobs/{job_id}/preview/download", preview)
}

func notImplementedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, constants.ErrorCodeNotImplemented, "endpoint not implemented in phase 0", map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
		})
	}
}
