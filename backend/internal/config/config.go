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
	ProviderProfile              string
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
	AzureBlobContainer           string
	AzureBlobSASToken            string
	AzureRenderURL               string
	AzureRenderAPIKey            string
	VultrAnalysisURL             string
	VultrAnalysisAPIKey          string
	VultrLLMURL                  string
	VultrLLMAPIKey               string
	VultrGenerationURL           string
	VultrGenerationAPIKey        string
	VultrObjectStorageEndpoint   string
	VultrObjectStorageRegion     string
	VultrObjectStorageBucket     string
	VultrObjectStorageAccessKey  string
	VultrObjectStorageSecretKey  string
	VultrRenderURL               string
	VultrRenderAPIKey            string
}

func Load() (Config, error) {
	repoRoot, err := resolveRepoRoot()
	if err != nil {
		return Config{}, err
	}

	return Config{
		RepoRoot:                     repoRoot,
		ProviderProfile:              normalizeProviderProfile(getEnv("CAFAI_PROVIDER_PROFILE", "azure")),
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
		AzureBlobContainer:           os.Getenv("AZURE_BLOB_CONTAINER"),
		AzureBlobSASToken:            os.Getenv("AZURE_BLOB_SAS_TOKEN"),
		AzureRenderURL:               os.Getenv("AZURE_RENDER_URL"),
		AzureRenderAPIKey:            os.Getenv("AZURE_RENDER_API_KEY"),
		VultrAnalysisURL:             os.Getenv("VULTR_ANALYSIS_URL"),
		VultrAnalysisAPIKey:          os.Getenv("VULTR_ANALYSIS_API_KEY"),
		VultrLLMURL:                  os.Getenv("VULTR_LLM_URL"),
		VultrLLMAPIKey:               os.Getenv("VULTR_LLM_API_KEY"),
		VultrGenerationURL:           os.Getenv("VULTR_GENERATION_URL"),
		VultrGenerationAPIKey:        os.Getenv("VULTR_GENERATION_API_KEY"),
		VultrObjectStorageEndpoint:   os.Getenv("VULTR_OBJECT_STORAGE_ENDPOINT"),
		VultrObjectStorageRegion:     getEnv("VULTR_OBJECT_STORAGE_REGION", "ewr1"),
		VultrObjectStorageBucket:     os.Getenv("VULTR_OBJECT_STORAGE_BUCKET"),
		VultrObjectStorageAccessKey:  os.Getenv("VULTR_OBJECT_STORAGE_ACCESS_KEY"),
		VultrObjectStorageSecretKey:  os.Getenv("VULTR_OBJECT_STORAGE_SECRET_KEY"),
		VultrRenderURL:               os.Getenv("VULTR_RENDER_URL"),
		VultrRenderAPIKey:            os.Getenv("VULTR_RENDER_API_KEY"),
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

func normalizeProviderProfile(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		return "azure"
	}
	return trimmed
}
