package services

import (
	"context"
	"log/slog"
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
