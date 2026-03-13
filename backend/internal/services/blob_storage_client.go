package services

import (
	"context"
	"log/slog"
)

type NoopBlobStorageClient struct {
	logger *slog.Logger
}

func NewNoopBlobStorageClient(logger *slog.Logger) *NoopBlobStorageClient {
	return &NoopBlobStorageClient{logger: logger}
}

func (c *NoopBlobStorageClient) Upload(_ context.Context, req BlobUploadRequest) (BlobUploadResponse, error) {
	c.logger.Info("phase 0 blob placeholder invoked", "job_id", req.JobID, "path", req.Path)
	return BlobUploadResponse{}, ErrPlaceholderClient
}
