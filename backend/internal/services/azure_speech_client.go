package services

import (
	"context"
	"log/slog"
)

type NoopSpeechClient struct {
	logger *slog.Logger
}

func NewNoopSpeechClient(logger *slog.Logger) *NoopSpeechClient {
	return &NoopSpeechClient{logger: logger}
}

func (c *NoopSpeechClient) Synthesize(_ context.Context, req SpeechRequest) (SpeechResponse, error) {
	c.logger.Info("phase 0 speech placeholder invoked", "job_id", req.JobID)
	return SpeechResponse{}, ErrPlaceholderClient
}
