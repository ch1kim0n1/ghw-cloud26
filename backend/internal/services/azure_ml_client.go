package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type NoopMLClient struct {
	logger *slog.Logger
}

func NewNoopMLClient(logger *slog.Logger) *NoopMLClient {
	return &NoopMLClient{logger: logger}
}

func (c *NoopMLClient) SubmitGeneration(_ context.Context, req GenerationRequest) (GenerationResponse, error) {
	c.logger.Info("phase 0 ml placeholder invoked", "job_id", req.JobID, "slot_id", req.SlotID)
	return GenerationResponse{}, ErrPlaceholderClient
}

func (c *NoopMLClient) PollGeneration(_ context.Context, req GenerationPollRequest) (GenerationResponse, error) {
	c.logger.Info("phase 0 ml poll placeholder invoked", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", req.RequestID)
	return GenerationResponse{}, ErrPlaceholderClient
}

type AzureMLClient struct {
	baseURL    string
	logger     *slog.Logger
	httpClient *http.Client
}

func NewAzureMLClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client) *AzureMLClient {
	return &AzureMLClient{
		baseURL:    strings.TrimRight(cfg.AzureMLURL, "/"),
		logger:     logger,
		httpClient: httpClient,
	}
}

func (c *AzureMLClient) SubmitGeneration(ctx context.Context, req GenerationRequest) (GenerationResponse, error) {
	if c.baseURL == "" {
		return GenerationResponse{}, fmt.Errorf("Azure ML client is not configured")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("marshal Azure ML generation request: %w", err)
	}

	endpoint := c.baseURL + "/generations"
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("create Azure ML generation request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("submit Azure ML generation request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return GenerationResponse{}, err
	}

	generation := generationResponseFromPayload(payload)
	if generation.RequestID == "" {
		return GenerationResponse{}, fmt.Errorf("Azure ML generation response missing request id")
	}
	if generation.Status == "" {
		generation.Status = "submitted"
	}

	c.logger.Info("submitted Azure ML generation request", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", generation.RequestID)
	return generation, nil
}

func (c *AzureMLClient) PollGeneration(ctx context.Context, req GenerationPollRequest) (GenerationResponse, error) {
	if c.baseURL == "" {
		return GenerationResponse{}, fmt.Errorf("Azure ML client is not configured")
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return GenerationResponse{}, fmt.Errorf("Azure ML generation poll request is missing request id")
	}

	endpoint, err := url.Parse(c.baseURL + "/generations/" + url.PathEscape(req.RequestID))
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("build Azure ML generation poll URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("job_id", req.JobID)
	query.Set("slot_id", req.SlotID)
	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("create Azure ML generation poll request: %w", err)
	}

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("poll Azure ML generation request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return GenerationResponse{}, err
	}

	generation := generationResponseFromPayload(payload)
	if generation.RequestID == "" {
		generation.RequestID = req.RequestID
	}

	c.logger.Info("polled Azure ML generation request", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", generation.RequestID, "status", generation.Status)
	return generation, nil
}

func generationResponseFromPayload(payload map[string]any) GenerationResponse {
	metadata := mapValue(payload["metadata"])
	clipPath := firstNonEmptyString(
		stringValue(payload["generated_clip_path"]),
		stringValue(payload["clip_path"]),
		stringValue(metadata["generated_clip_path"]),
	)
	audioPath := firstNonEmptyString(
		stringValue(payload["generated_audio_path"]),
		stringValue(payload["audio_path"]),
		stringValue(metadata["generated_audio_path"]),
	)

	response := GenerationResponse{
		RequestID:          firstNonEmptyString(stringValue(payload["request_id"]), stringValue(payload["id"])),
		Status:             strings.ToLower(firstNonEmptyString(stringValue(payload["status"]), stringValue(metadata["status"]))),
		GeneratedClipPath:  clipPath,
		GeneratedAudioPath: audioPath,
		PayloadRef:         firstNonEmptyString(stringValue(payload["payload_ref"]), stringValue(payload["provider_payload_ref"]), stringValue(metadata["payload_ref"])),
		Message:            firstNonEmptyString(stringValue(payload["message"]), stringValue(metadata["message"])),
		Metadata:           models.Metadata{},
	}

	for key, value := range metadata {
		response.Metadata[key] = value
	}
	if response.GeneratedClipPath != "" {
		response.Metadata["generated_clip_path"] = response.GeneratedClipPath
	}
	if response.GeneratedAudioPath != "" {
		response.Metadata["generated_audio_path"] = response.GeneratedAudioPath
	}
	return response
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
