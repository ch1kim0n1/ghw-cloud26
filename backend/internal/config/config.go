package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const version = "0.1.0-mvp"

type Config struct {
	RepoRoot                     string
	ServerAddr                   string
	DatabasePath                 string
	MigrationsDir                string
	UploadProductsDir            string
	UploadCampaignsDir           string
	ArtifactsDir                 string
	PreviewsDir                  string
	AllowedOrigins               []string
	WorkerInterval               time.Duration
	ShutdownTimeout              time.Duration
	Version                      string
	AzureVideoIndexerURL         string
	AzureVideoIndexerAccountID   string
	AzureVideoIndexerLocation    string
	AzureVideoIndexerAccessToken string
	AzureOpenAIURL               string
	AzureOpenAIApiKey            string
	AzureOpenAIApiVersion        string
	AzureOpenAIDeployment        string
	AzureMLURL                   string
	AzureSpeechURL               string
	AzureBlobURL                 string
	AzureRenderURL               string
}

func Load() (Config, error) {
	repoRoot, err := resolveRepoRoot()
	if err != nil {
		return Config{}, err
	}

	return Config{
		RepoRoot:                     repoRoot,
		ServerAddr:                   getEnv("CAFAI_SERVER_ADDR", ":8080"),
		DatabasePath:                 getEnv("CAFAI_DATABASE_PATH", filepath.Join(repoRoot, "tmp", "cafai_mvp.db")),
		MigrationsDir:                getEnv("CAFAI_MIGRATIONS_DIR", filepath.Join(repoRoot, "backend", "scripts", "migrations")),
		UploadProductsDir:            filepath.Join(repoRoot, "tmp", "uploads", "products"),
		UploadCampaignsDir:           filepath.Join(repoRoot, "tmp", "uploads", "campaigns"),
		ArtifactsDir:                 filepath.Join(repoRoot, "tmp", "artifacts"),
		PreviewsDir:                  filepath.Join(repoRoot, "tmp", "previews"),
		AllowedOrigins:               splitCSV(getEnv("CAFAI_ALLOWED_ORIGINS", "http://localhost:5173")),
		WorkerInterval:               getDurationEnv("CAFAI_WORKER_INTERVAL", 5*time.Second),
		ShutdownTimeout:              getDurationEnv("CAFAI_SHUTDOWN_TIMEOUT", 10*time.Second),
		Version:                      version,
		AzureVideoIndexerURL:         os.Getenv("AZURE_VIDEO_INDEXER_URL"),
		AzureVideoIndexerAccountID:   os.Getenv("AZURE_VIDEO_INDEXER_ACCOUNT_ID"),
		AzureVideoIndexerLocation:    os.Getenv("AZURE_VIDEO_INDEXER_LOCATION"),
		AzureVideoIndexerAccessToken: os.Getenv("AZURE_VIDEO_INDEXER_ACCESS_TOKEN"),
		AzureOpenAIURL:               os.Getenv("AZURE_OPENAI_URL"),
		AzureOpenAIApiKey:            os.Getenv("AZURE_OPENAI_API_KEY"),
		AzureOpenAIApiVersion:        getEnv("AZURE_OPENAI_API_VERSION", "2024-10-21"),
		AzureOpenAIDeployment:        os.Getenv("AZURE_OPENAI_DEPLOYMENT"),
		AzureMLURL:                   os.Getenv("AZURE_ML_URL"),
		AzureSpeechURL:               os.Getenv("AZURE_SPEECH_URL"),
		AzureBlobURL:                 os.Getenv("AZURE_BLOB_URL"),
		AzureRenderURL:               os.Getenv("AZURE_RENDER_URL"),
	}, nil
}

func resolveRepoRoot() (string, error) {
	if explicit := os.Getenv("CAFAI_REPO_ROOT"); explicit != "" {
		return filepath.Abs(explicit)
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	base := filepath.Base(wd)
	switch base {
	case "backend", "frontend":
		return filepath.Dir(wd), nil
	default:
		return wd, nil
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	results := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			results = append(results, trimmed)
		}
	}
	if len(results) == 0 {
		return []string{"http://localhost:5173"}
	}
	return results
}
