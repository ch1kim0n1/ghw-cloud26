export interface Campaign {
  campaign_id: string;
  job_id: string;
  product_id: string;
  name: string;
  status?: string;
  current_stage?: string;
  video_filename: string;
  video_path: string;
  source_fps?: number;
  duration_seconds?: number;
  target_ad_duration_seconds: number;
  created_at: string;
  updated_at?: string;
}
