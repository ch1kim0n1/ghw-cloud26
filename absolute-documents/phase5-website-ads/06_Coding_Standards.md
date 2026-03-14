# Phase 5: Dynamic Website Ads - Coding Standards

## 1. Purpose

Define code organization patterns, naming conventions, testing requirements, and architectural guidelines specific to Phase 5 implementation. Phase 5 should reuse Phase 0-4 patterns where applicable.

## 2. Code Organization (Backend - Go)

### Directory Structure

```
backend/internal/
├── api/
│   ├── website_ads_handler.go          [NEW - HTTP routes for Phase 5]
│   └── router.go                       [MODIFY - add Phase 5 routes]
│
├── services/
│   ├── website_ad_service.go           [NEW - orchestrator for Phase 5 workflow]
│   ├── article_analysis_client.go      [NEW - article → themes]
│   ├── image_generation_client.go      [NEW - prompt → image]
│   └── interfaces.go                   [MODIFY - add Phase 5 interfaces]
│
├── db/
│   ├── website_ads.go                  [NEW - website_ad_jobs CRUD]
│   ├── ad_creative_requests.go         [NEW - creative request CRUD]
│   ├── ad_variants.go                  [NEW - variant CRUD]
│   ├── ad_artifacts.go                 [NEW - artifact CRUD]
│   └── migrations.go                   [MODIFY - register Phase 5 migration]
│
├── models/
│   ├── website_ad.go                   [NEW - WebsiteAdJob struct]
│   ├── ad_creative.go                  [NEW - creative request struct]
│   └── common.go                        [MODIFY - shared types if needed]
│
├── worker/
│   └── job_processor.go                [MODIFY - add website_ad_job case]
│
└── constants/
    └── workflow.go                      [MODIFY - add Phase 5 states]
```

### Handler Organization

**File:** `backend/internal/api/website_ads_handler.go`

```go
package api

import (
  "net/http"
  "github.com/yourrepo/backend/internal/services"
)

// WebsiteAdsHandler handles all Phase 5 routes
type WebsiteAdsHandler struct {
  websiteAdService *services.WebsiteAdService
}

// POST /api/website-ads
func (h *WebsiteAdsHandler) CreateWebsiteAdJob(w http.ResponseWriter, r *http.Request) { ... }

// GET /api/website-ads/:job_id
func (h *WebsiteAdsHandler) GetWebsiteAdJob(w http.ResponseWriter, r *http.Request) { ... }

// POST /api/website-ads/:job_id/select
func (h *WebsiteAdsHandler) SelectCreative(w http.ResponseWriter, r *http.Request) { ... }

// POST /api/website-ads/:job_id/regenerate
func (h *WebsiteAdsHandler) RegenerateCreatives(w http.ResponseWriter, r *http.Request) { ... }

// GET /api/website-ads/:job_id/creative/:creative_id
func (h *WebsiteAdsHandler) GetCreativeImage(w http.ResponseWriter, r *http.Request) { ... }

// GET /api/website-ads/:job_id/variant/:format
func (h *WebsiteAdsHandler) GetVariantImage(w http.ResponseWriter, r *http.Request) { ... }

// GET /api/website-ads/:job_id/download
func (h *WebsiteAdsHandler) DownloadVariants(w http.ResponseWriter, r *http.Request) { ... }

// GET /api/website-ads
func (h *WebsiteAdsHandler) ListWebsiteAdJobs(w http.ResponseWriter, r *http.Request) { ... }
```

### Service Layer Organization

**File:** `backend/internal/services/website_ad_service.go`

Follow the same pattern as `job_service.go` (Phase 0-4):

```go
package services

type WebsiteAdService struct {
  db                    *sql.DB
  articleAnalysisClient ArticleAnalysisClient
  imageGenClient        ImageGenerationClient
  blobClient            BlobStorageClient
}

// Public methods (match API contract)
func (s *WebsiteAdService) CreateWebsiteAdJob(ctx context.Context, req CreateWebsiteAdJobRequest) (*models.WebsiteAdJob, error)
func (s *WebsiteAdService) GetWebsiteAdJob(ctx context.Context, jobID string) (*WebsiteAdJobResponse, error)
func (s *WebsiteAdService) SelectCreative(ctx context.Context, jobID string, creativeID int) error
func (s *WebsiteAdService) RegenerateCreatives(ctx context.Context, jobID string, brandVoice string) error
func (s *WebsiteAdService) ExportVariants(ctx context.Context, jobID string, formats []string) ([]byte, error)

// Private methods (internal orchestration)
func (s *WebsiteAdService) ProcessWebsiteAdJob(ctx context.Context, job *models.Job) error
func (s *WebsiteAdService) submitForAnalysis(ctx context.Context, jobID string) error
func (s *WebsiteAdService) submitForGeneration(ctx context.Context, jobID string) error
func (s *WebsiteAdService) pollGenerationStatus(ctx context.Context, jobID string) error
func (s *WebsiteAdService) renderVariants(ctx context.Context, jobID string, formats []string) error
```

### Provider Client Interfaces

**File:** `backend/internal/services/interfaces.go`

Add to existing interfaces:

```go
// ArticleAnalysisClient extracts themes from article text
type ArticleAnalysisClient interface {
  AnalyzeArticle(ctx context.Context, articleText string) (*ArticleTheme, error)
}

// ImageGenerationClient generates banner images from prompts
type ImageGenerationClient interface {
  GenerateImage(ctx context.Context, prompt string, width int, height int) (io.Reader, error)
}

// ArticleTheme contains extracted semantic information
type ArticleTheme struct {
  PrimaryThemes  []string
  Mood           string
  TimePeriod     string
  Location       string
  VisualElements []string
  SemanticSummary string
}
```

**Provider Implementations:**

```
backend/internal/services/
├── azure_article_analysis_client.go   [NEW]
├── azure_image_generation_client.go   [NEW]
├── vultr_article_analysis_client.go   [NEW]
└── vultr_image_generation_client.go   [NEW]
```

### Database Layer

**File:** `backend/internal/db/website_ads.go`

```go
package db

// Repository methods for website_ad_jobs table
func (conn *SqliteConnection) CreateWebsiteAdJob(ctx context.Context, job *models.WebsiteAdJob) error
func (conn *SqliteConnection) GetWebsiteAdJob(ctx context.Context, jobID string) (*models.WebsiteAdJob, error)
func (conn *SqliteConnection) UpdateWebsiteAdJob(ctx context.Context, job *models.WebsiteAdJob) error
func (conn *SqliteConnection) ListWebsiteAdJobs(ctx context.Context, filter WebsiteAdJobFilter) ([]*models.WebsiteAdJob, error)
```

**Similar files for:** `ad_creative_requests.go`, `ad_variants.go`, `ad_artifacts.go`

### Models

**File:** `backend/internal/models/website_ad.go`

```go
package models

type WebsiteAdJob struct {
  ID                   string     `json:"id"`
  ProductID            string     `json:"product_id"`
  ArticleSource        string     `json:"article_source"`       // 'url', 'text', 'headline_body'
  ArticleURL           *string    `json:"article_url,omitempty"`
  ArticleTitle         string     `json:"article_title"`
  ArticleTextHash      string     `json:"-"` // internal
  ArticleTextPreview   string     `json:"article_preview"`
  BrandVoice           *string    `json:"brand_voice,omitempty"`
  Status               string     `json:"status"`
  ErrorCode            *string    `json:"error_code,omitempty"`
  ErrorMessage         *string    `json:"error_message,omitempty"`
  SelectedCreativeID   *int       `json:"selected_creative_id,omitempty"`
  RegenerationCount    int        `json:"regeneration_count"`
  MaxRegenerations     int        `json:"max_regenerations"`
  CreatedAt            time.Time  `json:"created_at"`
  UpdatedAt            time.Time  `json:"updated_at"`
  CompletedAt          *time.Time `json:"completed_at,omitempty"`
}

type ArticleTheme struct {
  PrimaryThemes   []string `json:"primary_themes"`
  Mood            string   `json:"mood"`
  TimePeriod      string   `json:"time_period"`
  Location        string   `json:"location"`
  VisualElements  []string `json:"visual_elements"`
  SemanticSummary string   `json:"semantic_summary"`
}
```

### Constants and Enums

**File:** `backend/internal/constants/workflow.go`

Add Phase 5 job types and states:

```go
package constants

// Job type for Phase 5
const WebsiteAdJobType = "website_ad_job"

// Phase 5 job states
const (
  WebsiteAdStateRequested  = "requested"   // Awaiting processing
  WebsiteAdStateAnalyzing  = "analyzing"   // Article analysis + prompt generation
  WebsiteAdStateGenerating = "generating"  // Image generation, awaiting selection
  WebsiteAdStateCompleted  = "completed"   // Variants rendered, ready for download
  WebsiteAdStateFailed     = "failed"      // Error state
)
```

## 3. Code Organization (Frontend - React/TypeScript)

### Directory Structure

```
frontend/src/
├── pages/
│   ├── WebsiteAdCreatePage.tsx         [NEW]
│   ├── WebsiteAdDetailPage.tsx         [NEW]
│   └── WebsiteAdsListPage.tsx          [NEW]
│
├── components/
│   ├── WebsiteAdCreateForm.tsx         [NEW]
│   ├── CreativeGallery.tsx             [NEW]
│   ├── CreativeOptionCard.tsx          [NEW]
│   ├── VariantSelector.tsx             [NEW]
│   ├── WebsiteAdStatusCard.tsx         [NEW]
│   └── ProductSelector.tsx             [MODIFY - reuse from Phase 0-1]
│
├── services/
│   └── websiteAdsApi.ts                [NEW - API client]
│
├── types/
│   ├── WebsiteAd.ts                    [NEW - type definitions]
│   └── Api.ts                          [MODIFY - add Phase 5 types]
│
└── hooks/
    ├── useWebsiteAdJob.ts              [NEW - job polling]
    └── useWebsiteAdCreate.ts           [NEW - job creation]
```

### Types

**File:** `frontend/src/types/WebsiteAd.ts`

```typescript
export interface WebsiteAdJob {
  id: string;
  product_id: string;
  product_name: string;
  article_source: 'url' | 'text' | 'headline_body';
  article_url?: string;
  article_preview: string;
  status: WebsiteAdJobStatus;
  article_themes?: ArticleTheme;
  creative_candidates?: CreativeCandidate[];
  selected_creative_id?: number;
  regeneration_count: number;
  max_regenerations: number;
  variants?: AdVariants;
  created_at: string;
  updated_at: string;
  error?: string;
  error_message?: string;
}

export type WebsiteAdJobStatus = 'requesting' | 'analyzing' | 'generating' | 'completed' | 'failed';

export interface ArticleTheme {
  primary_themes: string[];
  mood: string;
  time_period: string;
  location: string;
  visual_elements: string[];
  semantic_summary: string;
}

export interface CreativeCandidate {
  id: number;
  theme: string;
  prompt: string;
  image_url: string;
  created_at: string;
}

export interface AdVariants {
  square: VariantFormat;
  vertical: VariantFormat;
  icon: VariantFormat;
}

export interface VariantFormat {
  url: string;
  dimensions: string;
  created_at: string;
}
```

### API Client

**File:** `frontend/src/services/websiteAdsApi.ts`

```typescript
import { apiClient } from './apiClient';

export const websiteAdsApi = {
  create: (req: CreateWebsiteAdRequest) =>
    apiClient.post('/website-ads', req),

  getJob: (jobId: string) =>
    apiClient.get<WebsiteAdJob>(`/website-ads/${jobId}`),

  selectCreative: (jobId: string, creativeId: number, formats: string[]) =>
    apiClient.post(`/website-ads/${jobId}/select`, {
      creative_id: creativeId,
      export_formats: formats,
    }),

  regenerate: (jobId: string, brandVoice?: string) =>
    apiClient.post(`/website-ads/${jobId}/regenerate`, {
      brand_voice: brandVoice,
    }),

  download: (jobId: string) =>
    apiClient.downloadFile(`/website-ads/${jobId}/download`),

  list: (limit = 20, offset = 0, status?: string) =>
    apiClient.get(`/website-ads?limit=${limit}&offset=${offset}${status ? `&status=${status}` : ''}`),
};
```

### Custom Hooks

**File:** `frontend/src/hooks/useWebsiteAdJob.ts`

```typescript
import { useEffect, useState } from 'react';
import { websiteAdsApi } from '../services/websiteAdsApi';
import { WebsiteAdJob } from '../types/WebsiteAd';

export const useWebsiteAdJob = (jobId: string) => {
  const [job, setJob] = useState<WebsiteAdJob | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const pollInterval = setInterval(async () => {
      try {
        const response = await websiteAdsApi.getJob(jobId);
        setJob(response);

        if (response.status === 'completed' || response.status === 'failed') {
          clearInterval(pollInterval);
        }
      } catch (err) {
        setError(err.message);
        clearInterval(pollInterval);
      } finally {
        setLoading(false);
      }
    }, 5000); // 5 second poll

    return () => clearInterval(pollInterval);
  }, [jobId]);

  return { job, loading, error };
};
```

## 4. Naming Conventions

### Go

**Identifiers:**
- Variables: `camelCase` (existing Phase 0-4 convention)
- Functions/Methods: `PascalCase` for exported, `camelCase` for private (existing convention)
- Constants: `SCREAMING_SNAKE_CASE` (existing convention)
- Package names: lowercase, no underscores

**IDs and Prefixes:**
- Website ad job IDs: `wad_` prefix (e.g., `wad_001`, `wad_abc123`)
- Creative request IDs: `acr_` prefix
- Variant IDs: `adv_` prefix
- Artifact IDs: `ada_` prefix

### React/TypeScript

**File names:**
- Components: `PascalCase.tsx` (existing convention)
- Services: `camelCase.ts`
- Types: `PascalCase.ts`
- Hooks: `useXxx.ts` (existing convention)

**React conventions:**
- Props interfaces: `ComponentNameProps`
- State variables: `camelCase`
- Handlers: `handleXxx`

### Database

- Table names: `snake_case` with Phase 5 prefix (`ad_*` or `website_ad_*`)
- Column names: `snake_case`
- Indexes: `idx_tablename_column`

## 5. Error Handling

### Backend Error Pattern

Follow Phase 0-4 error handling:

```go
// errors.go additions
func NewWebsiteAdError(code, message string) error {
  return &APIError{
    Code:      code,
    Message:   message,
    RequestID: generateRequestID(),
    Timestamp: time.Now().UTC(),
  }
}

// Usage in handlers
if err != nil {
  return NewWebsiteAdError("article_fetch_failed", "Failed to fetch article URL: " + err.Error())
}
```

### Frontend Error Handling

```typescript
try {
  await websiteAdsApi.selectCreative(jobId, creativeId, formats);
} catch (error) {
  if (error.response?.status === 429) {
    setError('Maximum regenerations exceeded');
  } else if (error.response?.status === 400) {
    setError(error.response.data.message);
  } else {
    setError('Unexpected error');
  }
}
```

## 6. Testing Requirements

### Backend Tests

**Location:** `backend/tests/`

**Files:**
- `website_ads_api_test.go` — HTTP endpoint tests
- `website_ad_service_test.go` — Business logic tests
- `article_analysis_client_test.go` — Provider client tests

**Pattern:**
```go
func TestCreateWebsiteAdJob_Success(t *testing.T) { ... }
func TestCreateWebsiteAdJob_InvalidArticleSource(t *testing.T) { ... }
func TestSelectCreative_UpdatesJobState(t *testing.T) { ... }
func TestRegenerateCreatives_EnforcesLimit(t *testing.T) { ... }
```

**Test Coverage Target:** ≥ 80% for Phase 5 services

### Frontend Tests

**Location:** `frontend/src/`

**Files:**
- `pages/WebsiteAdCreatePage.test.tsx`
- `components/CreativeGallery.test.tsx`
- `components/VariantSelector.test.tsx`
- `services/websiteAdsApi.test.ts`

**Pattern:**
```typescript
describe('WebsiteAdCreateForm', () => {
  test('submits form with valid article URL', () => { ... });
  test('validates article text length', () => { ... });
  test('displays creative candidates after job completes', () => { ... });
});
```

**Test Tools:**
- Vitest 2.1.8 (matching Phase 0-4)
- React Testing Library
- MSW for API mocking

## 7. State Management

### Backend Job Processing

Follow Phase 0-4 job state machine pattern:

```go
// Worker polls for "requested" website ad jobs
// Transition: requested → analyzing
// Perform: article analysis + creative prompt generation
// Transition: analyzing → generating
// Perform: image generation, store artifacts
// Transition: generating → completed (when user selects)
// or: any state → failed (on error)
```

### Frontend State

Use React hooks for local state:

```typescript
const [job, setJob] = useState<WebsiteAdJob | null>(null);
const [selectedCreative, setSelectedCreative] = useState<number | null>(null);
const [selectedFormats, setSelectedFormats] = useState<string[]>(['square']);
const [isDownloading, setIsDownloading] = useState(false);
```

No Redux/context needed for MVP (simple polling + linear flow).

## 8. Security Considerations

### Input Validation

- **URLs:** Validate scheme (http/https only), prevent SSRF
- **Article text:** Max 50,000 chars, sanitize for LLM injection
- **Product IDs:** Verify existence before use
- **Image URLs:** Same SSRF prevention as articles

### File Handling

- Store images with secure paths (UUID-based)
- Enforce PNG format validation on download
- Set appropriate HTTP headers (Content-Type, Content-Disposition)
- Clean up temporary files after export

### API Security

- No auth in MVP (acceptable for hackathon)
- Rate limit regeneration endpoint (max 2 per job)
- CORS: respect `CAFAI_ALLOWED_ORIGINS` (existing)

## 9. Logging

### Backend Logging

Follow Phase 0-4 pattern (go's `log` or structured logging if available):

```go
log.Printf("website_ad_job_created: %s, product_id=%s, status=%s", jobID, productID, status)
log.Printf("article_analysis_started: %s, article_length=%d", jobID, len(articleText))
log.Printf("image_generation_completed: %s, creative_count=3, duration=%dms", jobID, duration)
```

### Frontend Logging

```typescript
console.log('[WebsiteAdCreateForm] Submitting job:', jobPayload);
console.log('[WebsiteAdDetailPage] Job status updated:', job.status);
console.error('[WebsiteAdsApi] Download failed:', error);
```

## 10. Performance Considerations

- **Article fetching:** Set timeout (30s max)
- **Image generation:** Parallel requests for 3 candidates (not sequential)
- **Variant rendering:** Use ffmpeg with caching if possible
- **Database queries:** Use prepared statements and indexes
- **Frontend polling:** 5s intervals (not aggressive)
- **File storage:** Clean up temporary images after 7 days

---

**Next Step:** See [07_Task_Decomposition_Plan.md](07_Task_Decomposition_Plan.md) for build plan and sprints.
