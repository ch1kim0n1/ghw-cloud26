package api

import (
	"errors"
	"net/http"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
)

func newProductsHandler(deps Dependencies) http.HandlerFunc {
	service := services.NewProductService(
		db.NewProductsRepository(deps.DB),
		services.NewLocalStorageService(),
		deps.Config.UploadProductsDir,
	)

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			products, err := service.List(r.Context())
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"products": products,
			})
		case http.MethodPost:
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid multipart form", nil)
				return
			}

			input, err := parseProductInput(r, "", "", "", "", "", "")
			if err != nil {
				writeServiceError(w, err)
				return
			}

			product, err := service.Create(r.Context(), input)
			if err != nil {
				writeServiceError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, product)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", map[string]any{
				"method": r.Method,
			})
		}
	}
}

func parseProductInput(r *http.Request, nameField, descriptionField, categoryField, keywordsField, sourceURLField, imageField string) (services.CreateProductInput, error) {
	if nameField == "" {
		nameField = "name"
		descriptionField = "description"
		categoryField = "category"
		keywordsField = "context_keywords"
		sourceURLField = "source_url"
		imageField = "image_file"
	}

	input := services.CreateProductInput{
		Name:            r.FormValue(nameField),
		Description:     r.FormValue(descriptionField),
		Category:        r.FormValue(categoryField),
		ContextKeywords: services.ParseCommaSeparatedKeywords(r.FormValue(keywordsField)),
		SourceURL:       r.FormValue(sourceURLField),
	}

	file, header, err := r.FormFile(imageField)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return input, nil
		}
		return services.CreateProductInput{}, services.InvalidRequest("INVALID_REQUEST", "invalid file upload", map[string]any{
			"field": imageField,
		})
	}

	input.ImageHeader = header
	input.ImageFile = file
	return input, nil
}
