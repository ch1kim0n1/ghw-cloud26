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
				_, err := services.NewNoopMLClient(logger).SubmitGeneration(context.Background(), services.GenerationRequest{JobID: "job_1"})
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
				_, err := services.NewNoopBlobStorageClient(logger).Upload(context.Background(), services.BlobUploadRequest{JobID: "job_1"})
				return err
			},
		},
		{
			name: "render",
			run: func() error {
				_, err := services.NewNoopRenderClient(logger).SubmitRender(context.Background(), services.RenderRequest{JobID: "job_1"})
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
