# Data Schema Definitions

## 1. Overview
This document defines the core persistence model and schema contracts for the system.

## 2. Relational Schema (SQL-Oriented)

### Table: campaigns
```sql
CREATE TABLE campaigns (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    max_ad_duration_sec NUMERIC NOT NULL DEFAULT 10,
    preferred_strategy TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Table: products
```sql
CREATE TABLE products (
    id UUID PRIMARY KEY,
    campaign_id UUID NOT NULL REFERENCES campaigns(id),
    name TEXT NOT NULL,
    description TEXT,
    primary_asset_url TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Table: jobs
```sql
CREATE TABLE jobs (
    id UUID PRIMARY KEY,
    campaign_id UUID REFERENCES campaigns(id),
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    source_video_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Table: assets
```sql
CREATE TABLE assets (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id),
    role TEXT NOT NULL,
    filename TEXT NOT NULL,
    object_key TEXT NOT NULL,
    content_type TEXT NOT NULL,
    status TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Table: scenes
```sql
CREATE TABLE scenes (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id),
    scene_index INT NOT NULL,
    start_time_sec NUMERIC NOT NULL,
    end_time_sec NUMERIC NOT NULL,
    motion_score NUMERIC,
    stability_score NUMERIC,
    transcript_summary TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Table: candidate_slots
```sql
CREATE TABLE candidate_slots (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id),
    scene_id UUID REFERENCES scenes(id),
    start_time_sec NUMERIC NOT NULL,
    end_time_sec NUMERIC NOT NULL,
    score NUMERIC NOT NULL,
    reasons JSONB,
    ranking_metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Table: insertion_plans
```sql
CREATE TABLE insertion_plans (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id),
    slot_id UUID NOT NULL REFERENCES candidate_slots(id),
    product_id UUID NOT NULL REFERENCES products(id),
    strategy TEXT NOT NULL,
    anchor_start_frame_url TEXT,
    anchor_end_frame_url TEXT,
    prompt_payload JSONB,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Table: generated_clips
```sql
CREATE TABLE generated_clips (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id),
    plan_id UUID NOT NULL REFERENCES insertion_plans(id),
    clip_url TEXT NOT NULL,
    duration_sec NUMERIC,
    strategy TEXT NOT NULL,
    generation_metadata JSONB,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Table: outputs
```sql
CREATE TABLE outputs (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id),
    clip_id UUID REFERENCES generated_clips(id),
    output_type TEXT NOT NULL,
    output_url TEXT NOT NULL,
    render_metadata JSONB,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## 3. JSON Schema Definitions

### Candidate Slot Schema
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "CandidateSlot",
  "type": "object",
  "required": ["slot_id", "start_time_sec", "end_time_sec", "score", "reasons"],
  "properties": {
    "slot_id": { "type": "string" },
    "scene_id": { "type": "string" },
    "start_time_sec": { "type": "number" },
    "end_time_sec": { "type": "number" },
    "score": { "type": "number" },
    "reasons": {
      "type": "array",
      "items": { "type": "string" }
    }
  }
}
```

### Insertion Plan Schema
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "InsertionPlan",
  "type": "object",
  "required": ["plan_id", "slot_id", "product_id", "strategy", "status"],
  "properties": {
    "plan_id": { "type": "string" },
    "slot_id": { "type": "string" },
    "product_id": { "type": "string" },
    "strategy": {
      "type": "string",
      "enum": ["environment_insert", "micro_bridge", "character_interaction"]
    },
    "status": { "type": "string" },
    "prompt_payload": { "type": "object" }
  }
}
```

## 4. Protobuf Example
```proto
syntax = "proto3";

package dynamicads;

message CandidateSlot {
  string slot_id = 1;
  string scene_id = 2;
  double start_time_sec = 3;
  double end_time_sec = 4;
  double score = 5;
  repeated string reasons = 6;
}

message GenerationRequest {
  string job_id = 1;
  string plan_id = 2;
  string strategy = 3;
}

message GenerationResult {
  string clip_id = 1;
  string clip_url = 2;
  double duration_sec = 3;
  string status = 4;
}
```

## 5. Data Modeling Rules
- all primary business entities use UUIDs
- status fields must be enum-constrained at application layer
- metadata that evolves quickly belongs in JSONB unless it needs relational querying
- large binary/video assets never belong in the database directly
- DB stores references, not heavy media payloads

## 6. Suggested Enums
### Job Status
- `created`
- `assets_registered`
- `analysis_queued`
- `scene_detection_complete`
- `context_analysis_complete`
- `slot_ranking_complete`
- `generation_queued`
- `generation_complete`
- `render_queued`
- `render_complete`
- `failed`

### Strategy Type
- `environment_insert`
- `micro_bridge`
- `character_interaction`

### Output Type
- `preview`
- `final`

## 7. Indexing Recommendations
- index `jobs(status)`
- index `scenes(job_id, scene_index)`
- index `candidate_slots(job_id, score DESC)`
- index `insertion_plans(job_id, status)`
- index `generated_clips(job_id, status)`
- index `outputs(job_id, output_type)`
