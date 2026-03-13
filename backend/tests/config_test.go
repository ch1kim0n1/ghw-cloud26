package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

func TestLoadConfigDefaultsAndOverrides(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CAFAI_REPO_ROOT", root)
	t.Setenv("CAFAI_ALLOWED_ORIGINS", "http://localhost:5173,https://example.com")
	t.Setenv("CAFAI_WORKER_INTERVAL", "2s")
	t.Setenv("CAFAI_SERVER_ADDR", ":9090")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ServerAddr != ":9090" {
		t.Fatalf("expected server addr override, got %q", cfg.ServerAddr)
	}
	if cfg.WorkerInterval != 2*time.Second {
		t.Fatalf("expected worker interval override, got %v", cfg.WorkerInterval)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Fatalf("expected 2 allowed origins, got %d", len(cfg.AllowedOrigins))
	}

	expectedDB := filepath.Join(root, "tmp", "cafai_mvp.db")
	if cfg.DatabasePath != expectedDB {
		t.Fatalf("expected database path %q, got %q", expectedDB, cfg.DatabasePath)
	}
}

func TestLoadConfigFromBackendWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	backendDir := filepath.Join(root, "backend")
	if err := os.MkdirAll(backendDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	if err := os.Chdir(backendDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	t.Setenv("CAFAI_REPO_ROOT", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.RepoRoot != root {
		t.Fatalf("expected repo root %q, got %q", root, cfg.RepoRoot)
	}
}
