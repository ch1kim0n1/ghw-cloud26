export interface Preview {
  id?: string;
  job_id: string;
  slot_id: string;
  status: string;
  output_video_path?: string;
  download_path?: string;
  duration_seconds?: number;
  render_retry_count?: number;
  error_code?: string | null;
  error_message?: string | null;
  created_at?: string;
  artifact_manifest?: Record<string, unknown>;
  render_metrics?: Record<string, unknown>;
  completed_at?: string | null;
}
