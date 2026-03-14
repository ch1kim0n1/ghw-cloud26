package services

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

func NewPhaseThreeClient(cfg config.Config, logger *slog.Logger) (MLClient, error) {
	profile := normalizeProviderProfile(cfg.ProviderProfile)
	if err := validateProviderProfile(profile); err != nil {
		return nil, err
	}
	httpClient := &http.Client{Timeout: 60 * time.Second}

	switch profile {
	case ProviderProfileAzure:
		if strings.TrimSpace(cfg.AzureMLURL) == "" {
			return nil, fmt.Errorf(
				"phase 3 generation is enabled by design and cannot run until azure configuration is complete; set the missing environment variables before starting the server: AZURE_ML_URL",
			)
		}
		azureClient := NewAzureMLClient(cfg, logger, httpClient)
		if strings.TrimSpace(cfg.HiggsfieldAPIKey) != "" && strings.TrimSpace(cfg.HiggsfieldAPISecret) != "" {
			blobClient := NewAzureBlobStorageClient(cfg, logger, newPhaseFourHTTPClient())
			higgsfieldClient := NewHiggsfieldClient(cfg, logger, httpClient, blobClient)
			return NewPriorityFallbackMLClient(GenerationProviderHiggsfield, higgsfieldClient, GenerationProviderAzureML, azureClient, logger), nil
		}
		return azureClient, nil
	case ProviderProfileVultr:
		missing := make([]string, 0)
		for _, field := range []struct {
			name  string
			value string
		}{
			{name: "VULTR_GENERATION_URL", value: cfg.VultrGenerationURL},
			{name: "VULTR_GENERATION_API_KEY", value: cfg.VultrGenerationAPIKey},
		} {
			if strings.TrimSpace(field.value) == "" {
				missing = append(missing, field.name)
			}
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf(
				"phase 3 generation is enabled by design and cannot run until vultr configuration is complete; set the missing environment variables before starting the server: %s",
				strings.Join(missing, ", "),
			)
		}
		return NewVultrGenerationClient(cfg, logger, httpClient), nil
	default:
		return nil, fmt.Errorf("unsupported provider profile %q", cfg.ProviderProfile)
	}
}
