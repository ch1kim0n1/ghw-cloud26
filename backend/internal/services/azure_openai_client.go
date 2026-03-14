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
	"path"
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
	baseURL, deployment, apiVersion := normalizeAzureOpenAIEndpoint(cfg.AzureOpenAIURL, cfg.AzureOpenAIDeployment, cfg.AzureOpenAIApiVersion)
	return &AzureOpenAIClient{
		baseURL:    baseURL,
		apiKey:     cfg.AzureOpenAIApiKey,
		apiVersion: apiVersion,
		deployment: deployment,
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

	endpoint := buildAzureOpenAIEndpoint(c.baseURL, c.deployment, c.apiVersion)
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

func normalizeAzureOpenAIEndpoint(rawBaseURL, deployment, apiVersion string) (string, string, string) {
	trimmed := strings.TrimRight(strings.TrimSpace(rawBaseURL), "/")
	if trimmed == "" {
		return "", deployment, apiVersion
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed, deployment, apiVersion
	}

	if strings.Contains(parsed.Path, "/openai/deployments/") {
		parts := strings.Split(parsed.Path, "/")
		for index := 0; index < len(parts)-2; index++ {
			if parts[index] == "deployments" && deployment == "" {
				deployment = parts[index+1]
			}
		}
		if queryVersion := strings.TrimSpace(parsed.Query().Get("api-version")); queryVersion != "" {
			apiVersion = queryVersion
		}
		parsed.RawQuery = ""
		parsed.Path = ""
		return strings.TrimRight(parsed.String(), "/"), deployment, apiVersion
	}

	return trimmed, deployment, apiVersion
}

func buildAzureOpenAIEndpoint(baseURL, deployment, apiVersion string) string {
	trimmedBase := strings.TrimRight(baseURL, "/")
	if trimmedBase == "" {
		return ""
	}
	parsed, err := url.Parse(trimmedBase)
	if err != nil {
		return fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", trimmedBase, url.PathEscape(deployment), url.QueryEscape(apiVersion))
	}
	parsed.Path = path.Join(parsed.Path, "openai", "deployments", deployment, "chat", "completions")
	query := parsed.Query()
	query.Set("api-version", apiVersion)
	parsed.RawQuery = query.Encode()
	return parsed.String()
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
