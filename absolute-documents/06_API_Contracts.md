# API Contracts

## 1. Purpose

Define the MVP REST API consumed by the React dashboard.

## 2. Conventions

- base path: `/api`
- auth: none in MVP
- response format: JSON unless downloading the preview MP4
- JSON naming style: snake_case
- standard API responses do not expose provider request IDs
- local filesystem paths may be returned intentionally for MVP debugging

## 3. Health

### GET `/api/health`

Response:

```json
{
  "status": "healthy",
  "timestamp": "2026-03-12T21:00:00Z",
  "version": "0.1.0-mvp",
  "provider_profile": "azure"
}
```

## 4. Products

### POST `/api/products`

Create a reusable product entry.

Content-Type:

- `multipart/form-data`

Form fields:

- `name` required
- `description` optional
- `category` optional
- `context_keywords` optional comma-separated string
- `source_url` optional retailer or product page URL
- `image_file` optional PNG or JPG

Validation:

- at least one of `image_file` or `source_url` must be present

Response:

```json
{
  "id": "prod_001",
  "name": "sparkling water",
  "description": "light citrus sparkling water",
  "category": "beverage",
  "context_keywords": ["drink", "water", "refreshment"],
  "source_url": "https://example.com/products/sparkling-water",
  "image_path": "tmp/uploads/products/prod_001.png",
  "created_at": "2026-03-12T21:00:00Z"
}
```

### GET `/api/products`

Response:

```json
{
  "products": [
    {
      "id": "prod_001",
      "name": "sparkling water",
      "description": "light citrus sparkling water",
      "category": "beverage",
      "context_keywords": ["drink", "water", "refreshment"],
      "source_url": "https://example.com/products/sparkling-water",
      "image_path": "tmp/uploads/products/prod_001.png",
      "created_at": "2026-03-12T21:00:00Z"
    }
  ]
}
```

## 5. Campaigns

### POST `/api/campaigns`

Create a campaign and upload the source video. This endpoint does not start analysis.

Content-Type:

- `multipart/form-data`

Required form fields:

- `name`
- `video_file`

Optional campaign fields:

- `target_ad_duration_seconds` default `6`

Product attachment options:

- provide `product_id`
- or create a product inline with:
  - `product_name`
  - `product_description`
  - `product_category`
  - `product_context_keywords`
  - `product_source_url`
  - `product_image_file`

Video validation:

- file must be H.264 MP4
- duration must be between 40 and 60 seconds for the baseline validation profile or between 10 and 20 minutes for the full MVP path

Response:

```json
{
  "campaign_id": "camp_001",
  "job_id": "job_001",
  "product_id": "prod_001",
  "name": "sparkling water test",
  "status": "queued",
  "current_stage": "ready_for_analysis",
  "video_filename": "clip.mp4",
  "video_path": "tmp/uploads/campaigns/camp_001.mp4",
  "target_ad_duration_seconds": 6,
  "created_at": "2026-03-12T21:05:00Z"
}
```

### GET `/api/campaigns/{campaign_id}`

Response:

```json
{
  "campaign_id": "camp_001",
  "job_id": "job_001",
  "product_id": "prod_001",
  "name": "sparkling water test",
  "video_filename": "clip.mp4",
  "video_path": "tmp/uploads/campaigns/camp_001.mp4",
  "target_ad_duration_seconds": 6,
  "created_at": "2026-03-12T21:05:00Z"
}
```

## 6. Analysis Start

### POST `/api/jobs/{job_id}/start-analysis`

Explicitly start the analysis pipeline.

Response:

```json
{
  "job_id": "job_001",
  "status": "analyzing",
  "current_stage": "analysis_submission",
  "message": "analysis started"
}
```

## 7. Jobs

### GET `/api/jobs/{job_id}`

Response:

```json
{
  "id": "job_001",
  "campaign_id": "camp_001",
  "status": "analyzing",
  "current_stage": "analysis_poll",
  "progress_percent": 20,
  "selected_slot_id": null,
  "error_message": null,
  "error_code": null,
  "created_at": "2026-03-12T21:05:00Z",
  "started_at": "2026-03-12T21:05:10Z",
  "completed_at": null,
  "metadata": {
    "source_fps": 23.976,
    "duration_seconds": 812.5,
    "content_language": "en",
    "repick_count": 0,
    "rejected_slot_ids": [],
    "top_slot_ids": [],
    "generation_provider_used": "higgsfield",
    "generation_fallback_used": false
  }
}
```

Allowed `status` values:

- `queued`
- `analyzing`
- `generating`
- `stitching`
- `completed`
- `failed`

### GET `/api/jobs/{job_id}/logs`

Response:

```json
{
  "job_id": "job_001",
  "logs": [
    {
      "timestamp": "2026-03-12T21:05:10Z",
      "event_type": "stage_started",
      "stage_name": "analysis_submission",
      "message": "submitted video for cloud analysis"
    }
  ]
}
```

### POST `/api/jobs/{job_id}/slots/manual-select`

Manually create and select a slot using operator-entered seconds after analysis is complete.

Request:

```json
{
  "start_seconds": 123.45,
  "end_seconds": 129.1
}
```

Response:

```json
{
  "job_id": "job_001",
  "slot_id": "slot_job_001_manual_123450_129100",
  "status": "analyzing",
  "current_stage": "line_review",
  "slot_status": "selected",
  "suggested_product_line": "I grabbed this sparkling water on the way here.",
  "message": "manual slot selected and product line prepared"
}
```

## 8. Slots

### GET `/api/jobs/{job_id}/slots`

Return the currently proposed valid candidate slots.

Manual slots may also appear later as selected slots with `metadata.manual = true`.

Response:

```json
{
  "job_id": "job_001",
  "slots": [
    {
      "id": "slot_001",
      "rank": 1,
      "scene_id": "scene_012",
      "anchor_start_frame": 5600,
      "anchor_end_frame": 5601,
      "source_fps": 23.976,
      "quiet_window_seconds": 4.2,
      "score": 0.91,
      "reasoning": "low motion, 4.2-second quiet window, strong beverage context match, stable continuity anchors",
      "status": "proposed",
      "metadata": {
        "manual": false
      },
      "generated_clip_path": null,
      "generation_error": null
    }
  ]
}
```

### GET `/api/jobs/{job_id}/slots/{slot_id}`

Response:

```json
{
  "id": "slot_001",
  "job_id": "job_001",
  "rank": 1,
  "scene_id": "scene_012",
  "anchor_start_frame": 5600,
  "anchor_end_frame": 5601,
  "source_fps": 23.976,
  "quiet_window_seconds": 4.2,
  "score": 0.91,
  "reasoning": "low motion, 4.2-second quiet window, strong beverage context match, stable continuity anchors",
  "status": "selected",
  "metadata": {
    "manual": true,
    "manual_start_seconds": 123.45,
    "manual_end_seconds": 129.1,
    "generation_provider_used": "higgsfield",
    "generation_fallback_used": false
  },
  "suggested_product_line": "I grabbed this sparkling water on the way here.",
  "final_product_line": null,
  "product_line_mode": null,
  "generated_clip_path": null,
  "generation_error": null
}
```

### POST `/api/jobs/{job_id}/slots/{slot_id}/select`

Select a slot and generate the suggested product line for review.

Response:

```json
{
  "job_id": "job_001",
  "slot_id": "slot_001",
  "status": "analyzing",
  "current_stage": "line_review",
  "slot_status": "selected",
  "suggested_product_line": "I grabbed this sparkling water on the way here.",
  "message": "slot selected and product line prepared"
}
```

### POST `/api/jobs/{job_id}/slots/{slot_id}/reject`

Reject a proposed slot.

Request:

```json
{
  "note": "too close to active dialogue"
}
```

Response:

```json
{
  "job_id": "job_001",
  "slot_id": "slot_001",
  "slot_status": "rejected",
  "message": "slot rejected"
}
```

### POST `/api/jobs/{job_id}/slots/re-pick`

Ask the system to return the next best valid candidates excluding rejected slots.

Response:

```json
{
  "job_id": "job_001",
  "status": "analyzing",
  "current_stage": "slot_selection",
  "message": "re-pick requested"
}
```

### POST `/api/jobs/{job_id}/slots/{slot_id}/generate`

Start CAFAI generation for the selected slot after product line review.

Request:

```json
{
  "product_line_mode": "operator",
  "custom_product_line": "I picked up this sparkling water earlier."
}
```

Allowed `product_line_mode` values:

- `auto`
- `operator`
- `disabled`

Response:

```json
{
  "job_id": "job_001",
  "slot_id": "slot_001",
  "status": "generating",
  "current_stage": "generation_submission",
  "slot_status": "generating",
  "message": "cafai generation started"
}
```

## 9. Preview Rendering

### POST `/api/jobs/{job_id}/preview/render`

Start preview rendering for a generated slot. This same endpoint may be called again after a render failure if the generated artifact still exists.

Request:

```json
{
  "slot_id": "slot_001"
}
```

Response:

```json
{
  "job_id": "job_001",
  "slot_id": "slot_001",
  "status": "stitching",
  "current_stage": "render_submission",
  "message": "preview render started"
}
```

### GET `/api/jobs/{job_id}/preview`

Response:

```json
{
  "job_id": "job_001",
  "slot_id": "slot_001",
  "status": "completed",
  "output_video_path": "tmp/previews/job_001_preview.mp4",
  "download_path": "/api/jobs/job_001/preview/download",
  "duration_seconds": 818.9,
  "created_at": "2026-03-12T21:12:00Z",
  "completed_at": "2026-03-12T21:14:30Z"
}
```

### GET `/api/jobs/{job_id}/preview/download`

Response:

- `200 OK`
- `Content-Type: video/mp4`
- binary MP4 stream

### GET `/api/jobs/{job_id}/preview/stream`

Response:

- `200 OK`
- `Content-Type: video/mp4`
- binary MP4 stream intended for inline playback

## 10. Error Response

Standard shape:

```json
{
  "error": "no suitable slot found",
  "error_code": "NO_SUITABLE_SLOT_FOUND",
  "http_status": 409,
  "details": {
    "job_id": "job_001",
    "current_stage": "slot_selection"
  },
  "timestamp": "2026-03-12T21:13:00Z"
}
```

Common error codes:

- `INVALID_REQUEST`
- `INVALID_VIDEO_CODEC`
- `INVALID_VIDEO_DURATION`
- `PRODUCT_INPUT_MISSING`
- `RESOURCE_NOT_FOUND`
- `NO_SUITABLE_SLOT_FOUND`
- `ANALYSIS_FAILED`
- `GENERATION_FAILED`
- `PREVIEW_RENDER_FAILED`
- `DATABASE_ERROR`
- `STORAGE_ERROR`
