package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

const defaultNotionVersion = "2022-06-28"

type NotionAuditLogger struct {
	logger           *slog.Logger
	client           *http.Client
	apiBaseURL       string
	apiKey           string
	notionVersion    string
	jobsDatabaseID   string
	eventsDatabaseID string
	timeout          time.Duration
}

func NewNotionAuditLogger(cfg config.Config, logger *slog.Logger) JobAuditLogger {
	apiKey := strings.TrimSpace(cfg.NotionAPIKey)
	jobsDatabaseID := strings.TrimSpace(cfg.NotionJobsDatabaseID)
	eventsDatabaseID := strings.TrimSpace(cfg.NotionEventsDatabaseID)
	if apiKey == "" || jobsDatabaseID == "" || eventsDatabaseID == "" {
		return NewNoopJobAuditLogger()
	}

	timeout := cfg.NotionRequestTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	baseURL := strings.TrimSpace(cfg.NotionAPIBaseURL)
	if baseURL == "" {
		baseURL = "https://api.notion.com/v1"
	}
	version := strings.TrimSpace(cfg.NotionVersion)
	if version == "" {
		version = defaultNotionVersion
	}

	return &NotionAuditLogger{
		logger:           logger,
		client:           &http.Client{Timeout: timeout},
		apiBaseURL:       strings.TrimRight(baseURL, "/"),
		apiKey:           apiKey,
		notionVersion:    version,
		jobsDatabaseID:   jobsDatabaseID,
		eventsDatabaseID: eventsDatabaseID,
		timeout:          timeout,
	}
}

func (n *NotionAuditLogger) Record(ctx context.Context, event JobAuditEvent) error {
	if strings.TrimSpace(event.JobID) == "" {
		return nil
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.Metadata == nil {
		event.Metadata = models.Metadata{}
	}

	recordCtx := ctx
	if recordCtx == nil {
		recordCtx = context.Background()
	}
	if _, hasDeadline := recordCtx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		recordCtx, cancel = context.WithTimeout(recordCtx, n.timeout)
		defer cancel()
	}

	if err := n.writeEvent(recordCtx, event); err != nil {
		return err
	}
	if err := n.upsertJob(recordCtx, event); err != nil {
		return err
	}
	return nil
}

func (n *NotionAuditLogger) Health(ctx context.Context) AuditHealth {
	health := AuditHealth{Enabled: true, Status: "degraded", Details: "notion connectivity check pending"}
	healthCtx := ctx
	if healthCtx == nil {
		healthCtx = context.Background()
	}
	if _, hasDeadline := healthCtx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		healthCtx, cancel = context.WithTimeout(healthCtx, n.timeout)
		defer cancel()
	}

	if err := n.doRequest(healthCtx, http.MethodGet, "/databases/"+n.jobsDatabaseID, nil, nil); err != nil {
		health.Details = "jobs database check failed"
		return health
	}
	if err := n.doRequest(healthCtx, http.MethodGet, "/databases/"+n.eventsDatabaseID, nil, nil); err != nil {
		health.Details = "events database check failed"
		return health
	}

	health.Status = "healthy"
	health.Details = "notion audit sink connected"
	return health
}

func (n *NotionAuditLogger) writeEvent(ctx context.Context, event JobAuditEvent) error {
	metadata := safeJSONString(event.Metadata, 1800)
	payload := map[string]any{
		"parent": map[string]any{
			"database_id": n.eventsDatabaseID,
		},
		"properties": map[string]any{
			"Event":         richTitle(fmt.Sprintf("%s | %s", event.EventType, event.JobID)),
			"Job ID":        richText(event.JobID),
			"Campaign ID":   richText(event.CampaignID),
			"Event Type":    richText(event.EventType),
			"Status":        richText(event.Status),
			"Current Stage": richText(event.CurrentStage),
			"Error Code":    richText(event.ErrorCode),
			"Timestamp":     dateValue(event.Timestamp),
			"Message":       richText(event.Message),
			"Metadata":      richText(metadata),
		},
	}
	return n.doRequest(ctx, http.MethodPost, "/pages", payload, nil)
}

func (n *NotionAuditLogger) upsertJob(ctx context.Context, event JobAuditEvent) error {
	jobPageID, err := n.lookupJobPageID(ctx, event.JobID)
	if err != nil {
		return err
	}

	properties := map[string]any{
		"Name":          richTitle(event.JobID),
		"Job ID":        richText(event.JobID),
		"Campaign ID":   richText(event.CampaignID),
		"Status":        richText(event.Status),
		"Current Stage": richText(event.CurrentStage),
		"Last Event":    richText(event.EventType),
		"Error Code":    richText(event.ErrorCode),
		"Updated At":    dateValue(event.Timestamp),
		"Summary":       richText(event.Message),
	}

	if jobPageID == "" {
		payload := map[string]any{
			"parent": map[string]any{
				"database_id": n.jobsDatabaseID,
			},
			"properties": properties,
		}
		return n.doRequest(ctx, http.MethodPost, "/pages", payload, nil)
	}

	payload := map[string]any{"properties": properties}
	return n.doRequest(ctx, http.MethodPatch, "/pages/"+jobPageID, payload, nil)
}

func (n *NotionAuditLogger) lookupJobPageID(ctx context.Context, jobID string) (string, error) {
	payload := map[string]any{
		"filter": map[string]any{
			"property": "Job ID",
			"rich_text": map[string]any{
				"equals": jobID,
			},
		},
		"page_size": 1,
	}
	var response struct {
		Results []struct {
			ID string `json:"id"`
		} `json:"results"`
	}
	if err := n.doRequest(ctx, http.MethodPost, "/databases/"+n.jobsDatabaseID+"/query", payload, &response); err != nil {
		return "", err
	}
	if len(response.Results) == 0 {
		return "", nil
	}
	return response.Results[0].ID, nil
}

func (n *NotionAuditLogger) doRequest(ctx context.Context, method, path string, payload any, responseDest any) error {
	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal notion payload: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, n.apiBaseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build notion request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+n.apiKey)
	req.Header.Set("Notion-Version", n.notionVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("execute notion request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read notion response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if n.logger != nil {
			n.logger.Warn("notion request failed", "method", method, "path", path, "status_code", resp.StatusCode, "response", string(respBody))
		}
		return fmt.Errorf("notion request failed with status %d", resp.StatusCode)
	}

	if responseDest != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, responseDest); err != nil {
			return fmt.Errorf("decode notion response: %w", err)
		}
	}
	return nil
}

func richTitle(value string) map[string]any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = "-"
	}
	return map[string]any{
		"title": []map[string]any{{
			"text": map[string]any{"content": truncate(trimmed, 1900)},
		}},
	}
}

func richText(value string) map[string]any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = "-"
	}
	return map[string]any{
		"rich_text": []map[string]any{{
			"text": map[string]any{"content": truncate(trimmed, 1900)},
		}},
	}
}

func dateValue(value time.Time) map[string]any {
	if value.IsZero() {
		value = time.Now().UTC()
	}
	return map[string]any{
		"date": map[string]any{
			"start": value.UTC().Format(time.RFC3339),
		},
	}
}

func safeJSONString(value models.Metadata, maxLen int) string {
	if value == nil {
		return "{}"
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return truncate(string(encoded), maxLen)
}

func truncate(value string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(value) <= maxLen {
		return value
	}
	if maxLen <= 3 {
		return value[:maxLen]
	}
	return value[:maxLen-3] + "..."
}
