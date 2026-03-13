package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
)

type vultrObjectStorageAPI interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type VultrObjectStorageClient struct {
	endpoint string
	region   string
	bucket   string
	logger   *slog.Logger
	client   vultrObjectStorageAPI
}

func NewVultrObjectStorageClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client) *VultrObjectStorageClient {
	awsConfig := aws.Config{
		Region:      strings.TrimSpace(cfg.VultrObjectStorageRegion),
		Credentials: credentials.NewStaticCredentialsProvider(strings.TrimSpace(cfg.VultrObjectStorageAccessKey), strings.TrimSpace(cfg.VultrObjectStorageSecretKey), ""),
		HTTPClient:  httpClient,
	}
	client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.UsePathStyle = true
		options.BaseEndpoint = aws.String(strings.TrimRight(strings.TrimSpace(cfg.VultrObjectStorageEndpoint), "/"))
	})

	return newVultrObjectStorageClientWithAPI(cfg, logger, client)
}

func newVultrObjectStorageClientWithAPI(cfg config.Config, logger *slog.Logger, client vultrObjectStorageAPI) *VultrObjectStorageClient {
	return &VultrObjectStorageClient{
		endpoint: strings.TrimRight(strings.TrimSpace(cfg.VultrObjectStorageEndpoint), "/"),
		region:   strings.TrimSpace(cfg.VultrObjectStorageRegion),
		bucket:   strings.TrimSpace(cfg.VultrObjectStorageBucket),
		logger:   logger,
		client:   client,
	}
}

func (c *VultrObjectStorageClient) Upload(ctx context.Context, req BlobUploadRequest) (BlobUploadResponse, error) {
	if c.endpoint == "" || c.region == "" || c.bucket == "" || c.client == nil {
		return BlobUploadResponse{}, fmt.Errorf("Vultr object storage client is not configured")
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

	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(strings.TrimLeft(objectName, "/")),
		Body:        file,
		ContentType: aws.String(detectBlobContentType(objectName)),
	})
	if err != nil {
		return BlobUploadResponse{}, fmt.Errorf("upload Vultr object %s: %w", objectName, err)
	}

	blobURI := fmt.Sprintf("s3://%s/%s", c.bucket, strings.TrimLeft(objectName, "/"))
	c.logger.Info("uploaded preview artifact to Vultr object storage", "job_id", req.JobID, "object_name", objectName, "blob_uri", blobURI)
	return BlobUploadResponse{RequestID: blobURI, BlobURI: blobURI}, nil
}

func (c *VultrObjectStorageClient) Download(ctx context.Context, req BlobDownloadRequest) (BlobDownloadResponse, error) {
	if c.endpoint == "" || c.region == "" || c.bucket == "" || c.client == nil {
		return BlobDownloadResponse{}, fmt.Errorf("Vultr object storage client is not configured")
	}

	bucket, key, err := parseVultrObjectRef(req.BlobURI, c.bucket)
	if err != nil {
		return BlobDownloadResponse{}, err
	}

	response, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return BlobDownloadResponse{}, fmt.Errorf("download Vultr object %s: %w", req.BlobURI, err)
	}

	c.logger.Info("downloaded preview artifact from Vultr object storage", "job_id", req.JobID, "blob_uri", req.BlobURI)
	return BlobDownloadResponse{RequestID: req.BlobURI, Body: response.Body}, nil
}

func parseVultrObjectRef(blobURI, fallbackBucket string) (string, string, error) {
	trimmed := strings.TrimSpace(blobURI)
	if trimmed == "" {
		return "", "", fmt.Errorf("Vultr object download is missing blob uri")
	}

	if strings.HasPrefix(trimmed, "s3://") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return "", "", fmt.Errorf("parse Vultr object uri: %w", err)
		}
		bucket := parsed.Host
		key := strings.TrimLeft(parsed.Path, "/")
		if bucket == "" || key == "" {
			return "", "", fmt.Errorf("invalid Vultr object uri %q", blobURI)
		}
		return bucket, key, nil
	}

	if fallbackBucket == "" {
		return "", "", fmt.Errorf("invalid Vultr object uri %q", blobURI)
	}
	return fallbackBucket, strings.TrimLeft(trimmed, "/"), nil
}
