CREATE TABLE IF NOT EXISTS products (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  category TEXT,
  context_keywords_json TEXT,
  source_url TEXT,
  image_path TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS campaigns (
  id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  name TEXT NOT NULL,
  video_filename TEXT NOT NULL,
  video_path TEXT NOT NULL,
  source_fps REAL NOT NULL,
  duration_seconds REAL NOT NULL,
  target_ad_duration_seconds INTEGER NOT NULL DEFAULT 6,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  campaign_id TEXT NOT NULL,
  status TEXT NOT NULL CHECK (
    status IN ('queued', 'analyzing', 'generating', 'stitching', 'completed', 'failed')
  ),
  current_stage TEXT,
  progress_percent INTEGER NOT NULL DEFAULT 0,
  selected_slot_id TEXT,
  repick_count INTEGER NOT NULL DEFAULT 0,
  error_code TEXT,
  error_message TEXT,
  metadata_json TEXT,
  created_at TEXT NOT NULL,
  started_at TEXT,
  completed_at TEXT,
  FOREIGN KEY (campaign_id) REFERENCES campaigns(id)
);

CREATE TABLE IF NOT EXISTS scenes (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL,
  scene_number INTEGER NOT NULL,
  start_frame INTEGER NOT NULL,
  end_frame INTEGER NOT NULL,
  start_seconds REAL NOT NULL,
  end_seconds REAL NOT NULL,
  motion_score REAL,
  stability_score REAL,
  dialogue_activity_score REAL,
  longest_quiet_window_seconds REAL,
  narrative_summary TEXT,
  context_keywords_json TEXT,
  action_intensity_score REAL,
  abrupt_cut_risk REAL,
  created_at TEXT NOT NULL,
  FOREIGN KEY (job_id) REFERENCES jobs(id)
);

CREATE TABLE IF NOT EXISTS slots (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL,
  scene_id TEXT NOT NULL,
  rank INTEGER NOT NULL,
  anchor_start_frame INTEGER NOT NULL,
  anchor_end_frame INTEGER NOT NULL,
  quiet_window_seconds REAL NOT NULL,
  score REAL NOT NULL,
  context_relevance_score REAL,
  narrative_fit_score REAL,
  anchor_continuity_score REAL,
  reasoning TEXT NOT NULL,
  status TEXT NOT NULL CHECK (
    status IN ('proposed', 'selected', 'rejected', 'generating', 'generated', 'failed')
  ),
  suggested_product_line TEXT,
  final_product_line TEXT,
  product_line_mode TEXT CHECK (
    product_line_mode IN ('auto', 'operator', 'disabled')
  ),
  rejection_note TEXT,
  generated_clip_path TEXT,
  generated_audio_path TEXT,
  generation_error TEXT,
  metadata_json TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (job_id) REFERENCES jobs(id),
  FOREIGN KEY (scene_id) REFERENCES scenes(id),
  UNIQUE (job_id, rank)
);

CREATE TABLE IF NOT EXISTS job_previews (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL UNIQUE,
  slot_id TEXT NOT NULL,
  status TEXT NOT NULL CHECK (
    status IN ('pending', 'stitching', 'completed', 'failed')
  ),
  output_video_path TEXT,
  duration_seconds REAL,
  render_retry_count INTEGER NOT NULL DEFAULT 0,
  artifact_manifest_json TEXT,
  render_metrics_json TEXT,
  error_code TEXT,
  error_message TEXT,
  created_at TEXT NOT NULL,
  completed_at TEXT,
  FOREIGN KEY (job_id) REFERENCES jobs(id),
  FOREIGN KEY (slot_id) REFERENCES slots(id)
);

CREATE TABLE IF NOT EXISTS job_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  job_id TEXT NOT NULL,
  timestamp TEXT NOT NULL,
  event_type TEXT NOT NULL,
  stage_name TEXT,
  message TEXT NOT NULL,
  details_json TEXT,
  FOREIGN KEY (job_id) REFERENCES jobs(id)
);

CREATE INDEX IF NOT EXISTS idx_campaigns_product_id ON campaigns(product_id);
CREATE INDEX IF NOT EXISTS idx_jobs_campaign_id ON jobs(campaign_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_scenes_job_id ON scenes(job_id);
CREATE INDEX IF NOT EXISTS idx_slots_job_id ON slots(job_id);
CREATE INDEX IF NOT EXISTS idx_slots_job_rank ON slots(job_id, rank);
CREATE INDEX IF NOT EXISTS idx_slots_status ON slots(status);
CREATE INDEX IF NOT EXISTS idx_job_previews_job_id ON job_previews(job_id);
CREATE INDEX IF NOT EXISTS idx_job_logs_job_id_timestamp ON job_logs(job_id, timestamp DESC);
