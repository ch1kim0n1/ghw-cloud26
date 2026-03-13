package models

type Job struct {
	ID              string   `json:"id"`
	CampaignID      string   `json:"campaign_id"`
	Status          string   `json:"status"`
	CurrentStage    string   `json:"current_stage,omitempty"`
	ProgressPercent int      `json:"progress_percent"`
	SelectedSlotID  *string  `json:"selected_slot_id"`
	ErrorMessage    *string  `json:"error_message"`
	ErrorCode       *string  `json:"error_code"`
	CreatedAt       string   `json:"created_at"`
	StartedAt       *string  `json:"started_at"`
	CompletedAt     *string  `json:"completed_at"`
	Metadata        Metadata `json:"metadata,omitempty"`
}
