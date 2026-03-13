package models

type Scene struct {
	ID                        string   `json:"id"`
	JobID                     string   `json:"job_id"`
	SceneNumber               int      `json:"scene_number"`
	StartFrame                int      `json:"start_frame"`
	EndFrame                  int      `json:"end_frame"`
	StartSeconds              float64  `json:"start_seconds"`
	EndSeconds                float64  `json:"end_seconds"`
	MotionScore               *float64 `json:"motion_score"`
	StabilityScore            *float64 `json:"stability_score"`
	DialogueActivityScore     *float64 `json:"dialogue_activity_score"`
	LongestQuietWindowSeconds *float64 `json:"longest_quiet_window_seconds"`
	NarrativeSummary          string   `json:"narrative_summary,omitempty"`
	ContextKeywords           []string `json:"context_keywords,omitempty"`
	ActionIntensityScore      *float64 `json:"action_intensity_score"`
	AbruptCutRisk             *float64 `json:"abrupt_cut_risk"`
	Metadata                  Metadata `json:"metadata,omitempty"`
	CreatedAt                 string   `json:"created_at"`
}
