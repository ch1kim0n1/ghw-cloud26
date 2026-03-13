export interface Job {
  id: string;
  campaign_id: string;
  status: string;
  current_stage?: string;
  progress_percent: number;
  selected_slot_id: string | null;
  error_message: string | null;
  error_code: string | null;
  created_at: string;
  started_at: string | null;
  completed_at: string | null;
  metadata?: Record<string, unknown>;
}

export interface JobLog {
  timestamp: string;
  event_type: string;
  stage_name?: string;
  message: string;
}
