# API Contracts Document (Go Backend)

## Document Purpose
Define all REST API endpoints that the React dashboard and frontend will consume. These contracts specify request/response schemas, error codes, and behavior to enable parallel frontend/backend development.

---

## 1. Authentication & CORS (MVP Simplified)
**Status:** No authentication in MVP (demo/internal only). All endpoints accept requests from React frontend.

```http
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

**API Version:** `/api` (no `/v1` in MVP for simplicity)

---

## 2. Root Endpoint

### GET `/api/health`
**Purpose:** Health check for backend readiness.

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2026-03-12T14:30:00Z",
  "version": "0.1.0-mvp"
}
```

---

## 3. Campaign Endpoints

### POST `/api/campaigns`
**Purpose:** Create a new campaign (video + product selection).

**Content-Type:** `multipart/form-data`

**Form Fields:**
- `name` (string, required): Campaign name
- `product_id` (string, required): Reference to existing product
- `target_ad_duration_seconds` (int, optional): Desired ad length, default 5
- `video_file` (file, required): MP4 video (H.264 codec)

**Response (201 Created):**
```json
{
  "id": "campaign_abc123",
  "name": "Sneaker Ad in Movie Trailer",
  "product_id": "prod_001",
  "video_filename": "campaign_abc123.mp4",
  "duration_seconds": 540,
  "target_ad_duration_seconds": 6,
  "created_at": "2026-03-12T14:30:00Z",
  "job_id": "job_xyz789"
}
```

**Error Responses:**
- `400 Bad Request`: Missing fields, invalid video format
- `413 Payload Too Large`: Video exceeds size limit
- `422 Unprocessable Entity`: Video codec not supported (not H.264 MP4)

---

### GET `/api/campaigns`
**Purpose:** List all campaigns (with pagination).

**Query Parameters:**
- `limit` (int, optional): Limit results, default 20
- `offset` (int, optional): Pagination offset, default 0

**Response (200 OK):**
```json
{
  "campaigns": [
    {
      "id": "campaign_abc123",
      "name": "Sneaker Ad in Movie Trailer",
      "product_id": "prod_001",
      "product_name": "Nike Air Max 270",
      "duration_seconds": 540,
      "target_ad_duration_seconds": 6,
      "created_at": "2026-03-12T14:30:00Z",
      "job_id": "job_xyz789",
      "job_status": "generating"
    }
  ],
  "total": 5,
  "limit": 20,
  "offset": 0
}
```

---

### GET `/api/campaigns/{campaign_id}`
**Purpose:** Get single campaign details.

**Response (200 OK):**
```json
{
  "id": "campaign_abc123",
  "name": "Sneaker Ad in Movie Trailer",
  "product_id": "prod_001",
  "product_name": "Nike Air Max 270",
  "video_filename": "campaign_abc123.mp4",
  "duration_seconds": 540,
  "target_ad_duration_seconds": 6,
  "created_at": "2026-03-12T14:30:00Z",
  "job_id": "job_xyz789"
}
```

---

## 4. Product Endpoints

### GET `/api/products`
**Purpose:** List available products for ad insertion.

**Response (200 OK):**
```json
{
  "products": [
    {
      "id": "prod_001",
      "name": "Nike Air Max 270",
      "category": "sneakers",
      "description": "Premium running shoe",
      "context_keywords": ["outdoor", "sports", "casual"]
    },
    {
      "id": "prod_002",
      "name": "Apple Watch Series 8",
      "category": "wearables",
      "description": "Fitness tracker",
      "context_keywords": ["tech", "health", "lifestyle"]
    }
  ]
}
```

---

## 5. Job Endpoints

### GET `/api/jobs/{job_id}`
**Purpose:** Poll job status (used by React dashboard for progress tracking).

**Response (200 OK):**
```json
{
  "id": "job_xyz789",
  "campaign_id": "campaign_abc123",
  "status": "generating",
  "progress_percent": 65,
  "current_stage": "ad_generation",
  "created_at": "2026-03-12T14:30:00Z",
  "started_at": "2026-03-12T14:31:00Z",
  "estimated_completion_seconds": 120,
  "error_message": null,
  "metadata": {
    "selected_slot_id": "slot_s42",
    "selected_slot_rank": 1,
    "total_scenes": 24,
    "top_3_slots": [
      { "rank": 1, "score": 0.92 },
      { "rank": 2, "score": 0.87 },
      { "rank": 3, "score": 0.82 }
    ]
  }
}
```

**Status Values:**
- `queued`: Waiting to start analysis
- `analyzing`: Scene detection + context analysis running
- `generating`: RIFE frame interpolation in progress
- `stitching`: ffmpeg stitching final output
- `completed`: Success
- `failed`: Error occurred (check `error_message`)

---

### GET `/api/jobs/{job_id}/logs`
**Purpose:** Retrieve event log for debugging/transparency.

**Query Parameters:**
- `limit` (int, optional): Max log entries, default 100

**Response (200 OK):**
```json
{
  "job_id": "job_xyz789",
  "logs": [
    {
      "timestamp": "2026-03-12T14:31:00Z",
      "event_type": "stage_started",
      "stage_name": "scene_detection",
      "message": "Starting scene detection with OpenCV",
      "details": {
        "video_path": "/tmp/uploads/campaign_abc123.mp4",
        "duration_frames": 12960
      }
    },
    {
      "timestamp": "2026-03-12T14:32:15Z",
      "event_type": "stage_completed",
      "stage_name": "scene_detection",
      "message": "Detected 24 scenes",
      "details": {
        "scenes_detected": 24,
        "avg_scene_duration": 15.3
      }
    }
  ]
}
```

---

## 6. Scene Endpoints

### GET `/api/jobs/{job_id}/scenes`
**Purpose:** Retrieve detected scenes for analysis dashboard.

**Query Parameters:**
- `limit` (int, optional): Max results
- `offset` (int, optional): Pagination offset

**Response (200 OK):**
```json
{
  "job_id": "job_xyz789",
  "scenes": [
    {
      "id": "scene_001",
      "scene_number": 1,
      "start_frame": 0,
      "end_frame": 300,
      "duration_seconds": 12.5,
      "motion_score": 0.7,
      "stability_score": 0.6,
      "dialogue_present": true,
      "dialogue_gap_start_frame": 50,
      "dialogue_gap_end_frame": 80,
      "scene_description": "Outdoor park scene with people walking"
    }
  ],
  "total": 24
}
```

---

## 7. Slot Endpoints

### GET `/api/jobs/{job_id}/slots`
**Purpose:** Retrieve ranked candidate ad insertion slots for dashboard display.

**Query Parameters:**
- `limit` (int, optional): Max slots, default 5

**Response (200 OK):**
```json
{
  "job_id": "job_xyz789",
  "campaign_id": "campaign_abc123",
  "product_name": "Nike Air Max 270",
  "slots": [
    {
      "id": "slot_s42",
      "rank": 1,
      "scene_id": "scene_012",
      "scene_number": 12,
      "insertion_frame": 5600,
      "slot_type": "dialogue_gap",
      "confidence": 0.98,
      "score": 0.92,
      "reasoning": "14-second dialogue gap in outdoor park scene. Low motion (0.4), high stability (0.8). Context matches sneaker product. No narrative disruption.",
      "selected": false,
      "generated_video_path": null,
      "generation_status": null
    },
    {
      "id": "slot_s38",
      "rank": 2,
      "scene_id": "scene_010",
      "scene_number": 10,
      "insertion_frame": 4800,
      "slot_type": "scene_boundary",
      "confidence": 0.87,
      "score": 0.87,
      "reasoning": "Scene transition point. Smooth cut. 8-second buffer.",
      "selected": false,
      "generated_video_path": null,
      "generation_status": null
    },
    {
      "id": "slot_s35",
      "rank": 3,
      "scene_id": "scene_008",
      "scene_number": 8,
      "insertion_frame": 4000,
      "slot_type": "low_motion",
      "confidence": 0.82,
      "score": 0.82,
      "reasoning": "Stationary shot. Perfect for product placement.",
      "selected": false,
      "generated_video_path": null,
      "generation_status": null
    }
  ]
}
```

---

### GET `/api/jobs/{job_id}/slots/{slot_id}`
**Purpose:** Get single slot details with generation progress.

**Response (200 OK):**
```json
{
  "id": "slot_s42",
  "rank": 1,
  "insertion_frame": 5600,
  "confidence": 0.98,
  "score": 0.92,
  "reasoning": "...",
  "selected": true,
  "generation_status": "in_progress",
  "generated_video_path": null,
  "metadata": {
    "replicate_prediction_id": "pred_abc123def456",
    "replicate_model": "deforum-research/rife",
    "started_at": "2026-03-12T14:40:00Z"
  }
}
```

---

### POST `/api/jobs/{job_id}/slots/{slot_id}/select`
**Purpose:** User selects a slot → triggers RIFE ad generation on Replicate.

**Request:**
```json
{
  "slot_id": "slot_s42"
}
```

**Response (202 Accepted - Async):**
```json
{
  "id": "slot_s42",
  "rank": 1,
  "generation_status": "pending",
  "message": "Ad generation queued on Replicate",
  "metadata": {
    "replicate_prediction_id": "pred_abc123def456",
    "replicate_model": "deforum-research/rife",
    "start_frame_path": "/tmp/frames/slot_s42_start.png",
    "end_frame_path": "/tmp/frames/slot_s42_end.png"
  }
}
```

**Next Step:**
- Frontend polls `GET /api/jobs/{job_id}/slots/{slot_id}` every 5-10 seconds
- Once `generation_status == "success"`, `generated_video_path` is populated
- User can then trigger rendering

---

## 8. Render Endpoints

### POST `/api/jobs/{job_id}/render`
**Purpose:** Stitch selected slot's generated ad into original video using ffmpeg.

**Request:**
```json
{
  "slot_id": "slot_s42"
}
```

**Response (202 Accepted - Async):**
```json
{
  "id": "render_xyz",
  "job_id": "job_xyz789",
  "slot_id": "slot_s42",
  "status": "in_progress",
  "output_video_path": null,
  "message": "Stitching started with ffmpeg",
  "started_at": "2026-03-12T14:43:00Z"
}
```

---

### GET `/api/jobs/{job_id}/renders`
**Purpose:** Get list of rendered outputs for a job.

**Response (200 OK):**
```json
{
  "job_id": "job_xyz789",
  "renders": [
    {
      "id": "render_xyz",
      "slot_id": "slot_s42",
      "status": "completed",
      "output_video_path": "/tmp/outputs/render_xyz_final.mp4",
      "metrics": {
        "total_frames": 12960,
        "insertion_start_frame": 5600,
        "insertion_duration_seconds": 6.5,
        "stitching_quality_score": 0.95
      },
      "created_at": "2026-03-12T14:42:00Z",
      "completed_at": "2026-03-12T14:44:30Z"
    }
  ]
}
```

---

### GET `/api/jobs/{job_id}/renders/{render_id}/download`
**Purpose:** Download final stitched video for preview.

**Response (200 OK):**
- Content-Type: `video/mp4`
- Content-Disposition: `attachment; filename="render_xyz_final.mp4"`
- Binary MP4 file stream

**Note:** For MVP, serves from local filesystem. Production will use signed URLs.

---

## 9. Error Response Format

All errors follow this standard:

```json
{
  "error": "Slot generation failed",
  "error_code": "GENERATION_TIMEOUT",
  "http_status": 504,
  "details": {
    "replicate_prediction_id": "pred_abc123",
    "replicate_error": "Prediction timed out after 600 seconds",
    "recommendation": "Retry with a different slot or check Replicate status"
  },
  "timestamp": "2026-03-12T14:45:00Z"
}
```

**Common HTTP Status Codes:**
- `200 OK`: Successful GET request
- `201 Created`: Resource created (POST)
- `202 Accepted`: Async operation queued
- `400 Bad Request`: Invalid input
- `404 Not Found`: Resource not found
- `413 Payload Too Large`: File upload too big
- `422 Unprocessable Entity`: Video format/codec invalid
- `500 Internal Server Error`: Unexpected backend error
- `504 Gateway Timeout`: Replicate inference timeout

**Common Error Codes:**
- `INVALID_REQUEST`: Malformed request
- `RESOURCE_NOT_FOUND`: Job/slot/campaign doesn't exist
- `GENERATION_TIMEOUT`: RIFE inference timeout
- `STITCHING_FAILED`: ffmpeg rendering error
- `STORAGE_ERROR`: File read/write error
- `DATABASE_ERROR`: SQL query failed

---

## 10. Example Frontend Workflow

```
User creates campaign:
1. POST /api/campaigns (multipart upload) → get campaign_id + job_id
2. While job.status != "completed":
   - GET /api/jobs/{job_id}  → poll every 3 seconds
3. GET /api/jobs/{job_id}/slots → display top 3 ranked slots

User selects slot:
4. POST /api/jobs/{job_id}/slots/{slot_id}/select → triggers RIFE
5. While slot.generation_status != "success":
   - GET /api/jobs/{job_id}/slots/{slot_id}  → poll every 5 seconds
6. POST /api/jobs/{job_id}/render → stitch video
7. While render.status != "completed":
   - GET /api/jobs/{job_id}/renders → check progress
8. GET /api/jobs/{job_id}/renders/{render_id}/download → preview final output
```

---

## 11. Implementation Notes for Go Developer

**Routing Library:** `github.com/gorilla/mux` (simple, sufficient for MVP)

**Key Handlers:**
```go
router.HandleFunc("/api/health", getHealth).Methods("GET")
router.HandleFunc("/api/campaigns", createCampaign).Methods("POST")
router.HandleFunc("/api/campaigns", listCampaigns).Methods("GET")
router.HandleFunc("/api/campaigns/{id}", getCampaign).Methods("GET")
router.HandleFunc("/api/jobs/{id}", getJob).Methods("GET")
router.HandleFunc("/api/jobs/{id}/logs", getJobLogs).Methods("GET")
router.HandleFunc("/api/jobs/{id}/slots", listSlots).Methods("GET")
router.HandleFunc("/api/jobs/{id}/slots/{slotId}", getSlot).Methods("GET")
router.HandleFunc("/api/jobs/{id}/slots/{slotId}/select", selectSlot).Methods("POST")
router.HandleFunc("/api/jobs/{id}/render", renderVideo).Methods("POST")
router.HandleFunc("/api/jobs/{id}/renders", listRenders).Methods("GET")
router.HandleFunc("/api/jobs/{id}/renders/{renderId}/download", downloadRender).Methods("GET")
```

**Async Job Handling:**
- When campaign is created, insert job with status `queued`
- Spawn goroutine to process job asynchronously
- Return campaign/job info immediately (202 for long-running ops)
- Frontend polls status endpoint until completion

**File Upload:**
```go
// In createCampaign handler:
file, header, err := r.FormFile("video_file")
newPath := filepath.Join("/tmp/uploads", uniqueFilename)
// Save to local filesystem
```

**Error Response Template:**
```go
type ErrorResponse struct {
  Error       string                 `json:"error"`
  ErrorCode   string                 `json:"error_code"`
  HTTPStatus  int                    `json:"http_status"`
  Details     map[string]interface{} `json:"details"`
  Timestamp   string                 `json:"timestamp"`
}
```

---

## 12. Rate Limiting & Auth (Future)
- MVP: No rate limiting, no authentication
- Production: Add auth middleware + rate limiting per IP/API key


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
