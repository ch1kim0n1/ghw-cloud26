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
	CacheDir                     string
	PreviewsDir                  string
	WebsiteAdsDir                string
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
	NotionAPIBaseURL             string
	NotionAPIKey                 string
	NotionVersion                string
	NotionJobsDatabaseID         string
	NotionEventsDatabaseID       string
	NotionRequestTimeout         time.Duration
	HiggsfieldAPIKey             string
	HiggsfieldAPISecret          string
	HiggsfieldBaseURL            string
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
	HuggingFaceAPIToken          string
	HuggingFaceBaseURL           string
	HuggingFaceImageModel        string
	HuggingFaceRequestTimeout    time.Duration
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
		CacheDir:                     filepath.Join(repoRoot, "tmp", "cache"),
		PreviewsDir:                  filepath.Join(repoRoot, "tmp", "previews"),
		WebsiteAdsDir:                filepath.Join(repoRoot, "tmp", "website_ads"),
		AllowedOrigins:               splitCSV(getEnv("CAFAI_ALLOWED_ORIGINS", "http://localhost:5173")),
		WorkerInterval:               getDurationEnv("CAFAI_WORKER_INTERVAL", 5*time.Second),
		ShutdownTimeout:              getDurationEnv("CAFAI_SHUTDOWN_TIMEOUT", 10*time.Second),
		Version:                      version,
		AzureVideoIndexerURL:         os.Getenv("AZURE_VIDEO_INDEXER_URL"),
		AzureVideoIndexerAccountID:   os.Getenv("AZURE_VIDEO_INDEXER_ACCOUNT_ID"),
		AzureVideoIndexerLocation:    os.Getenv("AZURE_VIDEO_INDEXER_LOCATION"),
		AzureVideoIndexerAccessToken: os.Getenv("AZURE_VIDEO_INDEXER_ACCESS_TOKEN"),
		AzureOpenAIURL:               os.Getenv("AZURE_OPENAI_URL"),
		AzureOpenAIApiKey:            firstNonEmpty(os.Getenv("AZURE_OPENAI_API_KEY"), os.Getenv("AZURE_OPENAI_API_KEY1"), os.Getenv("AZURE_OPENAI_API_KEY2")),
		AzureOpenAIApiVersion:        getEnv("AZURE_OPENAI_API_VERSION", "2024-10-21"),
		AzureOpenAIDeployment:        os.Getenv("AZURE_OPENAI_DEPLOYMENT"),
		AzureMLURL:                   os.Getenv("AZURE_ML_URL"),
		AzureSpeechURL:               os.Getenv("AZURE_SPEECH_URL"),
		AzureBlobURL:                 os.Getenv("AZURE_BLOB_URL"),
		AzureBlobContainer:           os.Getenv("AZURE_BLOB_CONTAINER"),
		AzureBlobSASToken:            os.Getenv("AZURE_BLOB_SAS_TOKEN"),
		AzureRenderURL:               os.Getenv("AZURE_RENDER_URL"),
		AzureRenderAPIKey:            os.Getenv("AZURE_RENDER_API_KEY"),
		NotionAPIBaseURL:             getEnv("NOTION_API_BASE_URL", "https://api.notion.com/v1"),
		NotionAPIKey:                 os.Getenv("NOTION_API_KEY"),
		NotionVersion:                getEnv("NOTION_API_VERSION", "2022-06-28"),
		NotionJobsDatabaseID:         os.Getenv("NOTION_JOBS_DATABASE_ID"),
		NotionEventsDatabaseID:       os.Getenv("NOTION_EVENTS_DATABASE_ID"),
		NotionRequestTimeout:         getDurationEnv("NOTION_REQUEST_TIMEOUT", 5*time.Second),
		HiggsfieldAPIKey:             os.Getenv("HIGGSFIELD_API_KEY"),
		HiggsfieldAPISecret:          os.Getenv("HIGGSFIELD_API_SECRET"),
		HiggsfieldBaseURL:            getEnv("HIGGSFIELD_BASE_URL", "https://platform.higgsfield.ai"),
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
		HuggingFaceAPIToken:          firstNonEmpty(os.Getenv("HUGGINGFACE_API_TOKEN"), os.Getenv("HF_TOKEN")),
		HuggingFaceBaseURL:           getEnv("HUGGINGFACE_BASE_URL", "https://router.huggingface.co/hf-inference/models"),
		HuggingFaceImageModel:        getEnv("HUGGINGFACE_IMAGE_MODEL", "stabilityai/stable-diffusion-xl-base-1.0"),
		HuggingFaceRequestTimeout:    getDurationEnv("HUGGINGFACE_REQUEST_TIMEOUT", 90*time.Second),
	}, nil
}

func resolveRepoRoot() (string, error) {
	if explicit := os.Getenv("CAFAI_REPO_ROOT"); explicit != "" {
		return canonicalizePath(explicit)
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	base := filepath.Base(wd)
	switch base {
	case "backend", "frontend":
		return canonicalizePath(filepath.Dir(wd))
	default:
		return canonicalizePath(wd)
	}
}

func canonicalizePath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return resolved, nil
	}
	return absolute, nil
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
