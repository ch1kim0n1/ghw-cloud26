# Phase 5: Dynamic Website Ads - Technical Specifications

## 1. Purpose

Define implementation-accurate MVP behavior for Phase 5 backend services, frontend flows, database layer, and provider integrations. This document is the source of truth for engineering decisions.

## 2. Canonical Phase 5 Contract

The system must:

- accept one article as input (URL, text, or headline + body)
- accept one advertised product with metadata
- analyze the article to extract semantic themes
- generate 3 creative prompts that blend article context + product
- render 3 candidate banner images (1200x628 px each)
- let the operator select one design
- render 3 format variants from selected design: square (1200x628), vertical (300x600), icon (256x256)
- export all variants as a downloadable ZIP file with metadata
- support regeneration up to 2 times if operator rejects all candidates

The system must not:

- auto-generate banners without operator review
- expose provider API request IDs in standard responses
- fail silently if image generation produces low-quality output
- block banner generation if product image is unavailable
- support animated or video banners in Phase 5 (PNG only)

## 3. Provider Service Choices

The system uses `CAFAI_PROVIDER_PROFILE` (existing from Phases 0-4) with `azure` as default and `vultr` as alternative.

### Phase 5 Cloud Services Required

#### Text/Article Analysis

**Azure Provider:**
- Azure Text Analytics API: Extract key phrases, entities, sentiment
- Azure OpenAI API: Generate structured theme summary from article

**Vultr Provider:**
- Vultr LLM Service: Extract themes via LLM prompt

**Contract:**
```
Input: article text (up to 50,000 chars)
Output: ArticleTheme struct with:
  - primary_themes: []string (e.g., ["historic Rome", "ancient architecture"])
  - mood: string (e.g., "solemn", "adventurous")
  - time_period: string (e.g., "Ancient Rome", "Renaissance")
  - location: string (e.g., "Colosseum, Rome")
  - visual_elements: []string (e.g., ["marble columns", "torchlight"])
  - semantic_summary: string (2-3 sentence narrative)
```

#### Image Generation

**Azure Provider:**
- Azure OpenAI DALL-E 3 API: Generate banner images

**Vultr Provider:**
- Vultr Image Generation (or Replicate/Stability AI integration)

**Contract:**
```
Input: creative_prompt (string), width (int), height (int)
Output: image bytes (PNG format)
```

#### Image Processing (Local or Cloud)

For variant resizing (square → vertical → icon):

- **Local Option:** Use ffmpeg (already required for Phase 0-4)
- **Cloud Option:** Image resize service if local ffmpeg unavailable

PNG output must support transparency for web deployment.

## 4. Input and Output Constraints

### Input

**Article Source (one of three forms):**

1. **URL**: `https://example.com/article`
   - System fetches using HTTP GET
   - Extracts text via HTML parsing (heading + body)
   - Maximum content: 50,000 characters
   - Supports common news/blog sites; may fail on paywalled content

2. **Direct Text**: Full article markdown or plain text
   - User pastes into textarea
   - Maximum 50,000 characters
   - System uses directly for analysis

3. **Headline + Body**: Separate inputs for headline and body
   - Headline: max 200 characters
   - Body: max 49,800 characters
   - System concatenates for analysis

**Product:**
- Product ID (if in catalog) OR
- Product name + description + optional image URL

**User Options:**
- Regeneration count: User can reject and regenerate up to 2 times per session
- Export format selection: Choose which variants to include in ZIP (square/vertical/icon)

### Output

**Phase 1: Creative Candidates** (3 banner designs at 1200x628)
- Shown to user for selection
- Stored locally in `tmp/website_ads/{job_id}_creative_1/2/3.png`

**Phase 2: Exported Variants** (after selection)
- Square: 1200x628 PNG
- Vertical: 300x600 PNG
- Icon: 256x256 PNG
- All with transparent backgrounds
- Packaged in ZIP with metadata file (JSON)

**ZIP Contents:**
```
{job_id}_export.zip
├── {job_id}_square.png
├── {job_id}_vertical.png
├── {job_id}_icon.png
└── metadata.json (includes job_id, product_name, themes, export_date)
```

## 5. Job State Machine

**States:** `requested` → `analyzing` → `generating` → `completed` | `failed`

**State Transitions:**

```
requested
  ↓ worker submits for article analysis
analyzing
  ↓ analysis complete, creative prompts generated, image generation submitted
generating
  ↓ all 3 candidate images received, waiting for user selection
  ↑ (operator can request regeneration here, loops back to analyzing)
completed
  ↓ variants rendered after user selects design
failed (from any state)
  ↓ error preserved, user can retry via "Create New Job" (not state recovery)
```

**Deferred States (Future):**
- `selected` (after operator picks one design)
- `rendering` (while variants are being generated)

(In MVP, we combine into single flow: generating → completed)

## 6. API Endpoints

### 6.1 Website Ad Endpoints (New)

#### POST `/api/website-ads`

Create a new website ad job.

**Content-Type:** `application/json`

**Request:**
```json
{
  "article_source": "url" | "text" | "headline_body",
  "article_url": "https://example.com/article",  // if source == "url"
  "article_text": "full article text...",          // if source == "text"
  "article_headline": "Article Headline",          // if source == "headline_body"
  "article_body": "Article body text...",          // if source == "headline_body"
  "product_id": "prod_001",  // if product in catalog
  "product_name": "Wireless Headphones",  // if creating inline
  "product_description": "Premium noise-cancelling headphones",
  "product_image_url": "https://...",  // optional
  "brand_voice": "sophisticated"  // optional: premium|casual|professional|playful
}
```

**Validation:**
- Exactly one of: article_url, article_text, (article_headline + article_body)
- Either product_id OR (product_name + product_description)
- article_text length ≤ 50,000 chars

**Response (201 Created):**
```json
{
  "id": "wad_001",
  "product_id": "prod_001",
  "product_name": "Wireless Headphones",
  "status": "requested",
  "article_preview": "The Colosseum: Engineering Marvel of Ancient Rome...",
  "created_at": "2026-03-14T10:00:00Z",
  "updated_at": "2026-03-14T10:00:00Z"
}
```

#### GET `/api/website-ads/:job_id`

Fetch website ad job status and creative candidates.

**Response (200 OK when generating/completed):**
```json
{
  "id": "wad_001",
  "product_id": "prod_001",
  "product_name": "Wireless Headphones",
  "status": "generating",
  "article_themes": {
    "primary_themes": ["historic Rome", "ancient architecture"],
    "mood": "solemn and majestic",
    "time_period": "Ancient Rome, 72 AD",
    "location": "Colosseum, Rome, Italy",
    "visual_elements": ["marble columns", "crowds", "torchlight"]
  },
  "creative_candidates": [
    {
      "id": 1,
      "prompt": "Julius Caesar in Roman toga listening to premium wireless headphones...",
      "image_url": "/api/website-ads/wad_001/creative/1",
      "created_at": "2026-03-14T10:02:00Z"
    },
    {
      "id": 2,
      "prompt": "Ancient Roman warrior with wireless headphones visible in background...",
      "image_url": "/api/website-ads/wad_001/creative/2",
      "created_at": "2026-03-14T10:02:30Z"
    },
    {
      "id": 3,
      "prompt": "Roman marketplace with merchant using modern wireless headphones...",
      "image_url": "/api/website-ads/wad_001/creative/3",
      "created_at": "2026-03-14T10:03:00Z"
    }
  ],
  "selected_creative_id": null,
  "regeneration_count": 0,
  "max_regenerations": 2,
  "updated_at": "2026-03-14T10:03:00Z"
}
```

#### POST `/api/website-ads/:job_id/select`

Operator selects one creative candidate and requests variant rendering.

**Request:**
```json
{
  "creative_id": 1,
  "export_formats": ["square", "vertical", "icon"]
}
```

**Response (202 Accepted):**
```json
{
  "id": "wad_001",
  "status": "rendering_variants",
  "selected_creative_id": 1,
  "export_formats": ["square", "vertical", "icon"],
  "updated_at": "2026-03-14T10:04:00Z"
}
```

#### POST `/api/website-ads/:job_id/regenerate`

Operator rejects all candidates and requests new creative direction.

**Request:**
```json
{
  "brand_voice": "playful"  // optional override
}
```

**Validation:**
- Can only be called if status == "generating"
- regeneration_count < max_regenerations (2)

**Response (202 Accepted):**
```json
{
  "id": "wad_001",
  "status": "analyzing",
  "regeneration_count": 1,
  "max_regenerations": 2,
  "message": "Generating new creative options...",
  "updated_at": "2026-03-14T10:05:00Z"
}
```

#### GET `/api/website-ads/:job_id/download`

Download final variants as ZIP.

**Response (200 OK with ZIP content):**
- Content-Type: `application/zip`
- Content-Disposition: `attachment; filename="wad_001_export.zip"`

**ZIP contents:**
```
wad_001_square.png
wad_001_vertical.png
wad_001_icon.png
metadata.json
```

#### GET `/api/website-ads/:job_id/creative/:creative_id`

Stream candidate image for preview.

**Response (200 OK with PNG content):**
- Content-Type: `image/png`

### 6.2 List Endpoint

#### GET `/api/website-ads`

List all website ad jobs (with pagination).

**Query params:**
- `limit`: default 20, max 100
- `offset`: default 0
- `status`: optional filter (requested|analyzing|generating|completed|failed)
- `product_id`: optional filter

**Response:**
```json
{
  "items": [
    {
      "id": "wad_001",
      "product_name": "Wireless Headphones",
      "status": "completed",
      "created_at": "2026-03-14T10:00:00Z"
    }
  ],
  "total": 42,
  "limit": 20,
  "offset": 0
}
```

## 7. Database Schema Additions

(Detailed in [05_Data_Schema_Definitions.md](05_Data_Schema_Definitions.md))

**New Tables:**
- `website_ad_jobs` — Parent job record
- `ad_creative_requests` — Track analysis + generation attempts
- `ad_variants` — Final exported variants
- `ad_artifacts` — References to stored image files

## 8. Frontend Flow

### New Routes

- `GET /website-ads` — List all jobs
- `GET /website-ads/create` — Create job form
- `GET /website-ads/:job_id` — View job and select creative
- `GET /website-ads/:job_id/download` — Trigger ZIP download

### New Components

**WebsiteAdCreateForm.tsx**
- Article source selector (URL / text / headline+body)
- Product selector (catalog or inline creation)
- Submit button
- Progress indicator

**CreativeGallery.tsx**
- Display 3 candidate banners in grid
- Show underlying theme summary
- "Select" button per candidate
- "Regenerate" button (with counter)

**VariantSelector.tsx**
- Checkboxes for square/vertical/icon
- Preview each format
- "Download" button
- "Generate More Jobs" button

**WebsiteAdStatusCard.tsx**
- Status badge (requesting|analyzing|generating|completed|failed)
- Progress indicator during processing
- ETA estimate

## 9. Processing Timelines and SLAs

| Phase | Task | Target Time | Notes |
|-------|------|-------------|-------|
| 1 | Article analysis | < 30 sec | Fetch + analyze |
| 2 | Creative prompt generation | < 10 sec | Local processing |
| 3 | Image generation (3 images) | < 90 sec | Cloud parallel |
| 4 | Operator review | N/A | Waiting for user input |
| 5 | Variant rendering | < 30 sec | Resize + optimize |
| **Total** | **Full job (excl. user review)** | **< 3 min** | **Or < 5 min with regeneration** |

## 10. Error Handling

### Error Response Format

**Standard error envelope (matches Phase 0-4):**
```json
{
  "error": "article_fetch_failed",
  "message": "Failed to fetch article from URL: connection timeout",
  "request_id": "req_abc123",
  "timestamp": "2026-03-14T10:00:00Z"
}
```

### Common Error Codes

| Error Code | HTTP Status | Cause | Recovery |
|------------|-------------|-------|----------|
| `article_fetch_failed` | 400 | URL unreachable or malformed | User provides direct text |
| `article_parse_failed` | 400 | Could not extract text from page | User provides direct text |
| `analysis_timeout` | 504 | Provider analysis took too long | User can retry |
| `generation_timeout` | 504 | Image generation took too long | User can retry |
| `generation_quality_issue` | 400 | Generated images too low quality | User can regenerate |
| `variant_rendering_failed` | 500 | PNG resize/export failed | User can retry |
| `storage_unavailable` | 503 | Local/cloud storage issue | System will retry |
| `max_regenerations_exceeded` | 429 | User hit regeneration limit | Start new job |

## 11. Local and Cloud Storage

### Local Storage

```
tmp/website_ads/
├── wad_001_creative_1.png        (1200x628)
├── wad_001_creative_2.png        (1200x628)
├── wad_001_creative_3.png        (1200x628)
├── wad_001_selected_square.png   (1200x628, selected variant)
├── wad_001_selected_vertical.png (300x600, selected variant)
├── wad_001_selected_icon.png     (256x256, selected variant)
└── wad_001_export.zip            (all variants + metadata)
```

### Cloud Storage (Temporary)

During generation:
- Upload article text to Blob/Object Storage for processing
- Store intermediate images while waiting for user selection
- Clean up after job completion (optional cost optimization)

## 12. Configuration

### New Environment Variables (Extend `.env.example`)

```bash
# Image Generation (Phase 5)
CAFAI_IMAGE_GEN_MODEL=dalle-3                    # Model choice
CAFAI_IMAGE_GEN_QUALITY=hd                       # Quality level
CAFAI_IMAGE_GEN_TIMEOUT=120                      # Timeout in seconds

# Article Analysis (Phase 5)
CAFAI_ARTICLE_ANALYSIS_TIMEOUT=30                # Timeout in seconds
CAFAI_ARTICLE_MAX_LENGTH=50000                   # Character limit

# Website Ads Feature
CAFAI_ENABLE_WEBSITE_ADS=true                    # Feature flag
CAFAI_WEBSITE_ADS_MAX_REGENERATIONS=2            # User regeneration limit
```

## 13. Testing Requirements

### Backend Tests (Go)

- `website_ads_api_test.go` — REST endpoint integration tests
- `website_ad_service_test.go` — Business logic tests (mocked providers)
- `article_analysis_client_test.go` — Provider client tests
- `image_generation_client_test.go` — Image generation tests

### Frontend Tests (React/Vitest)

- `WebsiteAdCreateForm.test.tsx` — Form submission
- `CreativeGallery.test.tsx` — Gallery rendering and selection
- `website-ads-api.test.ts` — API client integration

### Integration Tests

- End-to-end job creation → analysis → generation → selection → export

## 14. Out of Scope (Phase 5)

- A/B testing variants (single selection in MVP)
- Animated banners (MP4/GIF)
- Custom fonts
- Brand safety moderation
- Programmatic ad network placement
- Analytics/conversion tracking
- Batch job processing
- Video banner export

These are future enhancements post-Phase 5 validation.

---

**Next Step:** See [04_API_Contracts.md](04_API_Contracts.md) for full API reference.
