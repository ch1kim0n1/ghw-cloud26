# Phase 5: Dynamic Website Ads - System Architecture Document

## 1. Purpose

Define the high-level architecture for Phase 5 (Dynamic Website Ads) in relation to the existing CAFAI Phases 0-4 infrastructure. Explain component responsibilities, data flow, and how Phase 5 reuses existing patterns.

## 2. Architectural Principles

- **Reuse:** Phase 5 leverages existing job orchestration, provider abstraction, and state machine patterns from Phases 0-4
- **Isolation:** Website ad generation does not impact video ad insertion; separate job types, database tables, but shared queue
- **Provider Parity:** Azure and Vultr both support Phase 5 services (image generation, text analysis)
- **Local-First Control Plane:** Keep orchestration local (Go API + SQLite); offload heavy lifting to cloud

## 3. High-Level Control Plane Flow

### Current State (Phases 0-4)

```
┌─────────────────────────────────────────────────────────────┐
│                   React Dashboard                           │
│  (Products, Campaigns, Video Analysis, Job Management)     │
└─────────────────────────────────────────────────────────────┘
                           ↕ HTTP API
┌─────────────────────────────────────────────────────────────┐
│              Go Local Control Plane                          │
│  (HTTP handlers, job orchestration, provider abstraction)   │
└─────────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────────┐
│           SQLite Local Database                             │
│  (campaigns, jobs, products, slots, previews, artifacts)   │
└─────────────────────────────────────────────────────────────┘
                           ↕
┌──────────────────┬──────────────────┬──────────────────────┐
│  Local Storage   │  Temp Cloud      │  Cloud Services      │
│  (uploads,       │  Storage (Blob/  │  (Azure/Vultr        │
│   artifacts,     │   Object Store)  │   analysis, gen,     │
│   previews)      │                  │   rendering)         │
└──────────────────┴──────────────────┴──────────────────────┘
```

### Phase 5 Extension

```
┌─────────────────────────────────────────────────────────────┐
│                   React Dashboard                           │
│  (Products, Campaigns, VIDEO JOBS, WEBSITE AD JOBS)        │
└─────────────────────────────────────────────────────────────┘
                           ↕ HTTP API
┌─────────────────────────────────────────────────────────────┐
│              Go Local Control Plane                          │
│  (HTTP handlers, job orchestration, provider abstraction)   │
│  + NEW: Website ad handlers (POST /api/website-ads/...)    │
└─────────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────────┐
│           SQLite Local Database                             │
│  (campaigns, jobs, products, slots, previews,               │
│   + NEW: website_ad_jobs, ad_creative_requests,            │
│           ad_variants, ad_artifacts)                        │
└─────────────────────────────────────────────────────────────┘
                           ↕
┌──────────────────┬──────────────────┬──────────────────────┐
│  Local Storage   │  Temp Cloud      │  Cloud Services      │
│  (uploads,       │  Storage (Blob/  │  (Azure/Vultr        │
│   artifacts,     │   Object Store)  │   + NEW: Text        │
│   previews,      │                  │   analysis,          │
│   website ads)   │                  │   image gen)         │
└──────────────────┴──────────────────┴──────────────────────┘
```

## 4. Phase 5 Component Architecture

### 4.1 Frontend: New Routes and Pages

#### New Route: `/website-ads`

```
/website-ads
├── /create        → WebsiteAdCreatePage (article input + product selection)
├── /:job_id       → WebsiteAdDetailPage (view creative options, select)
└── /:job_id/download  → Trigger ZIP download of selected variants
```

#### New Components

- **WebsiteAdForm.tsx** — Article URL/text input, product selection, preview
- **CreativeOptionCard.tsx** — Display candidate banner designs
- **AdVariantSelector.tsx** — Select which formats to export (square/vertical/icon)
- **WebsiteAdStatusCard.tsx** — Job progress and status

### 4.2 Backend: New Service Layer

#### New Job Type: `website_ad_job`

**Service: `website_ad_service.go`**

Responsibilities:

- Coordinate website ad generation workflow
- Manage job state transitions (requested → analyzing → generating → completed|failed)
- Handle creative regeneration (user rejects and tries again)
- Orchestrate provider calls for analysis and image generation

**Core Methods:**

```go
func (s *WebsiteAdService) CreateWebsiteAdJob(ctx context.Context, req CreateWebsiteAdJobRequest) (*models.WebsiteAdJob, error)
func (s *WebsiteAdService) SubmitForAnalysis(ctx context.Context, jobID string) error
func (s *WebsiteAdService) SubmitForGeneration(ctx context.Context, jobID string, selectedCreativeIndex int) error
func (s *WebsiteAdService) RegenerateCreatives(ctx context.Context, jobID string) error
func (s *WebsiteAdService) ExportVariants(ctx context.Context, jobID string, formats []string) ([]byte, error)
```

#### New Provider Clients

**`article_analysis_client.go`**

Extracts article themes and semantic context.

Provider implementations:

- `AzureArticleAnalysisClient` (uses Azure OpenAI + Text Analytics)
- `VultrArticleAnalysisClient` (Vultr LLM service)

Responsibilities:

- Fetch article from URL or ingest provided text
- Extract themes, mood, visual elements, time period, location
- Return structured theme summary

**`image_generation_client.go`**

Generates banner designs from creative prompts.

Provider implementations:

- `AzureImageGenerationClient` (Azure OpenAI DALL-E 3)
- `VultrImageGenerationClient` (Vultr/external image gen service)

Responsibilities:

- Accept creative prompt
- Generate image in specified dimensions
- Return image bytes or blob reference
- Support multiple format generation

### 4.3 Backend: Database Layer

**New Tables:**

- `website_ad_jobs` (parent job record)
- `ad_creative_requests` (track analysis + generation attempts)
- `ad_variants` (final exported variants: square/vertical/icon)
- `ad_artifacts` (references to cloud/local storage)

(See [05_Data_Schema_Definitions.md](05_Data_Schema_Definitions.md) for full schema)

### 4.4 Backend: Worker Integration

Existing polling worker (`worker/job_processor.go`) extends to handle:

```go
case "website_ad_job":
    return s.websiteAdService.ProcessWebsiteAdJob(ctx, job)
```

Worker flow:

```
1. Job state = "requested"
   → Submit article for analysis
   → State = "analyzing"

2. Job state = "analyzing"
   → Poll for analysis completion
   → When ready: generate 3 creative prompts
   → Submit for image generation
   → State = "generating"

3. Job state = "generating"
   → Poll for image generation completion
   → When ready: render all format variants
   → State = "completed"

4. Job fails at any stage
   → State = "failed", preserve artifacts for manual retry
```

## 5. Data Flow: End-to-End Website Ad Creation

### Flow Diagram

```
User Input (Article + Product)
         ↓
[React] Create Website Ad Job
         ↓ POST /api/website-ads
[Go API] Create Job Record
         ↓ INSERT website_ad_jobs table
[SQLite] Store job_id, article_source, product_id, status="requested"
         ↓
[Worker] Detect "requested" job
         ↓
[Service] Fetch article (URL or provided text)
         ↓
[Provider] Analyze article → Extract themes
         ↓
[Service] Generate 3 creative prompts (theme + product blend)
         ↓
[Provider] Image generation × 3 (one per creative prompt)
         ↓
[Service] Download images, store locally
         ↓ UPDATE website_ad_jobs
[SQLite] Mark status="completed", store 3 variant references
         ↓
[React] GET /api/website-ads/:job_id
         ↓
[UI] Display 3 candidate banners for operator review
         ↓
User selects one design
         ↓
[React] POST /api/website-ads/:job_id/select
         ↓
[Go API] Render full format variants (square, vertical, icon)
         ↓
[Provider] Image rendering/resizing service OR local ffmpeg
         ↓
[Service] Create PNG variants with transparency
         ↓ UPDATE ad_variants table
[SQLite] Store variant references
         ↓
User clicks "Download"
         ↓
[React] GET /api/website-ads/:job_id/download
         ↓
[Go API] Stream ZIP containing all variants + metadata
         ↓
User receives ready-for-web assets
```

## 6. Provider Service Integration

### Azure Provider Profile

**Article Analysis:**
- Azure Text Analytics: Extract key phrases, entities, sentiment
- Azure OpenAI: Generate semantic theme summary

**Image Generation:**
- Azure OpenAI DALL-E 3: Generate banner images

**Infrastructure:**
- Azure Blob Storage: Temporary artifact storage
- Managed Identity: Authentication

### Vultr Provider Profile

**Article Analysis:**
- Vultr LLM Service: Extract themes and semantic context

**Image Generation:**
- Vultr Image Generation Service (or Replicate/Stability integration)

**Infrastructure:**
- Vultr Object Storage (S3-compatible): Temporary storage
- API keys: Authentication

### Provider Abstraction

Existing `provider_profile.go` extends with new interfaces:

```go
type ArticleAnalysisClient interface {
    AnalyzeArticle(ctx context.Context, articleText string) (*ArticleTheme, error)
}

type ImageGenerationClient interface {
    GenerateImage(ctx context.Context, prompt string, width int, height int) (io.Reader, error)
}
```

Both `azure` and `vultr` profiles implement both interfaces.

## 7. Async Job Orchestration

Phase 5 uses the **same job queue pattern** as Phases 0-4:

1. User creates website ad job → job record inserted with status="requested"
2. Worker polls for "requested" jobs
3. Worker transitions job through states: analyzing → generating → completed (or failed)
4. Frontend polls job status endpoint
5. When completed, UI shows results

**Key Difference from Video:** Website ad generation is faster (< 5 minutes vs. 10-20 minutes for video), so polling interval can be more frequent (5 seconds vs. 10 seconds).

## 8. Storage Model

### Local Storage Paths

```
tmp/
├── website_ads/
│   ├── {job_id}_creative_1.png
│   ├── {job_id}_creative_2.png
│   ├── {job_id}_creative_3.png
│   ├── {job_id}_final_square.png
│   ├── {job_id}_final_vertical.png
│   ├── {job_id}_final_icon.png
│   └── {job_id}_export.zip
```

### Cloud Storage (Temporary)

During processing:
- Upload article text + creative prompts to Blob/Object Storage
- Store intermediate image outputs
- Clean up after export (optional, cost optimization)

After completion:
- Final PNG variants stored locally
- Cloud artifacts eligible for cleanup

## 9. Error Handling and Retry

**Design Principle:** Match video CAFAI (Phase 0-4) error handling.

| Failure Scenario | Behavior |
|------------------|----------|
| Article fetch fails | Job fails, user can retry by resubmitting |
| Analysis times out | Job fails, user can retry |
| Image generation produces low-quality output | Display to user, let them reject and regenerate |
| Image generation provider unavailable | Job fails with error message |
| Variant rendering fails | Job partially succeeds; user can re-export |
| Download fails | HTTP 500, logs preserved for manual recovery |

**Regeneration Limit:** User can regenerate up to 2 times per session (cost control).

## 10. Relationship to Existing Architecture

### Reused Components

| Component | Reused In Phase 5 | Changes |
|-----------|------------------|---------|
| Job orchestration | ✅ Yes | New job type `website_ad_job` |
| State machine | ✅ Yes | Similar states (requested → analyzing → generating → completed) |
| Provider abstraction | ✅ Yes | New interfaces for article analysis + image generation |
| Worker polling | ✅ Yes | Same polling pattern, faster intervals |
| SQLite database | ✅ Yes | New tables, but same patterns |
| Error handling | ✅ Yes | Same error envelope + logging |
| Environment config | ✅ Yes | Reuse `CAFAI_PROVIDER_PROFILE` |

### New Components (Phase 5 Only)

| Component | Purpose |
|-----------|---------|
| WebsiteAdService | Orchestrate website ad workflow |
| ArticleAnalysisClient | Extract article themes |
| ImageGenerationClient | Generate banner images |
| WebsiteAdJob model | Database record for ad job |
| `/website-ads/*` API routes | Frontend integration endpoints |
| WebsiteAdForm, CreativeOptionCard components | New React UI pieces |

## 11. Deployment Model

Phase 5 deployment requires:

- **Existing:** Go backend binary, React frontend, SQLite
- **New cloud services needed:**
  - Image generation (DALL-E 3 or equivalent)
  - Text analysis (Azure Text Analytics or LLM)
- **Environment variables:** Extend `.env` with image generation credentials

No new infrastructure beyond what's already in use for Phases 0-4.

---

**Next Step:** See [03_Technical_Specifications.md](03_Technical_Specifications.md) for implementation details.
