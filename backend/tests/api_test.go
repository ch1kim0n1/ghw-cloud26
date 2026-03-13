package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/api"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	handler := api.NewRouter(api.Dependencies{
		Config: testConfig(t.TempDir()),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:     &sql.DB{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if payload["status"] != "healthy" {
		t.Fatalf("expected healthy status, got %#v", payload["status"])
	}
	if payload["version"] != "0.1.0-mvp" {
		t.Fatalf("expected version 0.1.0-mvp, got %#v", payload["version"])
	}
}

func TestPlaceholderRouteReturnsStandardEnvelope(t *testing.T) {
	handler := api.NewRouter(api.Dependencies{
		Config: testConfig(t.TempDir()),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:     &sql.DB{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}

	var payload api.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if payload.ErrorCode != "NOT_IMPLEMENTED" {
		t.Fatalf("expected NOT_IMPLEMENTED, got %q", payload.ErrorCode)
	}
}

func TestCORSPreflight(t *testing.T) {
	handler := api.NewRouter(api.Dependencies{
		Config: testConfig(t.TempDir()),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:     &sql.DB{},
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/products", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected allow origin header, got %q", got)
	}
}

func testConfig(root string) config.Config {
	return config.Config{
		RepoRoot:             root,
		ServerAddr:           ":8080",
		DatabasePath:         root + "/tmp/cafai_mvp.db",
		MigrationsDir:        root + "/backend/scripts/migrations",
		UploadProductsDir:    root + "/tmp/uploads/products",
		UploadCampaignsDir:   root + "/tmp/uploads/campaigns",
		ArtifactsDir:         root + "/tmp/artifacts",
		PreviewsDir:          root + "/tmp/previews",
		AllowedOrigins:       []string{"http://localhost:5173"},
		WorkerInterval:       5,
		ShutdownTimeout:      10,
		Version:              "0.1.0-mvp",
		AzureVideoIndexerURL: "",
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

func applyMigrations(t *testing.T, database *sql.DB) {
	t.Helper()
	if err := database.PingContext(context.Background()); err != nil {
		t.Fatalf("PingContext() error = %v", err)
	}
}
