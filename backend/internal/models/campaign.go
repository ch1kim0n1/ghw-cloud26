package models

type Campaign struct {
	ID                      string  `json:"campaign_id"`
	JobID                   string  `json:"job_id,omitempty"`
	ProductID               string  `json:"product_id"`
	Name                    string  `json:"name"`
	Status                  string  `json:"status,omitempty"`
	CurrentStage            string  `json:"current_stage,omitempty"`
	VideoFilename           string  `json:"video_filename"`
	VideoPath               string  `json:"video_path"`
	SourceFPS               float64 `json:"source_fps,omitempty"`
	DurationSeconds         float64 `json:"duration_seconds,omitempty"`
	TargetAdDurationSeconds int     `json:"target_ad_duration_seconds"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at,omitempty"`
}
