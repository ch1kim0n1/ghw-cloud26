package services

import (
	"context"
	"log/slog"
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
