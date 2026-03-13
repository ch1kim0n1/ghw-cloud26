package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
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

func (c *NoopBlobStorageClient) Download(_ context.Context, req BlobDownloadRequest) (BlobDownloadResponse, error) {
	c.logger.Info("phase 0 blob download placeholder invoked", "job_id", req.JobID, "blob_uri", req.BlobURI)
	return BlobDownloadResponse{}, ErrPlaceholderClient
}

type AzureBlobStorageClient struct {
	baseURL    string
	container  string
	sasToken   string
	logger     *slog.Logger
	httpClient *http.Client
}

func NewAzureBlobStorageClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client) *AzureBlobStorageClient {
	return &AzureBlobStorageClient{
		baseURL:    strings.TrimRight(cfg.AzureBlobURL, "/"),
		container:  strings.TrimSpace(cfg.AzureBlobContainer),
		sasToken:   strings.TrimSpace(cfg.AzureBlobSASToken),
		logger:     logger,
		httpClient: httpClient,
	}
}

func (c *AzureBlobStorageClient) Upload(ctx context.Context, req BlobUploadRequest) (BlobUploadResponse, error) {
	if c.baseURL == "" || c.container == "" || c.sasToken == "" {
		return BlobUploadResponse{}, fmt.Errorf("Azure Blob client is not configured")
	}

	objectName := strings.TrimSpace(req.ObjectName)
	if objectName == "" {
		objectName = path.Base(req.Path)
	}

	file, err := os.Open(req.Path)
	if err != nil {
		return BlobUploadResponse{}, fmt.Errorf("open upload path %s: %w", req.Path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return BlobUploadResponse{}, fmt.Errorf("stat upload path %s: %w", req.Path, err)
	}

	endpoint, err := c.blobEndpoint(objectName)
	if err != nil {
		return BlobUploadResponse{}, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint.String(), file)
	if err != nil {
		return BlobUploadResponse{}, fmt.Errorf("create Azure Blob upload request: %w", err)
	}
	httpRequest.Header.Set("x-ms-blob-type", "BlockBlob")
	httpRequest.Header.Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	httpRequest.Header.Set("Content-Type", detectBlobContentType(objectName))

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return BlobUploadResponse{}, fmt.Errorf("upload blob object %s: %w", objectName, err)
	}
	defer response.Body.Close()

	if _, err := io.Copy(io.Discard, response.Body); err != nil {
		return BlobUploadResponse{}, fmt.Errorf("drain Azure Blob upload response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return BlobUploadResponse{}, fmt.Errorf("Azure Blob upload failed with status %d", response.StatusCode)
	}

	requestID := strings.TrimSpace(response.Header.Get("x-ms-request-id"))
	blobURI := strings.TrimSuffix(endpoint.String(), "?"+endpoint.RawQuery)
	c.logger.Info("uploaded preview artifact to Azure Blob", "job_id", req.JobID, "object_name", objectName, "request_id", requestID)
	return BlobUploadResponse{RequestID: requestID, BlobURI: blobURI}, nil
}

func (c *AzureBlobStorageClient) Download(ctx context.Context, req BlobDownloadRequest) (BlobDownloadResponse, error) {
	if c.baseURL == "" || c.container == "" || c.sasToken == "" {
		return BlobDownloadResponse{}, fmt.Errorf("Azure Blob client is not configured")
	}
	if strings.TrimSpace(req.BlobURI) == "" {
		return BlobDownloadResponse{}, fmt.Errorf("Azure Blob download is missing blob uri")
	}

	endpoint, err := url.Parse(req.BlobURI)
	if err != nil {
		return BlobDownloadResponse{}, fmt.Errorf("parse blob uri: %w", err)
	}
	query := endpoint.Query()
	if query.Encode() == "" {
		sasQuery, parseErr := url.ParseQuery(strings.TrimPrefix(c.sasToken, "?"))
		if parseErr != nil {
			return BlobDownloadResponse{}, fmt.Errorf("parse Azure Blob SAS token: %w", parseErr)
		}
		for key, values := range sasQuery {
			for _, value := range values {
				query.Add(key, value)
			}
		}
		endpoint.RawQuery = query.Encode()
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return BlobDownloadResponse{}, fmt.Errorf("create Azure Blob download request: %w", err)
	}

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return BlobDownloadResponse{}, fmt.Errorf("download blob %s: %w", req.BlobURI, err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		return BlobDownloadResponse{}, fmt.Errorf("Azure Blob download failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	requestID := strings.TrimSpace(response.Header.Get("x-ms-request-id"))
	c.logger.Info("downloaded preview artifact from Azure Blob", "job_id", req.JobID, "blob_uri", req.BlobURI, "request_id", requestID)
	return BlobDownloadResponse{RequestID: requestID, Body: response.Body}, nil
}

func (c *AzureBlobStorageClient) blobEndpoint(objectName string) (*url.URL, error) {
	base, err := url.Parse(fmt.Sprintf("%s/%s/%s", c.baseURL, url.PathEscape(c.container), strings.TrimLeft(objectName, "/")))
	if err != nil {
		return nil, fmt.Errorf("build Azure Blob endpoint: %w", err)
	}

	query, err := url.ParseQuery(strings.TrimPrefix(c.sasToken, "?"))
	if err != nil {
		return nil, fmt.Errorf("parse Azure Blob SAS token: %w", err)
	}
	base.RawQuery = query.Encode()
	return base, nil
}

func detectBlobContentType(objectName string) string {
	switch strings.ToLower(path.Ext(objectName)) {
	case ".mp4":
		return "video/mp4"
	case ".wav":
		return "audio/wav"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func newPhaseFourHTTPClient() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}
