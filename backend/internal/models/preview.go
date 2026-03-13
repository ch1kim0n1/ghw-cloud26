package models

type Preview struct {
	ID               string   `json:"id,omitempty"`
	JobID            string   `json:"job_id"`
	SlotID           string   `json:"slot_id"`
	Status           string   `json:"status"`
	OutputVideoPath  string   `json:"output_video_path,omitempty"`
	DownloadPath     string   `json:"download_path,omitempty"`
	DurationSeconds  float64  `json:"duration_seconds,omitempty"`
	CreatedAt        string   `json:"created_at,omitempty"`
	CompletedAt      *string  `json:"completed_at"`
	ArtifactManifest Metadata `json:"artifact_manifest,omitempty"`
	RenderMetrics    Metadata `json:"render_metrics,omitempty"`
}
