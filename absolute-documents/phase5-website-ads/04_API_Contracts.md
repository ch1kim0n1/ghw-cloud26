# Phase 5: Dynamic Website Ads - API Contracts

## 1. Purpose

Define the complete REST API surface for Phase 5 website ad generation, consumed by the React frontend.

## 2. Conventions

- **Base path:** `/api`
- **Authentication:** None in MVP
- **Response format:** JSON (except image and ZIP streams)
- **JSON naming:** snake_case
- **HTTP methods:** POST (create/action), GET (read), DELETE (not in Phase 5 MVP)
- **Status codes:** Follow HTTP standards
- **Error format:** Consistent error envelope (see section 3)
- **Image/ZIP responses:** Binary streams with appropriate Content-Type

## 3. Standard Error Response

All JSON error responses follow this envelope:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "request_id": "req_xyz123",
  "timestamp": "2026-03-14T10:00:00Z"
}
```

**HTTP Status Codes:**
- `200 OK` — Successful GET/successful state
- `201 Created` — Successful resource creation (POST)
- `202 Accepted` — Job submitted (async processing)
- `400 Bad Request` — Validation error (malformed request, bad article, etc.)
- `404 Not Found` — Resource not found
- `429 Too Many Requests` — Regeneration limit exceeded
- `500 Internal Server Error` — Server error
- `503 Service Unavailable` — Provider unavailable
- `504 Gateway Timeout` — Provider timeout

## 4. Website Ad Job API

### POST `/api/website-ads`

Create a new website ad generation job.

**Purpose:** Submit an article and product for context-aware banner generation.

**Content-Type:** `application/json`

**Request Body:**

```json
{
  "article_source": "url",
  "article_url": "https://en.wikipedia.org/wiki/Colosseum",
  "product_id": "prod_001",
  "brand_voice": "sophisticated"
}
```

Or alternatively with inline product:

```json
{
  "article_source": "text",
  "article_text": "The Colosseum is an iconic amphitheater in Rome...",
  "product_name": "Premium Wireless Headphones",
  "product_description": "Noise-cancelling, 30-hour battery, premium sound",
  "product_image_url": "https://example.com/headphones.png"
}
```

Or with headline + body:

```json
{
  "article_source": "headline_body",
  "article_headline": "The Colosseum: Engineering Marvel of Ancient Rome",
  "article_body": "The Colosseum is an iconic amphitheater...",
  "product_id": "prod_002"
}
```

**Request Fields:**

| Field | Type | Required | Constraints | Example |
|-------|------|----------|-------------|---------|
| `article_source` | string | Yes | One of: "url", "text", "headline_body" | "url" |
| `article_url` | string | Conditional | URL format, reachable | "https://example.com/article" |
| `article_text` | string | Conditional | Max 50,000 chars | "Full article text..." |
| `article_headline` | string | Conditional | Max 200 chars | "Article Title" |
| `article_body` | string | Conditional | Max 49,800 chars | "Article body..." |
| `product_id` | string | Conditional* | Must exist in catalog | "prod_001" |
| `product_name` | string | Conditional* | Max 100 chars | "Wireless Headphones" |
| `product_description` | string | Conditional* | Max 500 chars | "Premium noise-cancelling..." |
| `product_image_url` | string | Optional | URL format | "https://example.com/img.png" |
| `brand_voice` | string | Optional | One of: "premium", "casual", "professional", "playful" | "sophisticated" |

*Either `product_id` OR (`product_name` + `product_description`) required.

**Validation Rules:**

- Exactly one of: `article_url`, `article_text`, or (`article_headline` + `article_body`)
- If `article_url`: must be valid HTTP(S) URL
- If `article_text`: must be > 100 chars and ≤ 50,000 chars
- If `article_headline` + `article_body`: both required, combined ≤ 50,000 chars
- Either `product_id` or (`product_name` + `product_description`)
- If `product_image_url`: must be valid HTTP(S) URL

**Response (201 Created):**

```json
{
  "id": "wad_001",
  "product_id": "prod_001",
  "product_name": "Premium Wireless Headphones",
  "article_source": "url",
  "article_preview": "The Colosseum is an iconic amphitheater in Rome, Italy. Built between 70 and 80 AD...",
  "status": "requested",
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:00:00Z"
}
```

**Error Responses:**

| Status | Error Code | Message | Cause |
|--------|-----------|---------|-------|
| 400 | `invalid_article_source` | Must specify one of: url, text, headline_body | Missing or invalid source type |
| 400 | `invalid_url` | URL must be valid HTTP(S) | Malformed URL |
| 400 | `article_too_short` | Article must be at least 100 characters | Article text < 100 chars |
| 400 | `article_too_long` | Article text exceeds 50,000 character limit | Article text > 50,000 chars |
| 400 | `missing_product_id_or_details` | Either product_id or (product_name + product_description) required | Both missing |
| 400 | `product_not_found` | Product with ID {product_id} not found | Invalid product_id |
| 400 | `invalid_brand_voice` | brand_voice must be: premium, casual, professional, playful | Invalid brand_voice |

---

### GET `/api/website-ads/:job_id`

Fetch website ad job status and creative candidates.

**Purpose:** Poll job progress and retrieve generated candidate banners.

**Response (200 OK - Status: requesting):**

```json
{
  "id": "wad_001",
  "product_id": "prod_001",
  "product_name": "Premium Wireless Headphones",
  "status": "requesting",
  "message": "Submitting article for analysis...",
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:00:10Z"
}
```

**Response (200 OK - Status: analyzing):**

```json
{
  "id": "wad_001",
  "product_id": "prod_001",
  "product_name": "Premium Wireless Headphones",
  "status": "analyzing",
  "message": "Analyzing article themes and generating creative prompts...",
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:00:30Z"
}
```

**Response (200 OK - Status: generating):**

```json
{
  "id": "wad_001",
  "product_id": "prod_001",
  "product_name": "Premium Wireless Headphones",
  "article_themes": {
    "primary_themes": ["ancient Rome", "classical architecture", "historical preservation"],
    "mood": "solemn and majestic",
    "time_period": "1st century CE, Flavian Dynasty",
    "location": "Colosseum, Rome, Italy",
    "visual_elements": ["marble columns", "arched structures", "amphitheater seating", "crowds"],
    "semantic_summary": "An iconic Roman amphitheater showcasing ancient engineering prowess and architectural grandeur."
  },
  "creative_candidates": [
    {
      "id": 1,
      "theme": "Direct Integration",
      "prompt": "Julius Caesar in Roman toga listening to premium wireless headphones, Colosseum in background, marble columns, solemn mood, classical Roman aesthetic",
      "image_url": "/api/website-ads/wad_001/creative/1",
      "created_at": "2026-03-14T10:02:00Z"
    },
    {
      "id": 2,
      "theme": "Ambient Integration",
      "prompt": "Ancient Roman marketplace, merchant showing customer premium wireless headphones, classical architecture, natural lighting, historical accuracy",
      "image_url": "/api/website-ads/wad_001/creative/2",
      "created_at": "2026-03-14T10:02:30Z"
    },
    {
      "id": 3,
      "theme": "Narrative Integration",
      "prompt": "Roman scholar in library with premium wireless headphones, listening to historical audiobook, scrolls and ancient texts, warm lighting",
      "image_url": "/api/website-ads/wad_001/creative/3",
      "created_at": "2026-03-14T10:03:00Z"
    }
  ],
  "status": "generating",
  "selected_creative_id": null,
  "regeneration_count": 0,
  "max_regenerations": 2,
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:03:00Z"
}
```

**Response (200 OK - Status: completed):**

```json
{
  "id": "wad_001",
  "product_id": "prod_001",
  "product_name": "Premium Wireless Headphones",
  "article_themes": { ... },
  "creative_candidates": [ ... ],
  "status": "completed",
  "selected_creative_id": 1,
  "selected_creative_image_url": "/api/website-ads/wad_001/creative/1",
  "variants": {
    "square": {
      "url": "/api/website-ads/wad_001/variant/square",
      "dimensions": "1200x628",
      "created_at": "2026-03-14T10:04:00Z"
    },
    "vertical": {
      "url": "/api/website-ads/wad_001/variant/vertical",
      "dimensions": "300x600",
      "created_at": "2026-03-14T10:04:10Z"
    },
    "icon": {
      "url": "/api/website-ads/wad_001/variant/icon",
      "dimensions": "256x256",
      "created_at": "2026-03-14T10:04:20Z"
    }
  },
  "download_url": "/api/website-ads/wad_001/download",
  "regeneration_count": 0,
  "max_regenerations": 2,
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:04:30Z"
}
```

**Response (200 OK - Status: failed):**

```json
{
  "id": "wad_001",
  "product_id": "prod_001",
  "status": "failed",
  "error": "article_fetch_failed",
  "error_message": "Failed to fetch article from URL: HTTP 403 Forbidden",
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:01:00Z"
}
```

**Error Response (404 Not Found):**

```json
{
  "error": "job_not_found",
  "message": "Website ad job wad_xyz not found",
  "request_id": "req_001",
  "timestamp": "2026-03-14T10:00:00Z"
}
```

---

### POST `/api/website-ads/:job_id/select`

Select one creative candidate and request variant rendering.

**Purpose:** Operator chooses their preferred banner design; system renders variants in multiple formats.

**Content-Type:** `application/json`

**Request Body:**

```json
{
  "creative_id": 1,
  "export_formats": ["square", "vertical", "icon"]
}
```

**Request Fields:**

| Field | Type | Required | Constraints | Example |
|-------|------|----------|-------------|---------|
| `creative_id` | integer | Yes | Must match one of 3 candidates (1-3) | 1 |
| `export_formats` | array | Yes | Subset of ["square", "vertical", "icon"] | ["square", "vertical"] |

**Validation:**
- Job status must be "generating"
- `creative_id` must be 1, 2, or 3
- `export_formats` must not be empty
- Each format must be one of: square, vertical, icon

**Response (202 Accepted):**

```json
{
  "id": "wad_001",
  "status": "rendering_variants",
  "message": "Rendering selected creative in requested formats...",
  "selected_creative_id": 1,
  "export_formats": ["square", "vertical", "icon"],
  "updated_at": "2026-03-14T10:04:00Z"
}
```

**Error Responses:**

| Status | Error Code | Message | Cause |
|--------|-----------|---------|-------|
| 400 | `invalid_creative_id` | Creative ID must be 1, 2, or 3 | Invalid creative_id |
| 400 | `invalid_export_formats` | Export formats must be subset of: square, vertical, icon | Invalid format |
| 400 | `empty_export_formats` | At least one export format required | Empty array |
| 400 | `invalid_job_status` | Can only select creative when job status is "generating" | Wrong job state |
| 404 | `job_not_found` | Website ad job wad_xyz not found | Job doesn't exist |

---

### POST `/api/website-ads/:job_id/regenerate`

Reject all candidates and request new creative direction.

**Purpose:** Operator is unsatisfied with all 3 options; system generates new candidates.

**Content-Type:** `application/json`

**Request Body:**

```json
{
  "brand_voice": "playful"
}
```

Or minimal (empty body):

```json
{}
```

**Request Fields:**

| Field | Type | Required | Constraints | Example |
|-------|------|----------|-------------|---------|
| `brand_voice` | string | Optional | One of: "premium", "casual", "professional", "playful" | "playful" |

**Validation:**
- Job status must be "generating"
- `regeneration_count` < `max_regenerations` (2)
- If provided, `brand_voice` must be valid

**Response (202 Accepted):**

```json
{
  "id": "wad_001",
  "status": "analyzing",
  "message": "Generating new creative options with fresh direction...",
  "regeneration_count": 1,
  "max_regenerations": 2,
  "updated_at": "2026-03-14T10:05:00Z"
}
```

**Error Responses:**

| Status | Error Code | Message | Cause |
|--------|-----------|---------|-------|
| 400 | `invalid_job_status` | Can only regenerate when job status is "generating" | Wrong state |
| 400 | `invalid_brand_voice` | brand_voice must be: premium, casual, professional, playful | Bad value |
| 429 | `max_regenerations_exceeded` | Maximum regenerations (2) exceeded. Create a new job. | Limit hit |
| 404 | `job_not_found` | Website ad job wad_xyz not found | Job doesn't exist |

---

### GET `/api/website-ads/:job_id/creative/:creative_id`

Stream candidate banner image for preview.

**Purpose:** Display 1200x628 PNG candidate image in UI.

**Response (200 OK):**
- Content-Type: `image/png`
- Content-Length: (PNG file size)
- Cache-Control: `public, max-age=3600`
- Binary PNG data

**Error Responses:**

| Status | Error Code | Cause |
|--------|-----------|-------|
| 404 | `creative_not_found` | Candidate image doesn't exist |
| 404 | `job_not_found` | Job doesn't exist |
| 410 | `creative_expired` | Temporary image storage cleaned up |

---

### GET `/api/website-ads/:job_id/variant/:format`

Stream a specific format variant (square/vertical/icon) for preview.

**URL Parameters:**
- `format`: one of "square", "vertical", "icon"

**Response (200 OK):**
- Content-Type: `image/png`
- Binary PNG data with transparency

**Error Responses:**

| Status | Error Code | Cause |
|--------|-----------|-------|
| 404 | `variant_not_found` | Format not available or not yet rendered |
| 400 | `invalid_format` | Format must be: square, vertical, icon |
| 404 | `job_not_found` | Job doesn't exist |

---

### GET `/api/website-ads/:job_id/download`

Download all selected variants as a ZIP file.

**Purpose:** User downloads production-ready PNG assets.

**Response (200 OK):**
- Content-Type: `application/zip`
- Content-Disposition: `attachment; filename="wad_001_export.zip"`
- Content-Length: (ZIP file size)
- Binary ZIP data

**ZIP Contents:**

```
wad_001_square.png          (1200x628, transparent background)
wad_001_vertical.png        (300x600, transparent background)
wad_001_icon.png            (256x256, transparent background)
metadata.json               (JSON with job details)
```

**metadata.json format:**

```json
{
  "job_id": "wad_001",
  "product_name": "Premium Wireless Headphones",
  "article_themes": {
    "primary_themes": ["ancient Rome", "classical architecture"],
    "mood": "solemn and majestic"
  },
  "selected_creative_id": 1,
  "export_formats": ["square", "vertical", "icon"],
  "exported_at": "2026-03-14T10:05:00Z",
  "export_note": "Ready for web deployment. Transparent backgrounds included."
}
```

**Error Responses:**

| Status | Error Code | Cause |
|--------|-----------|-------|
| 400 | `variants_not_ready` | Selection not yet completed or still rendering | 
| 404 | `job_not_found` | Job doesn't exist |

---

### GET `/api/website-ads`

List all website ad jobs with pagination and filtering.

**Purpose:** Dashboard listing of all generated ad jobs.

**Query Parameters:**

| Param | Type | Default | Constraints | Example |
|-------|------|---------|-------------|---------|
| `limit` | integer | 20 | 1-100 | 50 |
| `offset` | integer | 0 | ≥ 0 | 0 |
| `status` | string | (none) | One of: requested, analyzing, generating, completed, failed | "completed" |
| `product_id` | string | (none) | Valid product ID | "prod_001" |

**URL Example:**
```
GET /api/website-ads?limit=20&offset=0&status=completed
```

**Response (200 OK):**

```json
{
  "items": [
    {
      "id": "wad_001",
      "product_id": "prod_001",
      "product_name": "Premium Wireless Headphones",
      "status": "completed",
      "selected_creative_id": 1,
      "created_at": "2026-03-14T10:00:00Z",
      "updated_at": "2026-03-14T10:05:00Z"
    },
    {
      "id": "wad_002",
      "product_id": "prod_002",
      "product_name": "Coffee Maker",
      "status": "failed",
      "error": "article_fetch_failed",
      "created_at": "2026-03-14T09:00:00Z",
      "updated_at": "2026-03-14T09:05:00Z"
    }
  ],
  "total": 42,
  "limit": 20,
  "offset": 0
}
```

**Error Response (400 Bad Request):**

```json
{
  "error": "invalid_query_params",
  "message": "limit must be between 1 and 100",
  "request_id": "req_002",
  "timestamp": "2026-03-14T10:00:00Z"
}
```

---

## 5. Status Polling Recommendations

**Polling Intervals:**

During processing:
- `requesting` / `analyzing` / `generating`: poll every **5 seconds**
- Total expected time: < 3 minutes under normal conditions
- Max polling attempts: 36 (180 seconds / 5 seconds)

**Client-side pseudocode:**

```javascript
async function pollJobStatus(jobId) {
  let attempts = 0;
  const maxAttempts = 36;
  const pollInterval = 5000; // 5 seconds

  while (attempts < maxAttempts) {
    const response = await fetch(`/api/website-ads/${jobId}`);
    const job = await response.json();

    if (job.status === 'generating' || job.status === 'completed') {
      return job;
    }

    if (job.status === 'failed') {
      throw new Error(`Job failed: ${job.error}`);
    }

    await new Promise(resolve => setTimeout(resolve, pollInterval));
    attempts++;
  }

  throw new Error('Job polling timeout');
}
```

---

**Next Step:** See [05_Data_Schema_Definitions.md](05_Data_Schema_Definitions.md) for database schema details.
