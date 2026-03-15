package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

type WebsiteAdsImageGenerateRequest struct {
	Prompt string
	Width  int
	Height int
}

type WebsiteAdsImageGenerateResponse struct {
	Image       []byte
	ContentType string
}

type WebsiteAdsImageClient interface {
	Generate(context.Context, WebsiteAdsImageGenerateRequest) (WebsiteAdsImageGenerateResponse, error)
}

type huggingFaceWebsiteAdsImageClient struct {
	baseURL    string
	model      string
	token      string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewWebsiteAdsImageClient(cfg config.Config, logger *slog.Logger) WebsiteAdsImageClient {
	return &huggingFaceWebsiteAdsImageClient{
		baseURL: strings.TrimRight(cfg.HuggingFaceBaseURL, "/"),
		model:   strings.TrimSpace(cfg.HuggingFaceImageModel),
		token:   strings.TrimSpace(cfg.HuggingFaceAPIToken),
		httpClient: &http.Client{
			Timeout: cfg.HuggingFaceRequestTimeout,
		},
		logger: logger,
	}
}

func (c *huggingFaceWebsiteAdsImageClient) Generate(ctx context.Context, input WebsiteAdsImageGenerateRequest) (WebsiteAdsImageGenerateResponse, error) {
	if c.token == "" {
		return WebsiteAdsImageGenerateResponse{}, fmt.Errorf("missing HUGGINGFACE_API_TOKEN")
	}
	if c.model == "" {
		return WebsiteAdsImageGenerateResponse{}, fmt.Errorf("missing HUGGINGFACE_IMAGE_MODEL")
	}

	payload := map[string]any{
		"inputs": input.Prompt,
		"parameters": map[string]any{
			"width":  input.Width,
			"height": input.Height,
		},
		"options": map[string]any{
			"wait_for_model": true,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return WebsiteAdsImageGenerateResponse{}, fmt.Errorf("marshal hugging face request: %w", err)
	}

	url := c.baseURL + "/" + path.Clean(c.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return WebsiteAdsImageGenerateResponse{}, fmt.Errorf("build hugging face request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "image/png")

	startedAt := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return WebsiteAdsImageGenerateResponse{}, fmt.Errorf("call hugging face image endpoint: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return WebsiteAdsImageGenerateResponse{}, fmt.Errorf("read hugging face image response: %w", err)
	}

	c.logger.Info("website ads image generation completed", "model", c.model, "status", resp.StatusCode, "latency_ms", time.Since(startedAt).Milliseconds())

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var errorPayload map[string]any
		if json.Unmarshal(responseBody, &errorPayload) == nil {
			if message, ok := errorPayload["error"].(string); ok && strings.TrimSpace(message) != "" {
				return WebsiteAdsImageGenerateResponse{}, fmt.Errorf("hugging face image generation failed: %s", message)
			}
		}
		return WebsiteAdsImageGenerateResponse{}, fmt.Errorf("hugging face image generation failed with status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return WebsiteAdsImageGenerateResponse{}, fmt.Errorf("hugging face returned unexpected content type %q", contentType)
	}

	return WebsiteAdsImageGenerateResponse{
		Image:       responseBody,
		ContentType: contentType,
	}, nil
}

type noopWebsiteAdsImageClient struct{}

func NewNoopWebsiteAdsImageClient() WebsiteAdsImageClient {
	return &noopWebsiteAdsImageClient{}
}

func (c *noopWebsiteAdsImageClient) Generate(context.Context, WebsiteAdsImageGenerateRequest) (WebsiteAdsImageGenerateResponse, error) {
	return WebsiteAdsImageGenerateResponse{}, fmt.Errorf("website ads image client is not configured")
}
