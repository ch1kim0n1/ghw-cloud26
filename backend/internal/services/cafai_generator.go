package services

import (
	"context"
	"log/slog"
)

type NoopCafaiGenerator struct {
	logger *slog.Logger
}

func NewNoopCafaiGenerator(logger *slog.Logger) *NoopCafaiGenerator {
	return &NoopCafaiGenerator{logger: logger}
}

func (c *NoopCafaiGenerator) Generate(_ context.Context, req GenerationRequest) (GenerationResponse, error) {
	c.logger.Info("phase 0 cafai generator placeholder invoked", "job_id", req.JobID, "slot_id", req.SlotID)
	return GenerationResponse{}, ErrPlaceholderClient
}
