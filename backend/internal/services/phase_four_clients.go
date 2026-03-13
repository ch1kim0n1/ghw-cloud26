package services

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

func NewPhaseFourClients(cfg config.Config, logger *slog.Logger) (BlobStorageClient, RenderClient, error) {
	profile := normalizeProviderProfile(cfg.ProviderProfile)
	if err := validateProviderProfile(profile); err != nil {
		return nil, nil, err
	}
	httpClient := newPhaseFourHTTPClient()

	switch profile {
	case ProviderProfileAzure:
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
				"phase 4 rendering is enabled by design and cannot run until azure configuration is complete; set the missing environment variables before starting the server: %s",
				strings.Join(missing, ", "),
			)
		}
		return NewAzureBlobStorageClient(cfg, logger, httpClient), NewAzureRenderClient(cfg, logger, httpClient), nil
	case ProviderProfileVultr:
		missing := make([]string, 0)
		for _, field := range []struct {
			name  string
			value string
		}{
			{name: "VULTR_OBJECT_STORAGE_ENDPOINT", value: cfg.VultrObjectStorageEndpoint},
			{name: "VULTR_OBJECT_STORAGE_REGION", value: cfg.VultrObjectStorageRegion},
			{name: "VULTR_OBJECT_STORAGE_BUCKET", value: cfg.VultrObjectStorageBucket},
			{name: "VULTR_OBJECT_STORAGE_ACCESS_KEY", value: cfg.VultrObjectStorageAccessKey},
			{name: "VULTR_OBJECT_STORAGE_SECRET_KEY", value: cfg.VultrObjectStorageSecretKey},
			{name: "VULTR_RENDER_URL", value: cfg.VultrRenderURL},
			{name: "VULTR_RENDER_API_KEY", value: cfg.VultrRenderAPIKey},
		} {
			if strings.TrimSpace(field.value) == "" {
				missing = append(missing, field.name)
			}
		}
		if len(missing) > 0 {
			return nil, nil, fmt.Errorf(
				"phase 4 rendering is enabled by design and cannot run until vultr configuration is complete; set the missing environment variables before starting the server: %s",
				strings.Join(missing, ", "),
			)
		}
		return NewVultrObjectStorageClient(cfg, logger, httpClient), NewVultrRenderClient(cfg, logger, httpClient), nil
	default:
		return nil, nil, fmt.Errorf("unsupported provider profile %q", cfg.ProviderProfile)
	}
}
