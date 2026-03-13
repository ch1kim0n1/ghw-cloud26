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

type NoopRenderClient struct {
	logger *slog.Logger
}

func NewNoopRenderClient(logger *slog.Logger) *NoopRenderClient {
	return &NoopRenderClient{logger: logger}
}

func (c *NoopRenderClient) SubmitRender(_ context.Context, req RenderRequest) (RenderResponse, error) {
	c.logger.Info("phase 0 render placeholder invoked", "job_id", req.JobID, "slot_id", req.SlotID)
	return RenderResponse{}, ErrPlaceholderClient
}

func (c *NoopRenderClient) PollRender(_ context.Context, req RenderPollRequest) (RenderResponse, error) {
	c.logger.Info("phase 0 render poll placeholder invoked", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", req.RequestID)
	return RenderResponse{}, ErrPlaceholderClient
}

type AzureRenderClient struct {
	baseURL    string
	apiKey     string
	logger     *slog.Logger
	httpClient *http.Client
}

func NewAzureRenderClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client) *AzureRenderClient {
	return &AzureRenderClient{
		baseURL:    strings.TrimRight(cfg.AzureRenderURL, "/"),
		apiKey:     strings.TrimSpace(cfg.AzureRenderAPIKey),
		logger:     logger,
		httpClient: httpClient,
	}
}

func (c *AzureRenderClient) SubmitRender(ctx context.Context, req RenderRequest) (RenderResponse, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return RenderResponse{}, fmt.Errorf("Azure Render client is not configured")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return RenderResponse{}, fmt.Errorf("marshal render request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/renders", bytes.NewReader(body))
	if err != nil {
		return RenderResponse{}, fmt.Errorf("create render request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return RenderResponse{}, fmt.Errorf("submit render request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return RenderResponse{}, err
	}

	render := renderResponseFromPayload(payload)
	if render.RequestID == "" {
		return RenderResponse{}, fmt.Errorf("render response missing request id")
	}
	if render.Status == "" {
		render.Status = "submitted"
	}

	c.logger.Info("submitted preview render request", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", render.RequestID)
	return render, nil
}

func (c *AzureRenderClient) PollRender(ctx context.Context, req RenderPollRequest) (RenderResponse, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return RenderResponse{}, fmt.Errorf("Azure Render client is not configured")
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return RenderResponse{}, fmt.Errorf("render poll request is missing request id")
	}

	endpoint, err := url.Parse(c.baseURL + "/renders/" + url.PathEscape(req.RequestID))
	if err != nil {
		return RenderResponse{}, fmt.Errorf("build render poll URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("job_id", req.JobID)
	query.Set("slot_id", req.SlotID)
	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return RenderResponse{}, fmt.Errorf("create render poll request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return RenderResponse{}, fmt.Errorf("poll render request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return RenderResponse{}, err
	}

	render := renderResponseFromPayload(payload)
	if render.RequestID == "" {
		render.RequestID = req.RequestID
	}

	c.logger.Info("polled preview render request", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", render.RequestID, "status", render.Status)
	return render, nil
}

func renderResponseFromPayload(payload map[string]any) RenderResponse {
	metadata := mapValue(payload["metadata"])
	response := RenderResponse{
		RequestID:       firstNonEmptyString(stringValue(payload["request_id"]), stringValue(payload["id"])),
		Status:          strings.ToLower(firstNonEmptyString(stringValue(payload["status"]), stringValue(metadata["status"]))),
		PreviewBlobURI:  firstNonEmptyString(stringValue(payload["preview_blob_uri"]), stringValue(payload["output_blob_uri"]), stringValue(metadata["preview_blob_uri"])),
		DurationSeconds: floatValueFromAny(firstNonEmptyValue(payload["duration_seconds"], metadata["duration_seconds"])),
		PayloadRef:      firstNonEmptyString(stringValue(payload["payload_ref"]), stringValue(metadata["payload_ref"])),
		Message:         firstNonEmptyString(stringValue(payload["message"]), stringValue(metadata["message"])),
		Metadata:        models.Metadata{},
	}
	for key, value := range metadata {
		response.Metadata[key] = value
	}
	if response.PreviewBlobURI != "" {
		response.Metadata["preview_blob_uri"] = response.PreviewBlobURI
	}
	return response
}

func firstNonEmptyValue(values ...any) any {
	for _, value := range values {
		if stringValue(value) != "" {
			return value
		}
	}
	return nil
}

func floatValueFromAny(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		parsed, _ := typed.Float64()
		return parsed
	case string:
		if typed == "" {
			return 0
		}
		var parsed float64
		_, _ = fmt.Sscanf(typed, "%f", &parsed)
		return parsed
	default:
		return 0
	}
}
