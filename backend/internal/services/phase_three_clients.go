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
	if strings.TrimSpace(cfg.AzureMLURL) == "" {
		return nil, fmt.Errorf(
			"phase 3 generation is enabled by design and cannot run until Azure configuration is complete; set the missing environment variables before starting the server: AZURE_ML_URL",
		)
	}

	httpClient := &http.Client{Timeout: 60 * time.Second}
	return NewAzureMLClient(cfg, logger, httpClient), nil
}
