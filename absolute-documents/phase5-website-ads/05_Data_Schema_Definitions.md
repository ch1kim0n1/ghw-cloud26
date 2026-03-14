# Phase 5: Dynamic Website Ads - Data Schema Definitions

## 1. Purpose

Define the SQLite schema extensions for Phase 5 website ad generation. This includes new tables for ad jobs, creative requests, variants, and artifacts.

## 2. Design Principles

- **Reuse existing patterns** from Phases 0-4 schema (campaigns, jobs, artifacts)
- **Separate namespace** — website ads use distinct table prefix (`ad_*`) to avoid collision
- **Maintain relationships** — foreign keys to existing products table
- **Support job queuing** — status tracking for async processing
- **Preserve artifacts** — store references to cloud/local storage for downloads

## 3. New Tables

### 3.1 `website_ad_jobs`

Parent table for website ad generation jobs.

**Purpose:** Track each user-initiated website ad generation request.

```sql
CREATE TABLE website_ad_jobs (
  id TEXT PRIMARY KEY,
  -- Reference to advertised product
  product_id TEXT NOT NULL,
  FOREIGN KEY (product_id) REFERENCES products(id),

  -- Article source details
  article_source TEXT NOT NULL,           -- 'url', 'text', or 'headline_body'
  article_url TEXT,                        -- if source == 'url'
  article_title TEXT,                      -- extracted title or user-provided
  article_text_hash TEXT NOT NULL,        -- hash of article for dedup check
  article_text_preview TEXT NOT NULL,     -- first 500 chars for display

  -- Configuration
  brand_voice TEXT,                        -- 'premium', 'casual', 'professional', 'playful'

  -- Processing state
  status TEXT NOT NULL,                    -- 'requested', 'analyzing', 'generating', 'completed', 'failed'
  error_code TEXT,                         -- populated if status == 'failed'
  error_message TEXT,                      -- human-readable error

  -- Creative selection
  selected_creative_id INTEGER,            -- 1, 2, or 3; NULL until selection
  regeneration_count INTEGER NOT NULL DEFAULT 0,
  max_regenerations INTEGER NOT NULL DEFAULT 2,

  -- Timestamps
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT                        -- populated when status == 'completed'
);

-- Indexes for common queries
CREATE INDEX idx_website_ad_jobs_status ON website_ad_jobs(status);
CREATE INDEX idx_website_ad_jobs_product_id ON website_ad_jobs(product_id);
CREATE INDEX idx_website_ad_jobs_created_at ON website_ad_jobs(created_at DESC);
```

**Notes:**
- `id` format: `wad_` prefix (e.g., `wad_001`, `wad_abc123`)
- `status` values: matches Phase 0-4 job pattern (coarse states)
- `article_text_hash` enables dedup if user submits same article twice
- `selected_creative_id` is 1-based (creative candidate 1, 2, or 3)

---

### 3.2 `ad_creative_requests`

Track each creative generation request (main job + regenerations).

**Purpose:** Audit trail and artifact references for creative candidates.

```sql
CREATE TABLE ad_creative_requests (
  id TEXT PRIMARY KEY,
  website_ad_job_id TEXT NOT NULL,
  FOREIGN KEY (website_ad_job_id) REFERENCES website_ad_jobs(id),

  -- Request attempt number (1 = initial, 2-3 = regenerations)
  attempt_number INTEGER NOT NULL,

  -- Extracted article analysis
  primary_themes TEXT,                     -- JSON array of theme strings
  mood TEXT,
  time_period TEXT,
  location TEXT,
  visual_elements TEXT,                    -- JSON array
  semantic_summary TEXT,

  -- Generated creatives (3 per attempt)
  creative_1_prompt TEXT NOT NULL,
  creative_1_image_artifact_id TEXT,       -- FK to ad_artifacts

  creative_2_prompt TEXT NOT NULL,
  creative_2_image_artifact_id TEXT,       -- FK to ad_artifacts

  creative_3_prompt TEXT NOT NULL,
  creative_3_image_artifact_id TEXT,       -- FK to ad_artifacts

  -- Timestamps
  created_at TEXT NOT NULL,
  completed_at TEXT
);

CREATE INDEX idx_ad_creative_requests_job ON ad_creative_requests(website_ad_job_id);
CREATE INDEX idx_ad_creative_requests_attempt ON ad_creative_requests(attempt_number);
```

**Notes:**
- Each attempt (initial + regenerations) creates a new row
- Stores JSON for flexible theme data
- Image references use artifact foreign keys
- Allows auditing of all generated candidates

---

### 3.3 `ad_variants`

Final exported banner variants (square, vertical, icon).

**Purpose:** Store metadata and storage references for all rendered format variants.

```sql
CREATE TABLE ad_variants (
  id TEXT PRIMARY KEY,
  website_ad_job_id TEXT NOT NULL,
  FOREIGN KEY (website_ad_job_id) REFERENCES website_ad_jobs(id),

  -- Format type
  format TEXT NOT NULL,                    -- 'square', 'vertical', or 'icon'
  dimensions TEXT NOT NULL,                -- '1200x628', '300x600', '256x256'

  -- Storage reference
  artifact_id TEXT NOT NULL,
  FOREIGN KEY (artifact_id) REFERENCES ad_artifacts(id),

  -- Timestamps
  created_at TEXT NOT NULL,
  exported_at TEXT
};

CREATE INDEX idx_ad_variants_job ON ad_variants(website_ad_job_id);
CREATE INDEX idx_ad_variants_format ON ad_variants(format);
```

**Notes:**
- One row per format variant per job
- `artifact_id` references storage location
- `dimensions` is metadata (actual dimensions enforced at render time)

---

### 3.4 `ad_artifacts`

References to stored banner images (candidates and finals).

**Purpose:** Track all generated images and their storage locations.

```sql
CREATE TABLE ad_artifacts (
  id TEXT PRIMARY KEY,
  -- Reference back to job (nullable if artifact is from creative request)
  website_ad_job_id TEXT,
  FOREIGN KEY (website_ad_job_id) REFERENCES website_ad_jobs(id),

  -- Storage type
  artifact_type TEXT NOT NULL,             -- 'creative_candidate', 'variant'
  artifact_source TEXT NOT NULL,           -- 'local_storage', 'cloud_blob'

  -- Storage location
  local_path TEXT,                         -- e.g., 'tmp/website_ads/wad_001_creative_1.png'
  cloud_blob_uri TEXT,                     -- e.g., 'https://..../wad_001_creative_1.png'

  -- File metadata
  file_size INTEGER,                       -- bytes
  file_hash TEXT,                          -- SHA256 for integrity check
  content_type TEXT NOT NULL,              -- 'image/png'

  -- Lifecycle
  created_at TEXT NOT NULL,
  expires_at TEXT,                         -- optional, for temp artifacts
  deleted_at TEXT                          -- soft delete timestamp
};

CREATE INDEX idx_ad_artifacts_job ON ad_artifacts(website_ad_job_id);
CREATE INDEX idx_ad_artifacts_type ON ad_artifacts(artifact_type);
CREATE INDEX idx_ad_artifacts_expires ON ad_artifacts(expires_at);
```

**Notes:**
- Supports both local and cloud storage
- `file_hash` for integrity verification
- `expires_at` for cleanup of temporary artifacts
- `deleted_at` for soft deletion

---

## 4. Schema Changes to Existing Tables

### No changes to existing tables

Phase 5 does NOT modify:
- `products`
- `campaigns`
- `jobs`
- `jobs_logs`
- `scenes`
- `slots`
- `previews`

All Phase 5 data is isolated in new `ad_*` tables.

---

## 5. Data Relationships Diagram

```
┌─────────────────────────┐
│      products           │
│ (existing table)        │
│ - id (PK)               │
└───────────┬─────────────┘
            │ FK
            │
┌───────────▼─────────────┐
│  website_ad_jobs        │
│ - id (PK)               │
│ - product_id (FK)       │
│ - status                │
│ - created_at            │
└───────────┬─────────────┘
            │
      ┌─────┴──────────────────────┐
      │                            │
┌─────▼──────────────────┐  ┌─────▼──────────────────┐
│ ad_creative_requests   │  │   ad_variants         │
│ - id (PK)              │  │ - id (PK)             │
│ - website_ad_job_id FK │  │ - website_ad_job_id FK│
│ - attempt_number       │  │ - format              │
│ - creative_*_prompt    │  │ - artifact_id (FK)    │
│ - creative_*_artifact  │  └─────┬──────────────────┘
└─────┬──────────────────┘        │
      │                            │
      └─────────────┬──────────────┘
                    │ FK
            ┌───────▼──────────────┐
            │   ad_artifacts       │
            │ - id (PK)            │
            │ - artifact_type      │
            │ - local_path         │
            │ - cloud_blob_uri     │
            │ - created_at         │
            └──────────────────────┘
```

---

## 6. Migration Script

### Version: 003_website_ads_schema.sql

```sql
-- Create website_ad_jobs table
CREATE TABLE website_ad_jobs (
  id TEXT PRIMARY KEY,
  product_id TEXT NOT NULL,
  article_source TEXT NOT NULL,
  article_url TEXT,
  article_title TEXT,
  article_text_hash TEXT NOT NULL,
  article_text_preview TEXT NOT NULL,
  brand_voice TEXT,
  status TEXT NOT NULL,
  error_code TEXT,
  error_message TEXT,
  selected_creative_id INTEGER,
  regeneration_count INTEGER NOT NULL DEFAULT 0,
  max_regenerations INTEGER NOT NULL DEFAULT 2,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT,
  FOREIGN KEY (product_id) REFERENCES products(id)
);

CREATE INDEX idx_website_ad_jobs_status ON website_ad_jobs(status);
CREATE INDEX idx_website_ad_jobs_product_id ON website_ad_jobs(product_id);
CREATE INDEX idx_website_ad_jobs_created_at ON website_ad_jobs(created_at DESC);

-- Create ad_creative_requests table
CREATE TABLE ad_creative_requests (
  id TEXT PRIMARY KEY,
  website_ad_job_id TEXT NOT NULL,
  attempt_number INTEGER NOT NULL,
  primary_themes TEXT,
  mood TEXT,
  time_period TEXT,
  location TEXT,
  visual_elements TEXT,
  semantic_summary TEXT,
  creative_1_prompt TEXT NOT NULL,
  creative_1_image_artifact_id TEXT,
  creative_2_prompt TEXT NOT NULL,
  creative_2_image_artifact_id TEXT,
  creative_3_prompt TEXT NOT NULL,
  creative_3_image_artifact_id TEXT,
  created_at TEXT NOT NULL,
  completed_at TEXT,
  FOREIGN KEY (website_ad_job_id) REFERENCES website_ad_jobs(id)
);

CREATE INDEX idx_ad_creative_requests_job ON ad_creative_requests(website_ad_job_id);
CREATE INDEX idx_ad_creative_requests_attempt ON ad_creative_requests(attempt_number);

-- Create ad_variants table
CREATE TABLE ad_variants (
  id TEXT PRIMARY KEY,
  website_ad_job_id TEXT NOT NULL,
  format TEXT NOT NULL,
  dimensions TEXT NOT NULL,
  artifact_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  exported_at TEXT,
  FOREIGN KEY (website_ad_job_id) REFERENCES website_ad_jobs(id),
  FOREIGN KEY (artifact_id) REFERENCES ad_artifacts(id)
);

CREATE INDEX idx_ad_variants_job ON ad_variants(website_ad_job_id);
CREATE INDEX idx_ad_variants_format ON ad_variants(format);

-- Create ad_artifacts table
CREATE TABLE ad_artifacts (
  id TEXT PRIMARY KEY,
  website_ad_job_id TEXT,
  artifact_type TEXT NOT NULL,
  artifact_source TEXT NOT NULL,
  local_path TEXT,
  cloud_blob_uri TEXT,
  file_size INTEGER,
  file_hash TEXT,
  content_type TEXT NOT NULL,
  created_at TEXT NOT NULL,
  expires_at TEXT,
  deleted_at TEXT,
  FOREIGN KEY (website_ad_job_id) REFERENCES website_ad_jobs(id)
);

CREATE INDEX idx_ad_artifacts_job ON ad_artifacts(website_ad_job_id);
CREATE INDEX idx_ad_artifacts_type ON ad_artifacts(artifact_type);
CREATE INDEX idx_ad_artifacts_expires ON ad_artifacts(expires_at);

-- Add website_ad_jobs column to existing jobs table (optional, for unified monitoring)
-- ALTER TABLE jobs ADD COLUMN related_website_ad_job_id TEXT;

-- Update schema_version or migration tracking
INSERT INTO migrations (name, applied_at) VALUES ('003_website_ads_schema.sql', datetime('now'));
```

---

## 7. Query Examples

### List all completed website ad jobs

```sql
SELECT 
  j.id,
  j.product_id,
  j.selected_creative_id,
  COUNT(v.id) AS variant_count,
  j.created_at
FROM website_ad_jobs j
LEFT JOIN ad_variants v ON j.id = v.website_ad_job_id
WHERE j.status = 'completed'
GROUP BY j.id
ORDER BY j.created_at DESC;
```

### Get creative candidates for a job

```sql
SELECT 
  r.id,
  r.creative_1_prompt,
  r.creative_1_image_artifact_id,
  r.creative_2_prompt,
  r.creative_2_image_artifact_id,
  r.creative_3_prompt,
  r.creative_3_image_artifact_id
FROM ad_creative_requests r
WHERE r.website_ad_job_id = 'wad_001'
ORDER BY r.attempt_number DESC
LIMIT 1;
```

### Get all exported variants for a job

```sql
SELECT 
  v.format,
  v.dimensions,
  a.local_path,
  a.cloud_blob_uri,
  a.file_size
FROM ad_variants v
JOIN ad_artifacts a ON v.artifact_id = a.id
WHERE v.website_ad_job_id = 'wad_001'
ORDER BY v.format;
```

### Find stale artifacts for cleanup

```sql
SELECT id, local_path, cloud_blob_uri
FROM ad_artifacts
WHERE expires_at < datetime('now')
  AND deleted_at IS NULL;
```

---

## 8. Constraints and Considerations

### Data Integrity

- All timestamps in UTC ISO 8601 format (`2026-03-14T10:00:00Z`)
- `id` fields use consistent prefix naming (`wad_*`, `acr_*`, `adv_*`, `ada_*`)
- Foreign key constraints enforced (SQLite with `PRAGMA foreign_keys = ON`)

### Performance

- Indexes on frequently queried columns (status, product_id, created_at)
- Consider partitioning `ad_artifacts` by date if it grows large

### Storage

- `article_text_preview` (500 chars) suitable for UI display
- Full article text stored in cloud during processing (not in DB)
- PNG binaries stored in filesystem, not database

### Cleanup

- Implement periodic cleanup of `ad_artifacts` where `expires_at` < now
- Soft-delete pattern allows audit trail preservation

---

**Next Step:** See [06_Coding_Standards.md](06_Coding_Standards.md) for code organization guidelines.
