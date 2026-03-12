# API Contracts

## 1. Conventions
- content type: `application/json`
- authentication: bearer token for control APIs
- all responses are JSON
- all error responses use structured error format
- API version prefix: `/api/v1`

## 2. Create Job
### `POST /api/v1/jobs`
Create a new processing job.

#### Request
```json
{
  "title": "Movie Ad Test Job",
  "description": "Demo run for contextual ad insertion",
  "campaign_id": "camp_123"
}
```

#### Response
```json
{
  "job_id": "job_123",
  "status": "created"
}
```

---

## 3. Request Upload URLs
### `POST /api/v1/assets/upload-urls`
Request signed upload URLs for source video and product assets.

#### Request
```json
{
  "job_id": "job_123",
  "files": [
    {
      "filename": "movie_clip.mp4",
      "content_type": "video/mp4",
      "role": "source_video"
    },
    {
      "filename": "product.png",
      "content_type": "image/png",
      "role": "product_asset"
    }
  ]
}
```

#### Response
```json
{
  "uploads": [
    {
      "file_id": "file_1",
      "filename": "movie_clip.mp4",
      "upload_url": "https://...",
      "object_key": "jobs/job_123/source/movie_clip.mp4"
    },
    {
      "file_id": "file_2",
      "filename": "product.png",
      "upload_url": "https://...",
      "object_key": "jobs/job_123/assets/product.png"
    }
  ]
}
```

---

## 4. Finalize Asset Registration
### `POST /api/v1/assets/finalize`
Confirm uploaded assets so processing can begin.

#### Request
```json
{
  "job_id": "job_123",
  "files": [
    {
      "file_id": "file_1",
      "role": "source_video",
      "object_key": "jobs/job_123/source/movie_clip.mp4"
    },
    {
      "file_id": "file_2",
      "role": "product_asset",
      "object_key": "jobs/job_123/assets/product.png"
    }
  ]
}
```

#### Response
```json
{
  "job_id": "job_123",
  "status": "assets_registered"
}
```

---

## 5. Start Analysis Pipeline
### `POST /api/v1/jobs/{job_id}/start-analysis`
Trigger the analysis workflow.

#### Response
```json
{
  "job_id": "job_123",
  "status": "analysis_queued"
}
```

---

## 6. Get Job Status
### `GET /api/v1/jobs/{job_id}`
Retrieve job status and stage summaries.

#### Response
```json
{
  "job_id": "job_123",
  "status": "slot_ranking_complete",
  "stages": {
    "scene_detection": "complete",
    "context_analysis": "complete",
    "slot_ranking": "complete",
    "generation": "pending",
    "rendering": "pending"
  }
}
```

---

## 7. List Candidate Slots
### `GET /api/v1/jobs/{job_id}/slots`
Return ranked insertion candidates.

#### Response
```json
{
  "job_id": "job_123",
  "slots": [
    {
      "slot_id": "slot_1",
      "scene_id": "scene_10",
      "start_time_sec": 125.4,
      "end_time_sec": 131.0,
      "score": 0.92,
      "reasons": [
        "low motion",
        "dialogue pause",
        "stable framing"
      ]
    }
  ]
}
```

---

## 8. Generate Ad Plan
### `POST /api/v1/jobs/{job_id}/generate-plan`
Create a plan for a selected candidate slot and product.

#### Request
```json
{
  "slot_id": "slot_1",
  "product_id": "prod_123",
  "strategy": "environment_insert"
}
```

#### Response
```json
{
  "plan_id": "plan_1",
  "status": "planned",
  "strategy": "environment_insert"
}
```

---

## 9. Generate Ad Clip
### `POST /api/v1/jobs/{job_id}/generate-ad`
Trigger generation or composition of ad clip.

#### Request
```json
{
  "plan_id": "plan_1"
}
```

#### Response
```json
{
  "job_id": "job_123",
  "plan_id": "plan_1",
  "status": "generation_queued"
}
```

---

## 10. Get Generated Clip
### `GET /api/v1/jobs/{job_id}/generated-clips/{clip_id}`
Retrieve generated clip metadata.

#### Response
```json
{
  "clip_id": "clip_1",
  "status": "complete",
  "clip_url": "https://...",
  "duration_sec": 6.2,
  "strategy": "environment_insert"
}
```

---

## 11. Render Final Output
### `POST /api/v1/jobs/{job_id}/render`
Insert the ad clip and produce stitched output.

#### Request
```json
{
  "slot_id": "slot_1",
  "clip_id": "clip_1"
}
```

#### Response
```json
{
  "job_id": "job_123",
  "status": "render_queued"
}
```

---

## 12. Get Render Output
### `GET /api/v1/jobs/{job_id}/outputs`
Return available output assets.

#### Response
```json
{
  "job_id": "job_123",
  "outputs": [
    {
      "output_id": "out_1",
      "type": "preview",
      "url": "https://..."
    },
    {
      "output_id": "out_2",
      "type": "final",
      "url": "https://..."
    }
  ]
}
```

---

## 13. Error Response Contract
### Example
```json
{
  "error": {
    "code": "SLOT_NOT_FOUND",
    "message": "Requested slot was not found",
    "details": {
      "job_id": "job_123",
      "slot_id": "slot_missing"
    }
  }
}
```

## 14. Event Contracts (Internal)
### `scene_detection.completed`
```json
{
  "job_id": "job_123",
  "scene_count": 47,
  "status": "complete"
}
```

### `slot_ranking.completed`
```json
{
  "job_id": "job_123",
  "top_slot_ids": ["slot_1", "slot_2", "slot_3"],
  "status": "complete"
}
```

### `generation.failed`
```json
{
  "job_id": "job_123",
  "plan_id": "plan_1",
  "reason": "model_timeout",
  "fallback_recommended": true
}
```
