package models

type Slot struct {
	ID                    string   `json:"id"`
	JobID                 string   `json:"job_id,omitempty"`
	Rank                  int      `json:"rank"`
	SceneID               string   `json:"scene_id"`
	AnchorStartFrame      int      `json:"anchor_start_frame"`
	AnchorEndFrame        int      `json:"anchor_end_frame"`
	SourceFPS             float64  `json:"source_fps,omitempty"`
	QuietWindowSeconds    float64  `json:"quiet_window_seconds"`
	Score                 float64  `json:"score"`
	Reasoning             string   `json:"reasoning"`
	Status                string   `json:"status"`
	SuggestedProductLine  *string  `json:"suggested_product_line"`
	FinalProductLine      *string  `json:"final_product_line"`
	ProductLineMode       *string  `json:"product_line_mode"`
	GeneratedClipPath     *string  `json:"generated_clip_path"`
	GeneratedAudioPath    *string  `json:"generated_audio_path"`
	GenerationError       *string  `json:"generation_error"`
	ContextRelevanceScore *float64 `json:"context_relevance_score,omitempty"`
	NarrativeFitScore     *float64 `json:"narrative_fit_score,omitempty"`
	AnchorContinuityScore *float64 `json:"anchor_continuity_score,omitempty"`
	Metadata              Metadata `json:"metadata,omitempty"`
	CreatedAt             string   `json:"created_at,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
}
