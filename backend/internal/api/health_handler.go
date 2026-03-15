package api

import (
	"context"
	"net/http"
	"time"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
)

func newHealthHandler(version, providerProfile string, auditLogger services.JobAuditLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auditHealth := services.AuditHealth{Enabled: false, Status: "disabled", Details: "audit sink is not configured"}
		if auditLogger != nil {
			healthCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			auditHealth = auditLogger.Health(healthCtx)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status":           "healthy",
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
			"version":          version,
			"provider_profile": providerProfile,
			"audit": map[string]any{
				"enabled": auditHealth.Enabled,
				"status":  auditHealth.Status,
				"details": auditHealth.Details,
			},
		})
	}
}
