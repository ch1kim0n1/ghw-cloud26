package tests

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/worker"
)

func TestNoopClientsReturnPlaceholderError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	testCases := []struct {
		name string
		run  func() error
	}{
		{
			name: "analysis",
			run: func() error {
				_, err := services.NewNoopAnalysisClient(logger).SubmitAnalysis(context.Background(), services.AnalysisRequest{JobID: "job_1"})
				return err
			},
		},
		{
			name: "openai",
			run: func() error {
				_, err := services.NewNoopOpenAIClient(logger).Complete(context.Background(), services.OpenAIRequest{JobID: "job_1"})
				return err
			},
		},
		{
			name: "ml",
			run: func() error {
				client := services.NewNoopMLClient(logger)
				if _, err := client.SubmitGeneration(context.Background(), services.GenerationRequest{JobID: "job_1"}); err != nil {
					return err
				}
				_, err := client.PollGeneration(context.Background(), services.GenerationPollRequest{JobID: "job_1", SlotID: "slot_1", RequestID: "req_1"})
				return err
			},
		},
		{
			name: "speech",
			run: func() error {
				_, err := services.NewNoopSpeechClient(logger).Synthesize(context.Background(), services.SpeechRequest{JobID: "job_1"})
				return err
			},
		},
		{
			name: "blob",
			run: func() error {
				client := services.NewNoopBlobStorageClient(logger)
				if _, err := client.Upload(context.Background(), services.BlobUploadRequest{JobID: "job_1"}); err != nil {
					return err
				}
				_, err := client.Download(context.Background(), services.BlobDownloadRequest{JobID: "job_1", BlobURI: "https://example.com/blob.mp4"})
				return err
			},
		},
		{
			name: "render",
			run: func() error {
				client := services.NewNoopRenderClient(logger)
				if _, err := client.SubmitRender(context.Background(), services.RenderRequest{JobID: "job_1"}); err != nil {
					return err
				}
				_, err := client.PollRender(context.Background(), services.RenderPollRequest{JobID: "job_1", SlotID: "slot_1", RequestID: "req_1"})
				return err
			},
		},
		{
			name: "cafai",
			run: func() error {
				_, err := services.NewNoopCafaiGenerator(logger).Generate(context.Background(), services.GenerationRequest{JobID: "job_1"})
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run()
			if !errors.Is(err, services.ErrPlaceholderClient) {
				t.Fatalf("expected ErrPlaceholderClient, got %v", err)
			}
		})
	}
}

func TestWorkerTicksAndStops(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	processor := worker.NewProcessor(logger, 10*time.Millisecond)

	ticks := make(chan struct{}, 1)
	processor.SetOnTick(func(context.Context) {
		select {
		case ticks <- struct{}{}:
		default:
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		processor.Run(ctx)
	}()

	select {
	case <-ticks:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("worker did not tick in time")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("worker did not stop in time")
	}
}

func TestLocalStorageServiceSavesAndReadsBackFiles(t *testing.T) {
	root := t.TempDir()
	storage := services.NewLocalStorageService()

	targetPath := filepath.Join(root, "tmp", "artifacts", "phase0-check.txt")
	expected := []byte("phase 0 save and read back")

	if err := storage.Save(targetPath, expected); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	actual, err := storage.Read(targetPath)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if string(actual) != string(expected) {
		t.Fatalf("unexpected file contents: got %q want %q", string(actual), string(expected))
	}
}

func TestNewPhaseTwoClientsRequiresCompleteConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	_, _, err := services.NewPhaseTwoClients(config.Config{}, logger)
	if err == nil {
		t.Fatal("expected configuration error, got nil")
	}

	message := err.Error()
	for _, expected := range []string{
		"phase 2 analysis is enabled by design",
		"AZURE_VIDEO_INDEXER_URL",
		"AZURE_VIDEO_INDEXER_ACCOUNT_ID",
		"AZURE_VIDEO_INDEXER_LOCATION",
		"AZURE_VIDEO_INDEXER_ACCESS_TOKEN",
		"AZURE_OPENAI_URL",
		"AZURE_OPENAI_API_KEY",
		"AZURE_OPENAI_DEPLOYMENT",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected error to mention %q, got %q", expected, message)
		}
	}
}

func TestNewPhaseTwoClientsReturnsAzureClientsWhenConfigured(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{
		ProviderProfile:              services.ProviderProfileAzure,
		AzureVideoIndexerURL:         "https://video.example.com",
		AzureVideoIndexerAccountID:   "account",
		AzureVideoIndexerLocation:    "trial",
		AzureVideoIndexerAccessToken: "token",
		AzureOpenAIURL:               "https://openai.example.com",
		AzureOpenAIApiKey:            "api-key",
		AzureOpenAIApiVersion:        "2024-10-21",
		AzureOpenAIDeployment:        "gpt-slot-ranker",
	}

	analysisClient, openAIClient, err := services.NewPhaseTwoClients(cfg, logger)
	if err != nil {
		t.Fatalf("NewPhaseTwoClients() error = %v", err)
	}
	if analysisClient == nil || openAIClient == nil {
		t.Fatal("expected both phase 2 clients to be created")
	}
}

func TestNewPhaseTwoClientsReturnsVultrClientsWhenConfigured(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{
		ProviderProfile:     services.ProviderProfileVultr,
		VultrAnalysisURL:    "https://analysis.example.com",
		VultrAnalysisAPIKey: "analysis-key",
		VultrLLMURL:         "https://llm.example.com",
		VultrLLMAPIKey:      "llm-key",
	}

	analysisClient, openAIClient, err := services.NewPhaseTwoClients(cfg, logger)
	if err != nil {
		t.Fatalf("NewPhaseTwoClients() error = %v", err)
	}
	if _, ok := analysisClient.(*services.VultrAnalysisClient); !ok {
		t.Fatalf("expected VultrAnalysisClient, got %T", analysisClient)
	}
	if _, ok := openAIClient.(*services.VultrLLMClient); !ok {
		t.Fatalf("expected VultrLLMClient, got %T", openAIClient)
	}
}

func TestNewPhaseTwoClientsRequireOnlyVultrConfigWhenSelected(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	_, _, err := services.NewPhaseTwoClients(config.Config{ProviderProfile: services.ProviderProfileVultr}, logger)
	if err == nil {
		t.Fatal("expected configuration error, got nil")
	}

	message := err.Error()
	for _, expected := range []string{"VULTR_ANALYSIS_URL", "VULTR_ANALYSIS_API_KEY", "VULTR_LLM_URL", "VULTR_LLM_API_KEY"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected Vultr error to mention %q, got %q", expected, message)
		}
	}
	if strings.Contains(message, "AZURE_") {
		t.Fatalf("expected Vultr-only config error, got %q", message)
	}
}

func TestNewPhaseThreeClientRequiresCompleteConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	_, err := services.NewPhaseThreeClient(config.Config{}, logger)
	if err == nil {
		t.Fatal("expected configuration error, got nil")
	}
	if !strings.Contains(err.Error(), "phase 3 generation is enabled by design") || !strings.Contains(err.Error(), "AZURE_ML_URL") {
		t.Fatalf("unexpected phase 3 config error: %v", err)
	}
}

func TestNewPhaseThreeClientReturnsAzureClientWhenConfigured(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	client, err := services.NewPhaseThreeClient(config.Config{
		ProviderProfile: services.ProviderProfileAzure,
		AzureMLURL:      "https://ml.example.com",
	}, logger)
	if err != nil {
		t.Fatalf("NewPhaseThreeClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("expected Azure ML client to be created")
	}
}

func TestNewPhaseThreeClientReturnsVultrClientWhenConfigured(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	client, err := services.NewPhaseThreeClient(config.Config{
		ProviderProfile:       services.ProviderProfileVultr,
		VultrGenerationURL:    "https://generation.example.com",
		VultrGenerationAPIKey: "generation-key",
	}, logger)
	if err != nil {
		t.Fatalf("NewPhaseThreeClient() error = %v", err)
	}
	if _, ok := client.(*services.VultrGenerationClient); !ok {
		t.Fatalf("expected VultrGenerationClient, got %T", client)
	}
}

func TestNewPhaseThreeClientRequiresOnlyVultrConfigWhenSelected(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	_, err := services.NewPhaseThreeClient(config.Config{ProviderProfile: services.ProviderProfileVultr}, logger)
	if err == nil {
		t.Fatal("expected configuration error, got nil")
	}
	if !strings.Contains(err.Error(), "VULTR_GENERATION_URL") || !strings.Contains(err.Error(), "VULTR_GENERATION_API_KEY") || strings.Contains(err.Error(), "AZURE_") {
		t.Fatalf("unexpected Vultr phase 3 config error: %v", err)
	}
}

func TestNewPhaseFourClientsRequiresCompleteConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	_, _, err := services.NewPhaseFourClients(config.Config{}, logger)
	if err == nil {
		t.Fatal("expected configuration error, got nil")
	}

	message := err.Error()
	for _, expected := range []string{
		"phase 4 rendering is enabled by design",
		"AZURE_BLOB_URL",
		"AZURE_BLOB_CONTAINER",
		"AZURE_BLOB_SAS_TOKEN",
		"AZURE_RENDER_URL",
		"AZURE_RENDER_API_KEY",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected error to mention %q, got %q", expected, message)
		}
	}
}

func TestNewPhaseFourClientsReturnAzureClientsWhenConfigured(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	blobClient, renderClient, err := services.NewPhaseFourClients(config.Config{
		ProviderProfile:    services.ProviderProfileAzure,
		AzureBlobURL:       "https://blob.example.com",
		AzureBlobContainer: "cafai",
		AzureBlobSASToken:  "sv=test&sig=test",
		AzureRenderURL:     "https://render.example.com",
		AzureRenderAPIKey:  "api-key",
	}, logger)
	if err != nil {
		t.Fatalf("NewPhaseFourClients() error = %v", err)
	}
	if blobClient == nil || renderClient == nil {
		t.Fatal("expected both phase 4 clients to be created")
	}
}

func TestNewPhaseFourClientsReturnVultrClientsWhenConfigured(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	blobClient, renderClient, err := services.NewPhaseFourClients(config.Config{
		ProviderProfile:             services.ProviderProfileVultr,
		VultrObjectStorageEndpoint:  "https://ewr1.vultrobjects.com",
		VultrObjectStorageRegion:    "ewr1",
		VultrObjectStorageBucket:    "hack-bucket",
		VultrObjectStorageAccessKey: "access-key",
		VultrObjectStorageSecretKey: "secret-key",
		VultrRenderURL:              "https://render.example.com",
		VultrRenderAPIKey:           "render-key",
	}, logger)
	if err != nil {
		t.Fatalf("NewPhaseFourClients() error = %v", err)
	}
	if _, ok := blobClient.(*services.VultrObjectStorageClient); !ok {
		t.Fatalf("expected VultrObjectStorageClient, got %T", blobClient)
	}
	if _, ok := renderClient.(*services.VultrRenderClient); !ok {
		t.Fatalf("expected VultrRenderClient, got %T", renderClient)
	}
}

func TestNewPhaseFourClientsRequireOnlyVultrConfigWhenSelected(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	_, _, err := services.NewPhaseFourClients(config.Config{ProviderProfile: services.ProviderProfileVultr}, logger)
	if err == nil {
		t.Fatal("expected configuration error, got nil")
	}

	message := err.Error()
	for _, expected := range []string{
		"VULTR_OBJECT_STORAGE_ENDPOINT",
		"VULTR_OBJECT_STORAGE_REGION",
		"VULTR_OBJECT_STORAGE_BUCKET",
		"VULTR_OBJECT_STORAGE_ACCESS_KEY",
		"VULTR_OBJECT_STORAGE_SECRET_KEY",
		"VULTR_RENDER_URL",
		"VULTR_RENDER_API_KEY",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected Vultr error to mention %q, got %q", expected, message)
		}
	}
	if strings.Contains(message, "AZURE_") {
		t.Fatalf("expected Vultr-only config error, got %q", message)
	}
}
