package services

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

func NewPhaseFourClients(cfg config.Config, logger *slog.Logger) (BlobStorageClient, RenderClient, error) {
	missing := make([]string, 0)
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "AZURE_BLOB_URL", value: cfg.AzureBlobURL},
		{name: "AZURE_BLOB_CONTAINER", value: cfg.AzureBlobContainer},
		{name: "AZURE_BLOB_SAS_TOKEN", value: cfg.AzureBlobSASToken},
		{name: "AZURE_RENDER_URL", value: cfg.AzureRenderURL},
		{name: "AZURE_RENDER_API_KEY", value: cfg.AzureRenderAPIKey},
	} {
		if strings.TrimSpace(field.value) == "" {
			missing = append(missing, field.name)
		}
	}
	if len(missing) > 0 {
		return nil, nil, fmt.Errorf(
			"phase 4 rendering is enabled by design and cannot run until Azure configuration is complete; set the missing environment variables before starting the server: %s",
			strings.Join(missing, ", "),
		)
	}

	httpClient := newPhaseFourHTTPClient()
	return NewAzureBlobStorageClient(cfg, logger, httpClient), NewAzureRenderClient(cfg, logger, httpClient), nil
}
