export interface Preview {
  job_id: string;
  slot_id: string;
  status: string;
  output_video_path?: string;
  download_path?: string;
  duration_seconds?: number;
  created_at?: string;
  completed_at?: string | null;
}
