package services

import (
	"context"
	"io"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
)

type AnalysisRequest struct {
	JobID      string
	VideoPath  string
	ProductID  string
	CampaignID string
}

type AnalysisResponse struct {
	RequestID string
}

type AnalysisPollRequest struct {
	JobID      string
	RequestID  string
	VideoPath  string
	ProductID  string
	CampaignID string
	SourceFPS  float64
}

type AnalysisPollResponse struct {
	RequestID  string
	Status     string
	Scenes     []models.Scene
	PayloadRef string
	Message    string
}

type OpenAIRequest struct {
	JobID        string
	Purpose      string
	SystemPrompt string
	Prompt       string
	Temperature  float64
}

type OpenAIResponse struct {
	RequestID string
	Content   string
}

type GenerationRequest struct {
	JobID                   string
	SlotID                  string
	CampaignID              string
	ProductID               string
	SceneID                 string
	SourceVideoPath         string
	AnchorStartImagePath    string
	AnchorEndImagePath      string
	AnchorStartFrame        int
	AnchorEndFrame          int
	SourceFPS               float64
	TargetDurationSeconds   int
	ProductName             string
	ProductDescription      string
	ProductCategory         string
	ProductContextKeywords  []string
	ProductImagePath        string
	ProductSourceURL        string
	SceneNarrativeSummary   string
	ProductLineMode         string
	SuggestedProductLine    string
	FinalProductLine        string
	GenerationBrief         string
	SelectedSlotReasoning   string
	SelectedSlotQuietWindow float64
}

type GenerationResponse struct {
	RequestID          string
	Status             string
	GeneratedClipPath  string
	GeneratedAudioPath string
	PayloadRef         string
	Message            string
	Metadata           models.Metadata
}

type GenerationPollRequest struct {
	JobID     string
	SlotID    string
	RequestID string
}

type SpeechRequest struct {
	JobID string
	Text  string
}

type SpeechResponse struct {
	RequestID string
}

type BlobUploadRequest struct {
	JobID      string
	Path       string
	ObjectName string
}

type BlobUploadResponse struct {
	RequestID string
	BlobURI   string
}

type BlobDownloadRequest struct {
	JobID   string
	BlobURI string
}

type BlobDownloadResponse struct {
	RequestID string
	Body      io.ReadCloser
}

type RenderRequest struct {
	JobID                 string  `json:"job_id"`
	SlotID                string  `json:"slot_id"`
	SourceVideoBlobURI    string  `json:"source_video_blob_uri"`
	GeneratedClipBlobURI  string  `json:"generated_clip_blob_uri"`
	GeneratedAudioBlobURI string  `json:"generated_audio_blob_uri,omitempty"`
	AnchorStartFrame      int     `json:"anchor_start_frame"`
	AnchorEndFrame        int     `json:"anchor_end_frame"`
	SourceFPS             float64 `json:"source_fps"`
	TargetOutputName      string  `json:"target_output_name"`
	AudioStrategy         string  `json:"audio_strategy"`
}

type RenderResponse struct {
	RequestID       string          `json:"request_id"`
	Status          string          `json:"status"`
	PreviewBlobURI  string          `json:"preview_blob_uri,omitempty"`
	DurationSeconds float64         `json:"duration_seconds,omitempty"`
	PayloadRef      string          `json:"payload_ref,omitempty"`
	Message         string          `json:"message,omitempty"`
	Metadata        models.Metadata `json:"metadata,omitempty"`
}

type RenderPollRequest struct {
	JobID     string
	SlotID    string
	RequestID string
}

type AnchorFrameRequest struct {
	JobID            string
	SlotID           string
	VideoPath        string
	AnchorStartFrame int
	AnchorEndFrame   int
	SourceFPS        float64
}

type AnchorFrameArtifacts struct {
	AnchorStartImagePath string
	AnchorEndImagePath   string
}

type AnalysisClient interface {
	SubmitAnalysis(context.Context, AnalysisRequest) (AnalysisResponse, error)
	PollAnalysis(context.Context, AnalysisPollRequest) (AnalysisPollResponse, error)
}

type OpenAIClient interface {
	Complete(context.Context, OpenAIRequest) (OpenAIResponse, error)
}

type MLClient interface {
	SubmitGeneration(context.Context, GenerationRequest) (GenerationResponse, error)
	PollGeneration(context.Context, GenerationPollRequest) (GenerationResponse, error)
}

type SpeechClient interface {
	Synthesize(context.Context, SpeechRequest) (SpeechResponse, error)
}

type BlobStorageClient interface {
	Upload(context.Context, BlobUploadRequest) (BlobUploadResponse, error)
	Download(context.Context, BlobDownloadRequest) (BlobDownloadResponse, error)
}

type RenderClient interface {
	SubmitRender(context.Context, RenderRequest) (RenderResponse, error)
	PollRender(context.Context, RenderPollRequest) (RenderResponse, error)
}

type CafaiGenerator interface {
	Generate(context.Context, GenerationRequest) (GenerationResponse, error)
}

type AnchorFrameExtractor interface {
	Extract(context.Context, AnchorFrameRequest) (AnchorFrameArtifacts, error)
}
