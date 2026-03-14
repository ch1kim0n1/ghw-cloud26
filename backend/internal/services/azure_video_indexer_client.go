package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type AzureVideoIndexerClient struct {
	baseURL     string
	accountID   string
	location    string
	accessToken string
	logger      *slog.Logger
	httpClient  *http.Client
}

func NewAzureVideoIndexerClient(cfg config.Config, logger *slog.Logger, httpClient *http.Client) *AzureVideoIndexerClient {
	return &AzureVideoIndexerClient{
		baseURL:     strings.TrimRight(cfg.AzureVideoIndexerURL, "/"),
		accountID:   cfg.AzureVideoIndexerAccountID,
		location:    normalizeAzureLocation(cfg.AzureVideoIndexerLocation),
		accessToken: cfg.AzureVideoIndexerAccessToken,
		logger:      logger,
		httpClient:  httpClient,
	}
}

func (c *AzureVideoIndexerClient) SubmitAnalysis(ctx context.Context, req AnalysisRequest) (AnalysisResponse, error) {
	file, err := os.Open(req.VideoPath)
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("open source video: %w", err)
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filepath.Base(req.VideoPath))
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("create multipart file part: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return AnalysisResponse{}, fmt.Errorf("copy source video into request: %w", err)
	}
	if err := writer.Close(); err != nil {
		return AnalysisResponse{}, fmt.Errorf("close multipart writer: %w", err)
	}

	endpoint, err := c.videoEndpoint("/Videos")
	if err != nil {
		return AnalysisResponse{}, err
	}
	query := endpoint.Query()
	query.Set("name", req.JobID)
	query.Set("privacy", "private")
	query.Set("streamingPreset", "NoStreaming")
	query.Set("accessToken", c.accessToken)
	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), &body)
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("create video indexer submission request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", writer.FormDataContentType())

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("submit video to Azure Video Indexer: %w", err)
	}
	defer response.Body.Close()

	var payload map[string]any
	if err := decodeJSONResponse(response, &payload); err != nil {
		return AnalysisResponse{}, err
	}

	requestID := stringValue(payload["id"])
	if requestID == "" {
		requestID = stringValue(payload["videoId"])
	}
	if requestID == "" {
		return AnalysisResponse{}, fmt.Errorf("video indexer submission response missing video id")
	}

	c.logger.Info("submitted video to Azure Video Indexer", "job_id", req.JobID, "request_id", requestID)
	return AnalysisResponse{RequestID: requestID}, nil
}

func (c *AzureVideoIndexerClient) PollAnalysis(ctx context.Context, req AnalysisPollRequest) (AnalysisPollResponse, error) {
	endpoint, err := c.videoEndpoint("/Videos/" + url.PathEscape(req.RequestID) + "/Index")
	if err != nil {
		return AnalysisPollResponse{}, err
	}
	query := endpoint.Query()
	query.Set("accessToken", c.accessToken)
	endpoint.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return AnalysisPollResponse{}, fmt.Errorf("create video indexer poll request: %w", err)
	}

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return AnalysisPollResponse{}, fmt.Errorf("poll Azure Video Indexer: %w", err)
	}
	defer response.Body.Close()

	var payload map[string]any
	if err := decodeJSONResponse(response, &payload); err != nil {
		return AnalysisPollResponse{}, err
	}

	state := strings.ToLower(stringValue(payload["state"]))
	if state == "" {
		state = strings.ToLower(stringValue(payload["status"]))
	}
	switch state {
	case "", "uploaded", "processing", "running", "pending":
		return AnalysisPollResponse{RequestID: req.RequestID, Status: "pending"}, nil
	case "processed", "completed", "succeeded":
		scenes := extractVideoIndexerScenes(payload, req.JobID, req.SourceFPS)
		return AnalysisPollResponse{
			RequestID:  req.RequestID,
			Status:     "completed",
			Scenes:     scenes,
			PayloadRef: req.RequestID,
			Metadata:   extractVideoIndexerMetadata(payload),
		}, nil
	case "failed", "error":
		return AnalysisPollResponse{
			RequestID: req.RequestID,
			Status:    "failed",
			Message:   stringValue(payload["message"]),
		}, nil
	default:
		return AnalysisPollResponse{
			RequestID: req.RequestID,
			Status:    state,
		}, nil
	}
}

func (c *AzureVideoIndexerClient) videoEndpoint(path string) (*url.URL, error) {
	if c.baseURL == "" || c.accountID == "" || c.location == "" || c.accessToken == "" {
		return nil, fmt.Errorf("Azure Video Indexer client is not configured")
	}
	return url.Parse(fmt.Sprintf("%s/%s/Accounts/%s%s", c.baseURL, url.PathEscape(c.location), url.PathEscape(c.accountID), path))
}

func normalizeAzureLocation(value string) string {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return ""
	}
	return strings.ReplaceAll(trimmed, " ", "")
}

func decodeJSONResponse(response *http.Response, target any) error {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read provider response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("provider request failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	if len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode provider JSON response: %w", err)
	}
	return nil
}

func extractVideoIndexerScenes(payload map[string]any, jobID string, sourceFPS float64) []models.Scene {
	if sourceFPS <= 0 {
		sourceFPS = 24
	}

	summarizedScenes := sliceValue(nestedValue(payload, "summarizedInsights", "scenes"))
	if len(summarizedScenes) == 0 {
		summarizedScenes = sliceValue(nestedValue(firstVideo(payload), "insights", "shots"))
	}

	transcriptEntries := sliceValue(nestedValue(firstVideo(payload), "insights", "transcript"))
	labelEntries := append(
		sliceValue(nestedValue(firstVideo(payload), "insights", "labels")),
		sliceValue(nestedValue(firstVideo(payload), "insights", "keywords"))...,
	)

	scenes := make([]models.Scene, 0, len(summarizedScenes))
	for index, rawScene := range summarizedScenes {
		sceneMap := mapValue(rawScene)
		startSeconds, endSeconds := sceneRangeSeconds(sceneMap)
		if endSeconds <= startSeconds {
			continue
		}

		sceneNumber := intValue(sceneMap["id"])
		if sceneNumber == 0 {
			sceneNumber = index + 1
		}

		transcriptTexts := overlappingTexts(transcriptEntries, startSeconds, endSeconds, "text")
		contextKeywords := overlappingTexts(labelEntries, startSeconds, endSeconds, "name")
		quietWindow := estimateQuietWindowSeconds(startSeconds, endSeconds, transcriptEntries)
		dialogueActivity := estimateDialogueActivity(startSeconds, endSeconds, transcriptEntries)
		motionScore := estimateMotionScore(startSeconds, endSeconds, sceneMap)
		stabilityScore := clamp01(1 - motionScore*0.7)
		actionIntensity := clamp01(maxFloat(motionScore*0.8, dialogueActivity*0.6))
		abruptCutRisk := clamp01(motionScore * 0.6)

		scene := models.Scene{
			ID:                        fmt.Sprintf("scene_%s_%03d", jobID, sceneNumber),
			JobID:                     jobID,
			SceneNumber:               sceneNumber,
			StartFrame:                int(startSeconds * sourceFPS),
			EndFrame:                  int(endSeconds * sourceFPS),
			StartSeconds:              startSeconds,
			EndSeconds:                endSeconds,
			MotionScore:               floatPtr(roundScore(motionScore)),
			StabilityScore:            floatPtr(roundScore(stabilityScore)),
			DialogueActivityScore:     floatPtr(roundScore(dialogueActivity)),
			LongestQuietWindowSeconds: floatPtr(roundScore(quietWindow)),
			NarrativeSummary:          strings.Join(transcriptTexts, " "),
			ContextKeywords:           uniqueSorted(contextKeywords),
			ActionIntensityScore:      floatPtr(roundScore(actionIntensity)),
			AbruptCutRisk:             floatPtr(roundScore(abruptCutRisk)),
			Metadata: models.Metadata{
				"provider":        "azure_video_indexer",
				"provider_scene":  sceneMap["id"],
				"transcript_size": len(transcriptTexts),
			},
		}
		scenes = append(scenes, scene)
	}

	return scenes
}

func extractVideoIndexerMetadata(payload map[string]any) models.Metadata {
	metadata := models.Metadata{}
	language := firstNonEmptyString(
		stringValue(payload["sourceLanguage"]),
		stringValue(payload["language"]),
		stringValue(nestedValue(firstVideo(payload), "insights", "sourceLanguage")),
		stringValue(nestedValue(firstVideo(payload), "insights", "language")),
	)
	if language == "" {
		return metadata
	}
	metadata["content_language"] = language
	metadata["language_detection_source"] = "azure_video_indexer"
	metadata["language_confidence"] = 0.95
	return metadata
}

func firstVideo(payload map[string]any) map[string]any {
	videos := sliceValue(payload["videos"])
	if len(videos) == 0 {
		return map[string]any{}
	}
	return mapValue(videos[0])
}

func sceneRangeSeconds(scene map[string]any) (float64, float64) {
	instances := sliceValue(scene["instances"])
	if len(instances) > 0 {
		first := mapValue(instances[0])
		last := mapValue(instances[len(instances)-1])
		return durationValue(first["start"]), durationValue(last["end"])
	}
	return durationValue(scene["startTime"]), durationValue(scene["endTime"])
}

func overlappingTexts(entries []any, startSeconds, endSeconds float64, textKey string) []string {
	results := make([]string, 0)
	for _, entry := range entries {
		entryMap := mapValue(entry)
		text := strings.TrimSpace(stringValue(entryMap[textKey]))
		if text == "" {
			text = strings.TrimSpace(stringValue(entryMap["name"]))
		}
		if text == "" {
			continue
		}
		for _, instance := range sliceValue(entryMap["instances"]) {
			instanceMap := mapValue(instance)
			instanceStart := durationValue(instanceMap["start"])
			instanceEnd := durationValue(instanceMap["end"])
			if instanceEnd <= startSeconds || instanceStart >= endSeconds {
				continue
			}
			results = append(results, text)
			break
		}
	}
	return results
}

func estimateQuietWindowSeconds(startSeconds, endSeconds float64, transcriptEntries []any) float64 {
	quietWindow := endSeconds - startSeconds
	lastEdge := startSeconds
	for _, entry := range transcriptEntries {
		entryMap := mapValue(entry)
		for _, instance := range sliceValue(entryMap["instances"]) {
			instanceMap := mapValue(instance)
			instanceStart := durationValue(instanceMap["start"])
			instanceEnd := durationValue(instanceMap["end"])
			if instanceEnd <= startSeconds || instanceStart >= endSeconds {
				continue
			}
			quietWindow = maxFloat(quietWindow, instanceStart-lastEdge)
			if instanceEnd > lastEdge {
				lastEdge = instanceEnd
			}
		}
	}
	return roundScore(maxFloat(quietWindow, endSeconds-lastEdge))
}

func estimateDialogueActivity(startSeconds, endSeconds float64, transcriptEntries []any) float64 {
	duration := endSeconds - startSeconds
	if duration <= 0 {
		return 0
	}
	var totalTranscriptSeconds float64
	for _, entry := range transcriptEntries {
		entryMap := mapValue(entry)
		for _, instance := range sliceValue(entryMap["instances"]) {
			instanceMap := mapValue(instance)
			instanceStart := durationValue(instanceMap["start"])
			instanceEnd := durationValue(instanceMap["end"])
			if instanceEnd <= startSeconds || instanceStart >= endSeconds {
				continue
			}
			totalTranscriptSeconds += minFloat(instanceEnd, endSeconds) - maxFloat(instanceStart, startSeconds)
		}
	}
	return clamp01(totalTranscriptSeconds / duration)
}

func estimateMotionScore(startSeconds, endSeconds float64, scene map[string]any) float64 {
	duration := endSeconds - startSeconds
	if duration <= 0 {
		return 0
	}
	instances := sliceValue(scene["instances"])
	if len(instances) == 0 {
		return 0.18
	}
	return clamp01(float64(len(instances)) / maxFloat(1, duration/3))
}

func nestedValue(value any, keys ...string) any {
	current := value
	for _, key := range keys {
		currentMap := mapValue(current)
		current = currentMap[key]
	}
	return current
}

func mapValue(value any) map[string]any {
	typed, _ := value.(map[string]any)
	if typed == nil {
		return map[string]any{}
	}
	return typed
}

func sliceValue(value any) []any {
	typed, _ := value.([]any)
	if typed == nil {
		return []any{}
	}
	return typed
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	default:
		return ""
	}
}

func intValue(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, _ := strconv.Atoi(typed)
		return parsed
	default:
		return 0
	}
}

func durationValue(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case string:
		return parseDurationString(typed)
	default:
		return 0
	}
}

func parseDurationString(value string) float64 {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}
	if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return parsed
	}
	parts := strings.Split(trimmed, ":")
	if len(parts) == 0 {
		return 0
	}
	var total float64
	multiplier := 1.0
	for index := len(parts) - 1; index >= 0; index-- {
		segment, err := strconv.ParseFloat(parts[index], 64)
		if err != nil {
			return 0
		}
		total += segment * multiplier
		multiplier *= 60
	}
	return total
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	results := make([]string, 0, len(set))
	for value := range set {
		results = append(results, value)
	}
	sort.Strings(results)
	return results
}

func minFloat(left, right float64) float64 {
	if left < right {
		return left
	}
	return right
}

func maxFloat(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
}
