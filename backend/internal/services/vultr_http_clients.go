package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type VultrAnalysisClient struct {
	baseURL    string
	apiKey     string
	logger     *slog.Logger
	httpClient *http.Client
}

func NewVultrAnalysisClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client) *VultrAnalysisClient {
	return &VultrAnalysisClient{
		baseURL:    strings.TrimRight(cfg.VultrAnalysisURL, "/"),
		apiKey:     strings.TrimSpace(cfg.VultrAnalysisAPIKey),
		logger:     logger,
		httpClient: httpClient,
	}
}

func (c *VultrAnalysisClient) SubmitAnalysis(ctx context.Context, req AnalysisRequest) (AnalysisResponse, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return AnalysisResponse{}, fmt.Errorf("Vultr analysis client is not configured")
	}

	file, err := os.Open(req.VideoPath)
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("open source video: %w", err)
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(req.VideoPath))
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("create multipart file part: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return AnalysisResponse{}, fmt.Errorf("copy source video into request: %w", err)
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "job_id", value: req.JobID},
		{name: "campaign_id", value: req.CampaignID},
		{name: "product_id", value: req.ProductID},
	} {
		if strings.TrimSpace(field.value) == "" {
			continue
		}
		if err := writer.WriteField(field.name, field.value); err != nil {
			return AnalysisResponse{}, fmt.Errorf("write Vultr analysis form field %s: %w", field.name, err)
		}
	}
	if err := writer.Close(); err != nil {
		return AnalysisResponse{}, fmt.Errorf("close multipart writer: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/analysis/jobs", &body)
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("create Vultr analysis request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpRequest.Header.Set("Content-Type", writer.FormDataContentType())

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("submit Vultr analysis request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return AnalysisResponse{}, err
	}

	requestID := firstNonEmptyString(stringValue(payload["request_id"]), stringValue(payload["id"]))
	if requestID == "" {
		return AnalysisResponse{}, fmt.Errorf("Vultr analysis response missing request id")
	}

	c.logger.Info("submitted Vultr analysis request", "job_id", req.JobID, "request_id", requestID)
	return AnalysisResponse{RequestID: requestID}, nil
}

func (c *VultrAnalysisClient) PollAnalysis(ctx context.Context, req AnalysisPollRequest) (AnalysisPollResponse, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return AnalysisPollResponse{}, fmt.Errorf("Vultr analysis client is not configured")
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return AnalysisPollResponse{}, fmt.Errorf("Vultr analysis poll request is missing request id")
	}

	endpoint, err := url.Parse(c.baseURL + "/analysis/jobs/" + url.PathEscape(req.RequestID))
	if err != nil {
		return AnalysisPollResponse{}, fmt.Errorf("build Vultr analysis poll URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("job_id", req.JobID)
	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return AnalysisPollResponse{}, fmt.Errorf("create Vultr analysis poll request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return AnalysisPollResponse{}, fmt.Errorf("poll Vultr analysis request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return AnalysisPollResponse{}, err
	}

	status := strings.ToLower(firstNonEmptyString(stringValue(payload["status"]), stringValue(mapValue(payload["metadata"])["status"])))
	if status == "" {
		status = "pending"
	}

	result := AnalysisPollResponse{
		RequestID:  firstNonEmptyString(stringValue(payload["request_id"]), stringValue(payload["id"]), req.RequestID),
		Status:     status,
		PayloadRef: firstNonEmptyString(stringValue(payload["payload_ref"]), req.RequestID),
		Message:    firstNonEmptyString(stringValue(payload["message"]), stringValue(mapValue(payload["metadata"])["message"])),
	}
	if result.Status == "completed" || result.Status == "succeeded" {
		scenes, err := vultrScenesFromPayload(payload)
		if err != nil {
			return AnalysisPollResponse{}, err
		}
		result.Status = "completed"
		result.Scenes = scenes
		if result.PayloadRef == "" {
			result.PayloadRef = result.RequestID
		}
	}
	return result, nil
}

type VultrLLMClient struct {
	baseURL    string
	apiKey     string
	logger     *slog.Logger
	httpClient *http.Client
}

func NewVultrLLMClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client) *VultrLLMClient {
	return &VultrLLMClient{
		baseURL:    strings.TrimRight(cfg.VultrLLMURL, "/"),
		apiKey:     strings.TrimSpace(cfg.VultrLLMAPIKey),
		logger:     logger,
		httpClient: httpClient,
	}
}

func (c *VultrLLMClient) Complete(ctx context.Context, req OpenAIRequest) (OpenAIResponse, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return OpenAIResponse{}, fmt.Errorf("Vultr LLM client is not configured")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("marshal Vultr LLM request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/llm/completions", bytes.NewReader(body))
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("create Vultr LLM request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("submit Vultr LLM request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return OpenAIResponse{}, err
	}

	content := firstNonEmptyString(stringValue(payload["content"]), stringValue(mapValue(payload["message"])["content"]))
	if content == "" {
		return OpenAIResponse{}, fmt.Errorf("Vultr LLM response missing content")
	}

	requestID := firstNonEmptyString(stringValue(payload["request_id"]), stringValue(payload["id"]))
	c.logger.Info("received Vultr LLM completion", "job_id", req.JobID, "purpose", req.Purpose, "request_id", requestID)
	return OpenAIResponse{RequestID: requestID, Content: content}, nil
}

type VultrGenerationClient struct {
	baseURL    string
	apiKey     string
	logger     *slog.Logger
	httpClient *http.Client
}

func NewVultrGenerationClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client) *VultrGenerationClient {
	return &VultrGenerationClient{
		baseURL:    strings.TrimRight(cfg.VultrGenerationURL, "/"),
		apiKey:     strings.TrimSpace(cfg.VultrGenerationAPIKey),
		logger:     logger,
		httpClient: httpClient,
	}
}

func (c *VultrGenerationClient) SubmitGeneration(ctx context.Context, req GenerationRequest) (GenerationResponse, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return GenerationResponse{}, fmt.Errorf("Vultr generation client is not configured")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("marshal Vultr generation request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/generations", bytes.NewReader(body))
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("create Vultr generation request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("submit Vultr generation request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return GenerationResponse{}, err
	}

	generation := generationResponseFromPayload(payload)
	if generation.RequestID == "" {
		return GenerationResponse{}, fmt.Errorf("Vultr generation response missing request id")
	}
	if generation.Status == "" {
		generation.Status = "submitted"
	}
	c.logger.Info("submitted Vultr generation request", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", generation.RequestID)
	return generation, nil
}

func (c *VultrGenerationClient) PollGeneration(ctx context.Context, req GenerationPollRequest) (GenerationResponse, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return GenerationResponse{}, fmt.Errorf("Vultr generation client is not configured")
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return GenerationResponse{}, fmt.Errorf("Vultr generation poll request is missing request id")
	}

	endpoint, err := url.Parse(c.baseURL + "/generations/" + url.PathEscape(req.RequestID))
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("build Vultr generation poll URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("job_id", req.JobID)
	query.Set("slot_id", req.SlotID)
	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("create Vultr generation poll request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("poll Vultr generation request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return GenerationResponse{}, err
	}

	generation := generationResponseFromPayload(payload)
	if generation.RequestID == "" {
		generation.RequestID = req.RequestID
	}
	c.logger.Info("polled Vultr generation request", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", generation.RequestID, "status", generation.Status)
	return generation, nil
}

type VultrRenderClient struct {
	baseURL    string
	apiKey     string
	logger     *slog.Logger
	httpClient *http.Client
}

func NewVultrRenderClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client) *VultrRenderClient {
	return &VultrRenderClient{
		baseURL:    strings.TrimRight(cfg.VultrRenderURL, "/"),
		apiKey:     strings.TrimSpace(cfg.VultrRenderAPIKey),
		logger:     logger,
		httpClient: httpClient,
	}
}

func (c *VultrRenderClient) SubmitRender(ctx context.Context, req RenderRequest) (RenderResponse, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return RenderResponse{}, fmt.Errorf("Vultr render client is not configured")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return RenderResponse{}, fmt.Errorf("marshal Vultr render request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/renders", bytes.NewReader(body))
	if err != nil {
		return RenderResponse{}, fmt.Errorf("create Vultr render request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return RenderResponse{}, fmt.Errorf("submit Vultr render request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return RenderResponse{}, err
	}

	render := renderResponseFromPayload(payload)
	if render.RequestID == "" {
		return RenderResponse{}, fmt.Errorf("Vultr render response missing request id")
	}
	if render.Status == "" {
		render.Status = "submitted"
	}
	c.logger.Info("submitted Vultr render request", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", render.RequestID)
	return render, nil
}

func (c *VultrRenderClient) PollRender(ctx context.Context, req RenderPollRequest) (RenderResponse, error) {
	if c.baseURL == "" || c.apiKey == "" {
		return RenderResponse{}, fmt.Errorf("Vultr render client is not configured")
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return RenderResponse{}, fmt.Errorf("Vultr render poll request is missing request id")
	}

	endpoint, err := url.Parse(c.baseURL + "/renders/" + url.PathEscape(req.RequestID))
	if err != nil {
		return RenderResponse{}, fmt.Errorf("build Vultr render poll URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("job_id", req.JobID)
	query.Set("slot_id", req.SlotID)
	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return RenderResponse{}, fmt.Errorf("create Vultr render poll request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return RenderResponse{}, fmt.Errorf("poll Vultr render request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return RenderResponse{}, err
	}

	render := renderResponseFromPayload(payload)
	if render.RequestID == "" {
		render.RequestID = req.RequestID
	}
	c.logger.Info("polled Vultr render request", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", render.RequestID, "status", render.Status)
	return render, nil
}

func vultrScenesFromPayload(payload map[string]any) ([]models.Scene, error) {
	scenesValue := payload["scenes"]
	if scenesValue == nil {
		scenesValue = mapValue(payload["result"])["scenes"]
	}
	if scenesValue == nil {
		return []models.Scene{}, nil
	}

	body, err := json.Marshal(scenesValue)
	if err != nil {
		return nil, fmt.Errorf("marshal Vultr scenes payload: %w", err)
	}

	var scenes []models.Scene
	if err := json.Unmarshal(body, &scenes); err != nil {
		return nil, fmt.Errorf("decode Vultr scenes payload: %w", err)
	}
	return scenes, nil
}
