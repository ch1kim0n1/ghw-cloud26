package services

import (
	"context"

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
	JobID  string
	SlotID string
}

type GenerationResponse struct {
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
	JobID string
	Path  string
}

type BlobUploadResponse struct {
	RequestID string
	BlobURI   string
}

type RenderRequest struct {
	JobID  string
	SlotID string
}

type RenderResponse struct {
	RequestID string
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
}

type SpeechClient interface {
	Synthesize(context.Context, SpeechRequest) (SpeechResponse, error)
}

type BlobStorageClient interface {
	Upload(context.Context, BlobUploadRequest) (BlobUploadResponse, error)
}

type RenderClient interface {
	SubmitRender(context.Context, RenderRequest) (RenderResponse, error)
}

type CafaiGenerator interface {
	Generate(context.Context, GenerationRequest) (GenerationResponse, error)
}
