package api

import (
	"net/http"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
)

func newCampaignsHandler(deps Dependencies) http.HandlerFunc {
	service := services.NewCampaignService(
		deps.DB,
		db.NewCampaignsRepository(deps.DB),
		db.NewProductsRepository(deps.DB),
		db.NewJobsRepository(deps.DB),
		services.NewLocalStorageService(),
		services.NewMediaInspector(),
		deps.Config.UploadCampaignsDir,
		deps.Config.UploadProductsDir,
	)

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			campaignID := r.PathValue("campaign_id")
			campaign, err := service.Get(r.Context(), campaignID)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, campaign)
		case http.MethodPost:
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid multipart form", nil)
				return
			}

			videoFile, videoHeader, err := r.FormFile("video_file")
			if err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "video_file is required", map[string]any{
					"field": "video_file",
				})
				return
			}

			inlineProduct, productErr := parseProductInput(
				r,
				"product_name",
				"product_description",
				"product_category",
				"product_context_keywords",
				"product_source_url",
				"product_image_file",
			)
			if productErr != nil {
				writeServiceError(w, productErr)
				return
			}

			campaign, err := service.Create(r.Context(), services.CreateCampaignInput{
				Name:                    r.FormValue("name"),
				TargetAdDurationSeconds: r.FormValue("target_ad_duration_seconds"),
				ProductID:               r.FormValue("product_id"),
				VideoHeader:             videoHeader,
				VideoFile:               videoFile,
				InlineProduct:           inlineProduct,
			})
			if err != nil {
				writeServiceError(w, err)
				return
			}

			writeJSON(w, http.StatusOK, campaign)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", map[string]any{
				"method": r.Method,
			})
		}
	}
}
