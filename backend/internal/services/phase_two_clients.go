package services

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

func NewPhaseTwoClients(cfg config.Config, logger *slog.Logger) (AnalysisClient, OpenAIClient, error) {
	missing := make([]string, 0)
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "AZURE_VIDEO_INDEXER_URL", value: cfg.AzureVideoIndexerURL},
		{name: "AZURE_VIDEO_INDEXER_ACCOUNT_ID", value: cfg.AzureVideoIndexerAccountID},
		{name: "AZURE_VIDEO_INDEXER_LOCATION", value: cfg.AzureVideoIndexerLocation},
		{name: "AZURE_VIDEO_INDEXER_ACCESS_TOKEN", value: cfg.AzureVideoIndexerAccessToken},
		{name: "AZURE_OPENAI_URL", value: cfg.AzureOpenAIURL},
		{name: "AZURE_OPENAI_API_KEY", value: cfg.AzureOpenAIApiKey},
		{name: "AZURE_OPENAI_DEPLOYMENT", value: cfg.AzureOpenAIDeployment},
	} {
		if strings.TrimSpace(field.value) == "" {
			missing = append(missing, field.name)
		}
	}
	if len(missing) > 0 {
		return nil, nil, fmt.Errorf(
			"phase 2 analysis is enabled by design and cannot run until Azure configuration is complete; set the missing environment variables before starting the server: %s",
			strings.Join(missing, ", "),
		)
	}

	httpClient := &http.Client{Timeout: 60 * time.Second}
	return NewAzureVideoIndexerClient(cfg, logger, httpClient), NewAzureOpenAIClient(cfg, logger, httpClient), nil
}
