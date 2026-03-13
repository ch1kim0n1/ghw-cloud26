package constants

const (
	JobStatusQueued     = "queued"
	JobStatusAnalyzing  = "analyzing"
	JobStatusGenerating = "generating"
	JobStatusStitching  = "stitching"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)

const (
	StageReadyForAnalysis   = "ready_for_analysis"
	StageAnalysisSubmission = "analysis_submission"
	StageAnalysisPoll       = "analysis_poll"
	StageSlotSelection      = "slot_selection"
	StageLineReview         = "line_review"
	StageGenerationSubmit   = "generation_submission"
	StageGenerationPoll     = "generation_poll"
	StageRender             = "render"
	StageRenderSubmit       = "render_submission"
	StageRenderPoll         = "render_poll"
)

const (
	SlotStatusProposed   = "proposed"
	SlotStatusSelected   = "selected"
	SlotStatusRejected   = "rejected"
	SlotStatusGenerating = "generating"
	SlotStatusGenerated  = "generated"
	SlotStatusFailed     = "failed"
)

const (
	ErrorCodeInvalidRequest      = "INVALID_REQUEST"
	ErrorCodeInvalidVideoCodec   = "INVALID_VIDEO_CODEC"
	ErrorCodeInvalidVideoLength  = "INVALID_VIDEO_DURATION"
	ErrorCodeProductInputMissing = "PRODUCT_INPUT_MISSING"
	ErrorCodeResourceNotFound    = "RESOURCE_NOT_FOUND"
	ErrorCodeNoSuitableSlot      = "NO_SUITABLE_SLOT_FOUND"
	ErrorCodeAnalysisFailed      = "ANALYSIS_FAILED"
	ErrorCodeGenerationFailed    = "GENERATION_FAILED"
	ErrorCodePreviewRenderFailed = "PREVIEW_RENDER_FAILED"
	ErrorCodeDatabaseError       = "DATABASE_ERROR"
	ErrorCodeStorageError        = "STORAGE_ERROR"
	ErrorCodeNotImplemented      = "NOT_IMPLEMENTED"
)
