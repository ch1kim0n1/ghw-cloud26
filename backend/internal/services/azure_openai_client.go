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
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

type NoopOpenAIClient struct {
	logger *slog.Logger
}

func NewNoopOpenAIClient(logger *slog.Logger) *NoopOpenAIClient {
	return &NoopOpenAIClient{logger: logger}
}

func (c *NoopOpenAIClient) Complete(_ context.Context, req OpenAIRequest) (OpenAIResponse, error) {
	c.logger.Info("phase 0 openai placeholder invoked", "job_id", req.JobID, "purpose", req.Purpose)
	return OpenAIResponse{}, ErrPlaceholderClient
}

type AzureOpenAIClient struct {
	baseURL    string
	apiKey     string
	apiVersion string
	deployment string
	logger     *slog.Logger
	httpClient *http.Client
}

func NewAzureOpenAIClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client) *AzureOpenAIClient {
	return &AzureOpenAIClient{
		baseURL:    strings.TrimRight(cfg.AzureOpenAIURL, "/"),
		apiKey:     cfg.AzureOpenAIApiKey,
		apiVersion: cfg.AzureOpenAIApiVersion,
		deployment: cfg.AzureOpenAIDeployment,
		logger:     logger,
		httpClient: httpClient,
	}
}

func (c *AzureOpenAIClient) Complete(ctx context.Context, req OpenAIRequest) (OpenAIResponse, error) {
	if c.baseURL == "" || c.apiKey == "" || c.deployment == "" || c.apiVersion == "" {
		return OpenAIResponse{}, fmt.Errorf("Azure OpenAI client is not configured")
	}

	payload := map[string]any{
		"messages": []map[string]string{
			{"role": "system", "content": req.SystemPrompt},
			{"role": "user", "content": req.Prompt},
		},
		"temperature":     req.Temperature,
		"response_format": map[string]string{"type": "json_object"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("marshal Azure OpenAI request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", c.baseURL, url.PathEscape(c.deployment), url.QueryEscape(c.apiVersion))
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("create Azure OpenAI request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("api-key", c.apiKey)

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("call Azure OpenAI: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("read Azure OpenAI response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return OpenAIResponse{}, fmt.Errorf("Azure OpenAI request failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var payloadResponse struct {
		ID      string `json:"id"`
		Choices []struct {
			Message struct {
				Content any `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(responseBody, &payloadResponse); err != nil {
		return OpenAIResponse{}, fmt.Errorf("decode Azure OpenAI response: %w", err)
	}
	if len(payloadResponse.Choices) == 0 {
		return OpenAIResponse{}, fmt.Errorf("Azure OpenAI response did not include a completion choice")
	}

	content, err := messageContent(payloadResponse.Choices[0].Message.Content)
	if err != nil {
		return OpenAIResponse{}, err
	}
	c.logger.Info("received Azure OpenAI completion", "job_id", req.JobID, "purpose", req.Purpose, "request_id", payloadResponse.ID)
	return OpenAIResponse{RequestID: payloadResponse.ID, Content: content}, nil
}

func messageContent(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			itemMap, ok := item.(map[string]any)
			if !ok {
				continue
			}
			text, _ := itemMap["text"].(string)
			if text == "" {
				if nested, ok := itemMap["text"].(map[string]any); ok {
					text, _ = nested["value"].(string)
				}
			}
			if text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n"), nil
	default:
		return "", fmt.Errorf("Azure OpenAI response content had unsupported shape")
	}
}
