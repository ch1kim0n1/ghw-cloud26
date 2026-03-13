package models

type JobLog struct {
	ID        int64    `json:"id,omitempty"`
	JobID     string   `json:"job_id"`
	Timestamp string   `json:"timestamp"`
	EventType string   `json:"event_type"`
	StageName string   `json:"stage_name,omitempty"`
	Message   string   `json:"message"`
	Details   Metadata `json:"details,omitempty"`
}
