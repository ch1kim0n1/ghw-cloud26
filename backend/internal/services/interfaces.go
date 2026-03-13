package services

import "context"

type AnalysisRequest struct {
	JobID      string
	VideoPath  string
	ProductID  string
	CampaignID string
}

type AnalysisResponse struct {
	RequestID string
}

type OpenAIRequest struct {
	JobID   string
	Prompt  string
	Purpose string
}

type OpenAIResponse struct {
	RequestID string
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
