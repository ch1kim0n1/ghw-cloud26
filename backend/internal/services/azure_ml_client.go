package services

import (
	"context"
	"log/slog"
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
