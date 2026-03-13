package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

func TestVultrAnalysisClientSubmitAndPoll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	videoPath := filepath.Join(t.TempDir(), "clip.mp4")
	if err := os.WriteFile(videoPath, []byte("video bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer vultr-analysis-key" {
			t.Fatalf("expected auth header, got %q", got)
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/analysis/jobs":
			writeTestJSON(t, w, map[string]any{"request_id": "analysis_1"})
		case r.Method == http.MethodGet && r.URL.Path == "/analysis/jobs/analysis_1":
			writeTestJSON(t, w, map[string]any{
				"request_id":  "analysis_1",
				"status":      "completed",
				"payload_ref": "analysis_payload_1",
				"scenes": []map[string]any{
					{
						"id":            "scene_1",
						"job_id":        "job_1",
						"scene_number":  1,
						"start_frame":   10,
						"end_frame":     20,
						"start_seconds": 0.4,
						"end_seconds":   0.8,
						"metadata":      map[string]any{"provider": "vultr"},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewVultrAnalysisClient(config.Config{
		VultrAnalysisURL:    server.URL,
		VultrAnalysisAPIKey: "vultr-analysis-key",
	}, logger, server.Client())

	submit, err := client.SubmitAnalysis(context.Background(), AnalysisRequest{
		JobID:      "job_1",
		CampaignID: "camp_1",
		ProductID:  "prod_1",
		VideoPath:  videoPath,
	})
	if err != nil {
		t.Fatalf("SubmitAnalysis() error = %v", err)
	}
	if submit.RequestID != "analysis_1" {
		t.Fatalf("unexpected request id %q", submit.RequestID)
	}

	poll, err := client.PollAnalysis(context.Background(), AnalysisPollRequest{
		JobID:     "job_1",
		RequestID: submit.RequestID,
	})
	if err != nil {
		t.Fatalf("PollAnalysis() error = %v", err)
	}
	if poll.Status != "completed" {
		t.Fatalf("expected completed, got %q", poll.Status)
	}
	if len(poll.Scenes) != 1 || poll.Scenes[0].ID != "scene_1" {
		t.Fatalf("unexpected scenes payload %#v", poll.Scenes)
	}
}

func TestVultrLLMClientComplete(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/llm/completions" {
			http.NotFound(w, r)
			return
		}
		writeTestJSON(t, w, map[string]any{
			"request_id": "llm_1",
			"content":    `{"top_slot_ids":["slot_1"]}`,
		})
	}))
	defer server.Close()

	client := NewVultrLLMClient(config.Config{
		VultrLLMURL:    server.URL,
		VultrLLMAPIKey: "llm-key",
	}, logger, server.Client())

	response, err := client.Complete(context.Background(), OpenAIRequest{JobID: "job_1", Purpose: "slot_ranking"})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if response.RequestID != "llm_1" || !strings.Contains(response.Content, "slot_1") {
		t.Fatalf("unexpected llm response %#v", response)
	}
}

func TestVultrGenerationClientSubmitAndPoll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/generations":
			writeTestJSON(t, w, map[string]any{"request_id": "gen_1", "status": "submitted"})
		case r.Method == http.MethodGet && r.URL.Path == "/generations/gen_1":
			writeTestJSON(t, w, map[string]any{
				"request_id":           "gen_1",
				"status":               "completed",
				"generated_clip_path":  "tmp/artifacts/job_1/slot_1.mp4",
				"generated_audio_path": "tmp/artifacts/job_1/slot_1.wav",
				"metadata": map[string]any{
					"duration_seconds": 6.0,
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewVultrGenerationClient(config.Config{
		VultrGenerationURL:    server.URL,
		VultrGenerationAPIKey: "generation-key",
	}, logger, server.Client())

	submit, err := client.SubmitGeneration(context.Background(), GenerationRequest{JobID: "job_1", SlotID: "slot_1"})
	if err != nil {
		t.Fatalf("SubmitGeneration() error = %v", err)
	}
	if submit.RequestID != "gen_1" {
		t.Fatalf("unexpected request id %q", submit.RequestID)
	}

	poll, err := client.PollGeneration(context.Background(), GenerationPollRequest{JobID: "job_1", SlotID: "slot_1", RequestID: submit.RequestID})
	if err != nil {
		t.Fatalf("PollGeneration() error = %v", err)
	}
	if poll.Status != "completed" || poll.GeneratedClipPath == "" || poll.GeneratedAudioPath == "" {
		t.Fatalf("unexpected generation poll response %#v", poll)
	}
}

func TestVultrRenderClientSubmitAndPoll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/renders":
			writeTestJSON(t, w, map[string]any{"request_id": "render_1", "status": "submitted"})
		case r.Method == http.MethodGet && r.URL.Path == "/renders/render_1":
			writeTestJSON(t, w, map[string]any{
				"request_id":       "render_1",
				"status":           "completed",
				"preview_blob_uri": "s3://hack-bucket/renders/job_1_preview.mp4",
				"duration_seconds": 906.0,
				"payload_ref":      "render_payload_1",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewVultrRenderClient(config.Config{
		VultrRenderURL:    server.URL,
		VultrRenderAPIKey: "render-key",
	}, logger, server.Client())

	submit, err := client.SubmitRender(context.Background(), RenderRequest{JobID: "job_1", SlotID: "slot_1"})
	if err != nil {
		t.Fatalf("SubmitRender() error = %v", err)
	}
	if submit.RequestID != "render_1" {
		t.Fatalf("unexpected request id %q", submit.RequestID)
	}

	poll, err := client.PollRender(context.Background(), RenderPollRequest{JobID: "job_1", SlotID: "slot_1", RequestID: submit.RequestID})
	if err != nil {
		t.Fatalf("PollRender() error = %v", err)
	}
	if poll.Status != "completed" || poll.PreviewBlobURI == "" {
		t.Fatalf("unexpected render poll response %#v", poll)
	}
}

func TestVultrObjectStorageClientUploadAndDownload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fakeAPI := &fakeVultrObjectStorageAPI{objects: map[string][]byte{}}
	client := newVultrObjectStorageClientWithAPI(config.Config{
		VultrObjectStorageEndpoint:  "https://ewr1.vultrobjects.com",
		VultrObjectStorageRegion:    "ewr1",
		VultrObjectStorageBucket:    "hack-bucket",
		VultrObjectStorageAccessKey: "access",
		VultrObjectStorageSecretKey: "secret",
	}, logger, fakeAPI)

	filePath := filepath.Join(t.TempDir(), "artifact.mp4")
	if err := os.WriteFile(filePath, []byte("mp4 bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	upload, err := client.Upload(context.Background(), BlobUploadRequest{
		JobID:      "job_1",
		Path:       filePath,
		ObjectName: "renders/job_1_preview.mp4",
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if upload.BlobURI != "s3://hack-bucket/renders/job_1_preview.mp4" {
		t.Fatalf("unexpected blob uri %q", upload.BlobURI)
	}

	download, err := client.Download(context.Background(), BlobDownloadRequest{
		JobID:   "job_1",
		BlobURI: upload.BlobURI,
	})
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer download.Body.Close()

	body, err := io.ReadAll(download.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(body) != "mp4 bytes" {
		t.Fatalf("unexpected object body %q", string(body))
	}
}

type fakeVultrObjectStorageAPI struct {
	objects map[string][]byte
}

func (f *fakeVultrObjectStorageAPI) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	key := aws.ToString(input.Bucket) + "/" + aws.ToString(input.Key)
	f.objects[key] = append([]byte(nil), body...)
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeVultrObjectStorageAPI) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	key := aws.ToString(input.Bucket) + "/" + aws.ToString(input.Key)
	data, ok := f.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found")
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(data))}, nil
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
}
