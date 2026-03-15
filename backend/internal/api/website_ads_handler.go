package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
)

func newWebsiteAdsHandler(deps Dependencies) http.HandlerFunc {
	service := services.NewWebsiteAdService(
		db.NewWebsiteAdsRepository(deps.DB),
		db.NewProductsRepository(deps.DB),
		deps.WebsiteAdsImageClient,
		services.NewLocalStorageService(),
		deps.Config.WebsiteAdsDir,
	)

	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.PathValue("format") != "":
			job, err := service.GetByID(r.Context(), r.PathValue("job_id"))
			if err != nil {
				writeServiceError(w, err)
				return
			}
			serveWebsiteAdAsset(w, r, job, r.PathValue("format"))
		case r.Method == http.MethodGet && strings.TrimSpace(r.PathValue("job_id")) != "":
			job, err := service.GetByID(r.Context(), r.PathValue("job_id"))
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, serializeWebsiteAdJob(job))
		case r.Method == http.MethodGet:
			jobs, err := service.List(r.Context())
			if err != nil {
				writeServiceError(w, err)
				return
			}
			response := make([]models.WebsiteAdJob, 0, len(jobs))
			for _, job := range jobs {
				response = append(response, serializeWebsiteAdJob(job))
			}
			writeJSON(w, http.StatusOK, map[string]any{"jobs": response})
		case r.Method == http.MethodPost:
			var payload struct {
				ProductID          string `json:"product_id"`
				ProductName        string `json:"product_name"`
				ProductDescription string `json:"product_description"`
				ArticleHeadline    string `json:"article_headline"`
				ArticleBody        string `json:"article_body"`
				BrandStyle         string `json:"brand_style"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body", nil)
				return
			}

			job, err := service.Create(r.Context(), services.CreateWebsiteAdInput{
				ProductID:          payload.ProductID,
				ProductName:        payload.ProductName,
				ProductDescription: payload.ProductDescription,
				ArticleHeadline:    payload.ArticleHeadline,
				ArticleBody:        payload.ArticleBody,
				BrandStyle:         payload.BrandStyle,
			})
			if err != nil {
				writeServiceError(w, err)
				return
			}

			writeJSON(w, http.StatusCreated, serializeWebsiteAdJob(job))
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", map[string]any{"method": r.Method})
		}
	}
}

func serializeWebsiteAdJob(job models.WebsiteAdJob) models.WebsiteAdJob {
	job.BannerImageURL = ""
	job.VerticalImageURL = ""
	if job.BannerImagePath != "" {
		job.BannerImageURL = "/api/website-ads/" + job.ID + "/assets/banner"
	}
	if job.VerticalImagePath != "" {
		job.VerticalImageURL = "/api/website-ads/" + job.ID + "/assets/vertical"
	}
	return job
}

func serveWebsiteAdAsset(w http.ResponseWriter, r *http.Request, job models.WebsiteAdJob, format string) {
	var path string
	switch format {
	case "banner":
		path = job.BannerImagePath
	case "vertical":
		path = job.VerticalImagePath
	default:
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "unknown website ad asset format", map[string]any{"format": format})
		return
	}

	if strings.TrimSpace(path) == "" {
		writeError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "website ad asset not found", map[string]any{
			"job_id": job.ID,
			"format": format,
		})
		return
	}

	http.ServeFile(w, r, path)
}
