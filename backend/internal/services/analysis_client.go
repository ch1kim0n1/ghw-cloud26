package services

import (
	"context"
	"log/slog"
)

type NoopAnalysisClient struct {
	logger *slog.Logger
}

func NewNoopAnalysisClient(logger *slog.Logger) *NoopAnalysisClient {
	return &NoopAnalysisClient{logger: logger}
}

func (c *NoopAnalysisClient) SubmitAnalysis(_ context.Context, req AnalysisRequest) (AnalysisResponse, error) {
	c.logger.Info("phase 0 analysis placeholder invoked", "job_id", req.JobID, "campaign_id", req.CampaignID)
	return AnalysisResponse{}, ErrPlaceholderClient
}

func (c *NoopAnalysisClient) PollAnalysis(_ context.Context, req AnalysisPollRequest) (AnalysisPollResponse, error) {
	c.logger.Info("phase 0 analysis poll placeholder invoked", "job_id", req.JobID, "request_id", req.RequestID)
	return AnalysisPollResponse{}, ErrPlaceholderClient
}
