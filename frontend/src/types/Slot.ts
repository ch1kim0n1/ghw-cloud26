export interface Slot {
  id: string;
  job_id?: string;
  rank: number;
  scene_id: string;
  anchor_start_frame: number;
  anchor_end_frame: number;
  source_fps?: number;
  quiet_window_seconds: number;
  score: number;
  reasoning: string;
  status: string;
  suggested_product_line?: string | null;
  final_product_line?: string | null;
  product_line_mode?: string | null;
  generated_clip_path?: string | null;
  generated_audio_path?: string | null;
  generation_error?: string | null;
  context_relevance_score?: number | null;
  narrative_fit_score?: number | null;
  anchor_continuity_score?: number | null;
  metadata?: Record<string, unknown>;
}
