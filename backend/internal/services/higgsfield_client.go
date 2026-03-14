package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

const (
	GenerationProviderHiggsfield = "higgsfield"
	GenerationProviderAzureML    = "azureml"
	GenerationProviderVultr      = "vultr"

	higgsfieldDefaultModelID = "kling-video/v2.1/pro/image-to-video"
)

type FallbackGenerationSubmitter interface {
	SubmitGenerationFallback(context.Context, GenerationRequest, string) (GenerationResponse, error)
}

type HiggsfieldClient struct {
	baseURL      string
	apiKey       string
	apiSecret    string
	modelID      string
	artifactsDir string
	blobClient   BlobStorageClient
	logger       *slog.Logger
	httpClient   *http.Client
}

func NewHiggsfieldClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client, blobClient BlobStorageClient) *HiggsfieldClient {
	return &HiggsfieldClient{
		baseURL:      strings.TrimRight(cfg.HiggsfieldBaseURL, "/"),
		apiKey:       strings.TrimSpace(cfg.HiggsfieldAPIKey),
		apiSecret:    strings.TrimSpace(cfg.HiggsfieldAPISecret),
		modelID:      higgsfieldDefaultModelID,
		artifactsDir: cfg.ArtifactsDir,
		blobClient:   blobClient,
		logger:       logger,
		httpClient:   httpClient,
	}
}

func (c *HiggsfieldClient) SubmitGeneration(ctx context.Context, req GenerationRequest) (GenerationResponse, error) {
	if err := c.validate(); err != nil {
		return GenerationResponse{}, err
	}

	imageURL, err := c.uploadAnchorImage(ctx, req)
	if err != nil {
		return GenerationResponse{}, err
	}

	body, err := json.Marshal(map[string]any{
		"image_url": imageURL,
		"prompt":    buildHiggsfieldPrompt(req),
		"duration":  higgsfieldDuration(req.TargetDurationSeconds),
	})
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("marshal Higgsfield generation request: %w", err)
	}

	endpoint := strings.TrimRight(c.baseURL, "/") + "/" + c.modelID
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("create Higgsfield generation request: %w", err)
	}
	httpRequest.Header.Set("Authorization", c.authorizationHeader())
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("submit Higgsfield generation request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return GenerationResponse{}, err
	}

	generation, err := c.responseFromPayload(ctx, req, payload, "")
	if err != nil {
		return GenerationResponse{}, err
	}
	if generation.RequestID == "" && generation.Status != "completed" {
		return GenerationResponse{}, fmt.Errorf("Higgsfield generation response missing request id")
	}
	c.logger.Info("submitted Higgsfield generation request", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", generation.RequestID)
	return generation, nil
}

func (c *HiggsfieldClient) PollGeneration(ctx context.Context, req GenerationPollRequest) (GenerationResponse, error) {
	if err := c.validate(); err != nil {
		return GenerationResponse{}, err
	}

	requestID := decodeProviderRequestID(req.RequestID, GenerationProviderHiggsfield)
	if requestID == "" {
		requestID = strings.TrimSpace(req.RequestID)
	}
	if requestID == "" {
		return GenerationResponse{}, fmt.Errorf("Higgsfield generation poll request is missing request id")
	}

	endpoint := strings.TrimRight(c.baseURL, "/") + "/requests/" + url.PathEscape(requestID) + "/status"
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("create Higgsfield generation poll request: %w", err)
	}
	httpRequest.Header.Set("Authorization", c.authorizationHeader())
	httpRequest.Header.Set("Accept", "application/json")

	var payload map[string]any
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("poll Higgsfield generation request: %w", err)
	}
	defer response.Body.Close()
	if err := decodeJSONResponse(response, &payload); err != nil {
		return GenerationResponse{}, err
	}

	generation, err := c.responseFromPayload(ctx, GenerationRequest{JobID: req.JobID, SlotID: req.SlotID}, payload, requestID)
	if err != nil {
		return GenerationResponse{}, err
	}
	if generation.RequestID == "" {
		generation.RequestID = requestID
	}
	c.logger.Info("polled Higgsfield generation request", "job_id", req.JobID, "slot_id", req.SlotID, "request_id", generation.RequestID, "status", generation.Status)
	return generation, nil
}

func (c *HiggsfieldClient) validate() error {
	if c.baseURL == "" || c.apiKey == "" || c.apiSecret == "" {
		return fmt.Errorf("Higgsfield client is not configured")
	}
	if c.blobClient == nil {
		return fmt.Errorf("Higgsfield client is missing blob storage")
	}
	if c.artifactsDir == "" {
		return fmt.Errorf("Higgsfield client is missing artifacts directory")
	}
	return nil
}

func (c *HiggsfieldClient) authorizationHeader() string {
	return fmt.Sprintf("Key %s:%s", c.apiKey, c.apiSecret)
}

func (c *HiggsfieldClient) uploadAnchorImage(ctx context.Context, req GenerationRequest) (string, error) {
	if strings.TrimSpace(req.AnchorStartImagePath) == "" {
		return "", fmt.Errorf("Higgsfield generation requires an anchor start image")
	}
	objectName := fmt.Sprintf("%s/higgsfield/%s/anchor_start%s", req.JobID, req.SlotID, strings.ToLower(filepath.Ext(req.AnchorStartImagePath)))
	upload, err := c.blobClient.Upload(ctx, BlobUploadRequest{
		JobID:      req.JobID,
		Path:       req.AnchorStartImagePath,
		ObjectName: objectName,
	})
	if err != nil {
		return "", fmt.Errorf("upload Higgsfield conditioning image: %w", err)
	}
	if strings.TrimSpace(upload.BlobURI) == "" {
		return "", fmt.Errorf("Higgsfield conditioning image upload returned empty blob uri")
	}
	return upload.BlobURI, nil
}

func (c *HiggsfieldClient) responseFromPayload(ctx context.Context, req GenerationRequest, payload map[string]any, fallbackRequestID string) (GenerationResponse, error) {
	metadata := cloneMetadata(models.Metadata(mapValue(payload["metadata"])))
	if metadata == nil {
		metadata = models.Metadata{}
	}

	status := normalizeHiggsfieldStatus(firstNonEmptyString(
		stringValue(payload["status"]),
		stringValue(payload["state"]),
		stringValue(metadata["status"]),
	))
	requestID := firstNonEmptyString(
		stringValue(payload["request_id"]),
		stringValue(payload["id"]),
		stringValue(payload["uuid"]),
		stringValue(metadata["request_id"]),
		fallbackRequestID,
	)

	videoURL := extractHiggsfieldAssetURL(payload, metadata, "video")
	audioURL := extractHiggsfieldAssetURL(payload, metadata, "audio")
	generation := GenerationResponse{
		RequestID:  encodeProviderRequestID(GenerationProviderHiggsfield, requestID),
		Status:     status,
		PayloadRef: firstNonEmptyString(stringValue(payload["payload_ref"]), stringValue(metadata["payload_ref"])),
		Message:    firstNonEmptyString(stringValue(payload["message"]), stringValue(metadata["message"])),
		Metadata:   metadata,
	}

	if status == "completed" {
		if strings.TrimSpace(videoURL) == "" {
			return GenerationResponse{}, fmt.Errorf("Higgsfield generation completed without a downloadable video")
		}
		clipPath, err := c.downloadArtifact(ctx, req.JobID, req.SlotID, videoURL, "generated_clip", ".mp4")
		if err != nil {
			return GenerationResponse{}, fmt.Errorf("download Higgsfield generated clip: %w", err)
		}
		generation.GeneratedClipPath = clipPath
		if strings.TrimSpace(audioURL) != "" {
			audioPath, err := c.downloadArtifact(ctx, req.JobID, req.SlotID, audioURL, "generated_audio", path.Ext(audioURL))
			if err != nil {
				return GenerationResponse{}, fmt.Errorf("download Higgsfield generated audio: %w", err)
			}
			generation.GeneratedAudioPath = audioPath
		}
	}

	generation.Metadata = annotateGenerationMetadata(generation.Metadata, GenerationProviderHiggsfield, false, "")
	generation.Metadata["generation_provider_attempted"] = GenerationProviderHiggsfield
	if requestID != "" {
		generation.Metadata["provider_request_id"] = requestID
	}
	if videoURL != "" {
		generation.Metadata["higgsfield_video_url"] = videoURL
	}
	if audioURL != "" {
		generation.Metadata["higgsfield_audio_url"] = audioURL
	}
	generation.Metadata["higgsfield_model_id"] = c.modelID
	if generation.GeneratedClipPath != "" {
		generation.Metadata["generated_clip_path"] = generation.GeneratedClipPath
	}
	if generation.GeneratedAudioPath != "" {
		generation.Metadata["generated_audio_path"] = generation.GeneratedAudioPath
	}
	return generation, nil
}

func (c *HiggsfieldClient) downloadArtifact(ctx context.Context, jobID, slotID, rawURL, baseName, fallbackExt string) (string, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("create Higgsfield artifact download request: %w", err)
	}
	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return "", fmt.Errorf("download Higgsfield artifact: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("Higgsfield artifact download failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := os.MkdirAll(filepath.Join(c.artifactsDir, jobID), 0o755); err != nil {
		return "", fmt.Errorf("create Higgsfield artifact directory: %w", err)
	}

	parsedURL, _ := url.Parse(rawURL)
	extension := strings.ToLower(path.Ext(parsedURL.Path))
	if extension == "" {
		extension = strings.ToLower(strings.TrimSpace(fallbackExt))
	}
	if extension == "" {
		extension = ".bin"
	}
	targetPath := filepath.Join(c.artifactsDir, jobID, fmt.Sprintf("%s_%s%s", slotID, baseName, extension))
	file, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("create Higgsfield artifact file: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, response.Body); err != nil {
		return "", fmt.Errorf("write Higgsfield artifact: %w", err)
	}
	return targetPath, nil
}

type PriorityFallbackMLClient struct {
	primary      MLClient
	fallback     MLClient
	primaryName  string
	fallbackName string
	logger       *slog.Logger
}

func NewPriorityFallbackMLClient(primaryName string, primary MLClient, fallbackName string, fallback MLClient, logger *slog.Logger) *PriorityFallbackMLClient {
	return &PriorityFallbackMLClient{
		primary:      primary,
		fallback:     fallback,
		primaryName:  strings.TrimSpace(primaryName),
		fallbackName: strings.TrimSpace(fallbackName),
		logger:       logger,
	}
}

func (c *PriorityFallbackMLClient) SubmitGeneration(ctx context.Context, req GenerationRequest) (GenerationResponse, error) {
	response, err := c.primary.SubmitGeneration(ctx, req)
	if err == nil && isUsableGenerationSubmit(response) {
		return c.decorateResponse(response, c.primaryName, false, ""), nil
	}

	fallbackReason := summarizeFallbackReason(err, response)
	if c.logger != nil {
		c.logger.Warn("falling back after primary generation submission failure", "job_id", req.JobID, "slot_id", req.SlotID, "primary_provider", c.primaryName, "fallback_provider", c.fallbackName, "reason", fallbackReason)
	}
	return c.SubmitGenerationFallback(ctx, req, fallbackReason)
}

func (c *PriorityFallbackMLClient) SubmitGenerationFallback(ctx context.Context, req GenerationRequest, reason string) (GenerationResponse, error) {
	response, err := c.fallback.SubmitGeneration(ctx, req)
	if err != nil {
		return GenerationResponse{}, fmt.Errorf("%s failed: %s; %s failed: %w", c.primaryName, strings.TrimSpace(reason), c.fallbackName, err)
	}
	if !isUsableGenerationSubmit(response) {
		return GenerationResponse{}, fmt.Errorf("%s failed: %s; %s returned unusable generation response", c.primaryName, strings.TrimSpace(reason), c.fallbackName)
	}
	return c.decorateResponse(response, c.fallbackName, true, reason), nil
}

func (c *PriorityFallbackMLClient) PollGeneration(ctx context.Context, req GenerationPollRequest) (GenerationResponse, error) {
	providerName, rawRequestID := splitProviderRequestID(req.RequestID)
	if rawRequestID == "" {
		rawRequestID = req.RequestID
	}
	pollReq := req
	pollReq.RequestID = rawRequestID

	switch providerName {
	case "", c.fallbackName:
		response, err := c.fallback.PollGeneration(ctx, pollReq)
		if err != nil {
			return GenerationResponse{}, err
		}
		return c.decorateResponse(response, c.fallbackName, providerName == "", ""), nil
	case c.primaryName:
		response, err := c.primary.PollGeneration(ctx, pollReq)
		if err != nil {
			return GenerationResponse{}, err
		}
		return c.decorateResponse(response, c.primaryName, false, ""), nil
	default:
		return GenerationResponse{}, fmt.Errorf("unsupported generation provider %q", providerName)
	}
}

func (c *PriorityFallbackMLClient) decorateResponse(response GenerationResponse, providerName string, fallbackUsed bool, fallbackReason string) GenerationResponse {
	if strings.TrimSpace(providerName) == "" {
		return response
	}
	metadata := annotateGenerationMetadata(response.Metadata, providerName, fallbackUsed, fallbackReason)
	if providerName == c.primaryName {
		metadata["generation_provider_attempted"] = c.primaryName
	} else {
		metadata["generation_provider_attempted"] = c.primaryName
	}
	if rawRequestID := decodeProviderRequestID(response.RequestID, providerName); rawRequestID != "" {
		metadata["provider_request_id"] = rawRequestID
	} else if strings.TrimSpace(response.RequestID) != "" && !strings.Contains(response.RequestID, ":") {
		metadata["provider_request_id"] = strings.TrimSpace(response.RequestID)
	}
	response.Metadata = metadata
	if strings.TrimSpace(response.RequestID) != "" {
		response.RequestID = encodeProviderRequestID(providerName, decodeProviderRequestID(response.RequestID, providerName))
	}
	return response
}

func annotateGenerationMetadata(metadata models.Metadata, providerName string, fallbackUsed bool, fallbackReason string) models.Metadata {
	cloned := cloneMetadata(metadata)
	if cloned == nil {
		cloned = models.Metadata{}
	}
	cloned["generation_provider_used"] = providerName
	cloned["generation_fallback_used"] = fallbackUsed
	if strings.TrimSpace(fallbackReason) != "" {
		cloned["generation_fallback_reason"] = strings.TrimSpace(fallbackReason)
	}
	return cloned
}

func buildHiggsfieldPrompt(req GenerationRequest) string {
	parts := []string{
		strings.TrimSpace(req.GenerationBrief),
		fmt.Sprintf("The clip must begin from the provided start anchor and feel visually continuous with the original source scene \"%s\".", strings.TrimSpace(req.SceneNarrativeSummary)),
		"Keep motion realistic, safe, and physically plausible.",
		fmt.Sprintf("Feature the product \"%s\" naturally.", strings.TrimSpace(req.ProductName)),
	}
	if line := strings.TrimSpace(firstNonEmptyString(req.FinalProductLine, req.SuggestedProductLine)); line != "" {
		parts = append(parts, fmt.Sprintf("If spoken dialogue is represented visually, align it with this line: %s.", line))
	}
	if description := strings.TrimSpace(req.ProductDescription); description != "" {
		parts = append(parts, fmt.Sprintf("Product details: %s.", description))
	}
	if category := strings.TrimSpace(req.ProductCategory); category != "" {
		parts = append(parts, fmt.Sprintf("Product category: %s.", category))
	}
	if language := strings.TrimSpace(req.ContentLanguage); language != "" {
		parts = append(parts, fmt.Sprintf("Match the source content language: %s.", language))
	}
	if req.SelectedSlotQuietWindow > 0 {
		parts = append(parts, fmt.Sprintf("Respect the insertion window and keep the interaction concise enough for a %.2f second slot.", req.SelectedSlotQuietWindow))
	}
	return strings.Join(parts, " ")
}

func higgsfieldDuration(targetSeconds int) int {
	if targetSeconds <= 0 {
		return 5
	}
	if targetSeconds <= 7 {
		return 5
	}
	if targetSeconds > 10 {
		return 10
	}
	return 10
}

func normalizeHiggsfieldStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "complete", "completed", "succeeded", "success":
		return "completed"
	case "failed", "error", "cancelled", "canceled":
		return "failed"
	case "running", "processing", "queued", "pending", "submitted":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func extractHiggsfieldAssetURL(payload map[string]any, metadata models.Metadata, assetType string) string {
	metadataMap := map[string]any(metadata)
	candidates := []any{
		nestedValue(payload, assetType, "url"),
		nestedValue(payload, "result", assetType, "url"),
		nestedValue(payload, "output", assetType, "url"),
		nestedValue(payload, "data", assetType, "url"),
		nestedValue(metadataMap, assetType, "url"),
		payload[assetType+"_url"],
		metadata[assetType+"_url"],
	}
	for _, candidate := range candidates {
		if value := strings.TrimSpace(stringValue(candidate)); value != "" {
			return value
		}
	}
	return ""
}

func isUsableGenerationSubmit(response GenerationResponse) bool {
	status := strings.ToLower(strings.TrimSpace(response.Status))
	switch status {
	case "completed", "succeeded":
		return strings.TrimSpace(response.GeneratedClipPath) != ""
	case "failed", "error", "cancelled", "canceled":
		return false
	}
	return strings.TrimSpace(response.RequestID) != ""
}

func summarizeFallbackReason(err error, response GenerationResponse) string {
	if err != nil {
		return err.Error()
	}
	if message := strings.TrimSpace(response.Message); message != "" {
		return message
	}
	status := strings.TrimSpace(response.Status)
	if status == "" {
		return "primary provider returned an unusable response"
	}
	return fmt.Sprintf("primary provider returned status %s", status)
}

func encodeProviderRequestID(providerName, requestID string) string {
	trimmedProvider := strings.TrimSpace(providerName)
	trimmedRequest := strings.TrimSpace(requestID)
	if trimmedProvider == "" || trimmedRequest == "" {
		return trimmedRequest
	}
	if strings.HasPrefix(trimmedRequest, trimmedProvider+":") {
		return trimmedRequest
	}
	return trimmedProvider + ":" + trimmedRequest
}

func decodeProviderRequestID(requestID, providerName string) string {
	trimmed := strings.TrimSpace(requestID)
	prefix := strings.TrimSpace(providerName) + ":"
	if strings.HasPrefix(trimmed, prefix) {
		return strings.TrimPrefix(trimmed, prefix)
	}
	return trimmed
}

func splitProviderRequestID(requestID string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(requestID), ":", 2)
	if len(parts) != 2 {
		return "", strings.TrimSpace(requestID)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
