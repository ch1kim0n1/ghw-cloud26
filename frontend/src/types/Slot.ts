export interface Slot {
  id: string;
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
  generation_error?: string | null;
}
