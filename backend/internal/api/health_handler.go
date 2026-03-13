package api

import (
	"net/http"
	"time"
)

func newHealthHandler(version, providerProfile string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":           "healthy",
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
			"version":          version,
			"provider_profile": providerProfile,
		})
	}
}
