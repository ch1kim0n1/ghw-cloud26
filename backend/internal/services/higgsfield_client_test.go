package services

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type testBlobStorageClient struct {
	uploads []BlobUploadRequest
}

func (c *testBlobStorageClient) Upload(_ context.Context, req BlobUploadRequest) (BlobUploadResponse, error) {
	c.uploads = append(c.uploads, req)
	return BlobUploadResponse{
		RequestID: "blob_req",
		BlobURI:   "https://blob.example.com/" + strings.TrimLeft(req.ObjectName, "/"),
	}, nil
}

func (c *testBlobStorageClient) Download(context.Context, BlobDownloadRequest) (BlobDownloadResponse, error) {
	return BlobDownloadResponse{}, ErrPlaceholderClient
}

type testMLClient struct {
	submitResponse GenerationResponse
	submitErr      error
	pollResponse   GenerationResponse
	pollErr        error
	submitCalls    int
	pollCalls      int
}

func (c *testMLClient) SubmitGeneration(context.Context, GenerationRequest) (GenerationResponse, error) {
	c.submitCalls++
	if c.submitErr != nil {
		return GenerationResponse{}, c.submitErr
	}
	return c.submitResponse, nil
}

func (c *testMLClient) PollGeneration(context.Context, GenerationPollRequest) (GenerationResponse, error) {
	c.pollCalls++
	if c.pollErr != nil {
		return GenerationResponse{}, c.pollErr
	}
	return c.pollResponse, nil
}

func TestHiggsfieldClientPollDownloadsVideoWithoutAudio(t *testing.T) {
	tempDir := t.TempDir()
	startAnchor := filepath.Join(tempDir, "anchor.png")
	if err := os.WriteFile(startAnchor, []byte("png"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var submitPayload map[string]any
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/"+higgsfieldDefaultModelID:
			if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Key ") {
				t.Fatalf("expected Higgsfield auth header, got %q", got)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			if err := json.Unmarshal(body, &submitPayload); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"request_id": "hf_req_1",
				"status":     "submitted",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/requests/hf_req_1/status":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"request_id": "hf_req_1",
				"status":     "completed",
				"video": map[string]any{
					"url": server.URL + "/downloads/output.mp4",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/downloads/output.mp4":
			w.Header().Set("Content-Type", "video/mp4")
			_, _ = w.Write([]byte("mp4"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	blobClient := &testBlobStorageClient{}
	client := NewHiggsfieldClient(config.Config{
		HiggsfieldAPIKey:    "key",
		HiggsfieldAPISecret: "secret",
		HiggsfieldBaseURL:   server.URL,
		ArtifactsDir:        tempDir,
	}, logger, server.Client(), blobClient)

	submit, err := client.SubmitGeneration(context.Background(), GenerationRequest{
		JobID:                 "job_1",
		SlotID:                "slot_1",
		AnchorStartImagePath:  startAnchor,
		GenerationBrief:       "Bridge into the product interaction.",
		ProductName:           "Pepsi",
		TargetDurationSeconds: 6,
	})
	if err != nil {
		t.Fatalf("SubmitGeneration() error = %v", err)
	}
	if submit.RequestID != "higgsfield:hf_req_1" {
		t.Fatalf("expected prefixed request id, got %q", submit.RequestID)
	}
	if got := submitPayload["image_url"]; got == "" {
		t.Fatalf("expected image_url in Higgsfield payload, got %#v", submitPayload)
	}

	poll, err := client.PollGeneration(context.Background(), GenerationPollRequest{
		JobID:     "job_1",
		SlotID:    "slot_1",
		RequestID: submit.RequestID,
	})
	if err != nil {
		t.Fatalf("PollGeneration() error = %v", err)
	}
	if poll.Status != "completed" {
		t.Fatalf("expected completed status, got %q", poll.Status)
	}
	if poll.GeneratedClipPath == "" {
		t.Fatalf("expected generated clip path, got %#v", poll)
	}
	if poll.GeneratedAudioPath != "" {
		t.Fatalf("expected empty generated audio path, got %#v", poll)
	}
	if _, err := os.Stat(poll.GeneratedClipPath); err != nil {
		t.Fatalf("expected downloaded clip path to exist, got %v", err)
	}
	if len(blobClient.uploads) != 1 {
		t.Fatalf("expected one blob upload, got %d", len(blobClient.uploads))
	}
}

func TestPriorityFallbackMLClientFallsBackOnSubmitError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	primary := &testMLClient{submitErr: io.EOF}
	fallback := &testMLClient{
		submitResponse: GenerationResponse{
			RequestID: "az_req_1",
			Status:    "submitted",
		},
	}
	client := NewPriorityFallbackMLClient(GenerationProviderHiggsfield, primary, GenerationProviderAzureML, fallback, logger)

	response, err := client.SubmitGeneration(context.Background(), GenerationRequest{JobID: "job_1", SlotID: "slot_1"})
	if err != nil {
		t.Fatalf("SubmitGeneration() error = %v", err)
	}
	if primary.submitCalls != 1 || fallback.submitCalls != 1 {
		t.Fatalf("expected one primary and one fallback submit, got %d and %d", primary.submitCalls, fallback.submitCalls)
	}
	if response.RequestID != "azureml:az_req_1" {
		t.Fatalf("expected fallback-prefixed request id, got %q", response.RequestID)
	}
	if response.Metadata["generation_provider_used"] != GenerationProviderAzureML {
		t.Fatalf("expected azureml provider used, got %#v", response.Metadata["generation_provider_used"])
	}
	if response.Metadata["generation_fallback_used"] != true {
		t.Fatalf("expected fallback metadata, got %#v", response.Metadata["generation_fallback_used"])
	}
}

func TestPriorityFallbackMLClientPollRoutesByProviderPrefix(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	primary := &testMLClient{pollResponse: GenerationResponse{RequestID: "hf_req_1", Status: "pending"}}
	fallback := &testMLClient{pollResponse: GenerationResponse{RequestID: "az_req_1", Status: "pending"}}
	client := NewPriorityFallbackMLClient(GenerationProviderHiggsfield, primary, GenerationProviderAzureML, fallback, logger)

	if _, err := client.PollGeneration(context.Background(), GenerationPollRequest{JobID: "job_1", SlotID: "slot_1", RequestID: "higgsfield:hf_req_1"}); err != nil {
		t.Fatalf("PollGeneration() primary error = %v", err)
	}
	if _, err := client.PollGeneration(context.Background(), GenerationPollRequest{JobID: "job_1", SlotID: "slot_1", RequestID: "azureml:az_req_1"}); err != nil {
		t.Fatalf("PollGeneration() fallback error = %v", err)
	}
	if primary.pollCalls != 1 || fallback.pollCalls != 1 {
		t.Fatalf("expected one primary and one fallback poll, got %d and %d", primary.pollCalls, fallback.pollCalls)
	}
}

func TestGenerationOutputCacheKeyIncludesProvider(t *testing.T) {
	cache := NewProviderCache(t.TempDir())
	scene := models.Scene{ID: "scene_1", SceneNumber: 1, StartSeconds: 1, EndSeconds: 7}
	slot := models.Slot{AnchorStartFrame: 24, AnchorEndFrame: 144}

	higgsfieldKey := generationOutputCacheKey(cache, GenerationProviderHiggsfield, "video", "product", scene, slot, "en", "line")
	azureKey := generationOutputCacheKey(cache, GenerationProviderAzureML, "video", "product", scene, slot, "en", "line")
	if higgsfieldKey == azureKey {
		t.Fatal("expected provider-specific generation cache keys")
	}
}
