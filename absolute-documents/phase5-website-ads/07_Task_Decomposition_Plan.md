# Phase 5: Dynamic Website Ads - Task Decomposition Plan

## 1. Rule of Execution

**Complete all Phases 0-4 of CAFAI video ad insertion first.** Phase 5 work starts only after video insertion is production-ready and verified with real users.

Exit criteria for starting Phase 5:
- Phase 4 complete (preview rendering verified)
- Phases 0-4 merged to main branch
- Documentation aligned with shipped system
- Minimum one successful end-to-end demo with real provider configuration

## 2. Locked Phase 5 Decisions

- **Feature name:** Dynamic Website Ads (Phase 5 extension)
- **Input:** Article (URL, text, or headline+body) + product
- **Output:** 3 candidate banners, then 3 format variants (square/vertical/icon) as PNG
- **Operator interaction:** Select design, optionally regenerate (max 2x)
- **Supported formats:** PNG only (no MP4/GIF in Phase 5)
- **Export:** Downloadable ZIP with all variants + metadata
- **No auth:** Matches Phase 0-4 MVP approach
- **No moderation:** Manual operator review only
- **Reuse patterns:** Existing job queue, state machine, provider abstraction from Phases 0-4
- **Database:** New `ad_*` tables, no modifications to Phase 0-4 tables
- **Cloud services:** Azure OpenAI (DALL-E 3) + Text Analytics for Azure; Vultr LLM + image gen for Vultr

## 3. Architectural Constraints

- **Provider abstraction:** Both Azure and Vultr profiles must support Phase 5 from day one
- **Async orchestration:** Reuse existing worker + polling patterns
- **Local control plane:** All orchestration stays local (Go API + SQLite)
- **Temporary cloud storage:** Use Blob/Object Store for intermediate artifacts only
- **Error handling:** Match Phase 0-4 error envelope and retry strategy

## 4. Phase 5 Build Plan

Organized into sequential sprints. Each sprint is a deliverable increment.

---

## Sprint 1: Foundation and Database (Week 1)

### Goals

- Set up Phase 5 database schema
- Create data models
- Build database access layer

### Deliverables

1. **Schema Migration**
   - Create `003_website_ads_schema.sql`
   - Tables: `website_ad_jobs`, `ad_creative_requests`, `ad_variants`, `ad_artifacts`
   - Indexes and foreign keys
   - Verify migration runs without error

2. **Data Models (Go)**
   - `backend/internal/models/website_ad.go` — WebsiteAdJob struct
   - `backend/internal/models/ad_creative.go` — CreativeRequest struct
   - Add models for ArticleTheme, AdVariant, AdArtifact

3. **Database Access Layer**
   - `backend/internal/db/website_ads.go` — CRUD for jobs
   - `backend/internal/db/ad_creative_requests.go` — CRUD for creatives
   - `backend/internal/db/ad_variants.go` — CRUD for variants
   - `backend/internal/db/ad_artifacts.go` — CRUD for artifacts
   - Unit tests for each repository

### Exit Criteria

- `go test ./backend/internal/db` passes
- Schema migration script verified with fresh SQLite database
- Database connections and queries work end-to-end

---

## Sprint 2: Backend Service Layer (Week 1-2)

### Goals

- Implement core orchestration logic
- Define provider interfaces
- Build basic error handling

### Deliverables

1. **Provider Interfaces**
   - `backend/internal/services/interfaces.go` — Add ArticleAnalysisClient, ImageGenerationClient
   - Define contract methods and return types
   - Create mock implementations for testing

2. **WebsiteAdService**
   - `backend/internal/services/website_ad_service.go` — Main orchestrator
   - Methods: CreateWebsiteAdJob, GetWebsiteAdJob, SelectCreative, RegenerateCreatives, ExportVariants
   - Job state transitions: requested → analyzing → generating → completed|failed
   - Preserve artifacts on error for retry

3. **Article Analysis Client (Azure)**
   - `backend/internal/services/azure_article_analysis_client.go`
   - Fetch article from URL (or use provided text)
   - Call Azure Text Analytics API
   - Call Azure OpenAI to generate semantic theme summary
   - Return ArticleTheme struct

4. **Article Analysis Client (Vultr)**
   - `backend/internal/services/vultr_article_analysis_client.go`
   - Vultr LLM API call with article content
   - Parse response to extract themes
   - Match output format to Azure implementation

5. **Tests**
   - `backend/tests/website_ad_service_test.go` — Service logic tests (mocked providers)
   - `backend/tests/article_analysis_client_test.go` — Provider client tests

### Exit Criteria

- `go test ./backend/internal/services` passes (with mocked providers)
- Service state machine behaves correctly
- Error cases handled and logged appropriately

---

## Sprint 3: Image Generation Integration (Week 2)

### Goals

- Integrate image generation providers
- Build creative prompt generation logic
- Test full analysis → generation pipeline

### Deliverables

1. **Image Generation Client (Azure)**
   - `backend/internal/services/azure_image_generation_client.go`
   - Call Azure OpenAI DALL-E 3 API
   - Generate 1200x628 PNG for each prompt
   - Handle timeouts and retries

2. **Image Generation Client (Vultr)**
   - `backend/internal/services/vultr_image_generation_client.go`
   - Call Vultr image generation service (or Replicate/Stability integration)
   - Same output format as Azure

3. **Creative Prompt Generator**
   - `backend/internal/services/creative_prompt_generator.go` (or in website_ad_service.go)
   - Input: ArticleTheme + ProductName + ProductDescription
   - Output: 3 distinct creative prompts (direct, ambient, narrative integration)
   - Ensure variety while maintaining product-article connection

4. **Image Storage**
   - Local storage of candidate images in `tmp/website_ads/`
   - Cloud storage of intermediate artifacts (optional cleanup after export)

5. **Tests**
   - `backend/tests/image_generation_client_test.go` — Provider tests
   - End-to-end job test: submit → analyze → generate → receive 3 images

### Exit Criteria

- 3 candidate banner images generated successfully for test input
- Images verified as valid PNG files
- Image generation timeout handled gracefully
- Error cases tested and logged

---

## Sprint 4: Backend API (Week 3)

### Goals

- Implement REST endpoints
- Wire handlers to service layer
- Build response serialization

### Deliverables

1. **HTTP Handlers**
   - `backend/internal/api/website_ads_handler.go`
   - POST /api/website-ads (create job)
   - GET /api/website-ads/:job_id (get status + candidates)
   - POST /api/website-ads/:job_id/select (select creative + render variants)
   - POST /api/website-ads/:job_id/regenerate (new creative direction)
   - GET /api/website-ads/:job_id/creative/:creative_id (stream candidate image)
   - GET /api/website-ads/:job_id/variant/:format (stream variant image)
   - GET /api/website-ads/:job_id/download (download ZIP)
   - GET /api/website-ads (list all jobs)

2. **Router Integration**
   - `backend/internal/api/router.go` — Add Phase 5 routes
   - Mount WebsiteAdsHandler

3. **Request/Response Types**
   - Input validation (article source, product, formats)
   - Response serialization (jobs, candidates, variants)
   - Error responses (match Phase 0-4 envelope)

4. **Tests**
   - `backend/tests/api_test.go` — Extend with Phase 5 API tests
   - Happy path: create → analyze → generate → select
   - Error cases: bad input, invalid job state, max regenerations exceeded

### Exit Criteria

- All endpoints respond with correct HTTP status codes
- Request validation catches bad input
- Response payloads match API contract
- API tests pass (with mocked providers)

---

## Sprint 5: Worker Integration (Week 3)

### Goals

- Integrate Phase 5 jobs into async worker
- Test job state transitions
- Verify retry logic

### Deliverables

1. **Worker Job Processing**
   - `backend/internal/worker/job_processor.go` — Add Phase 5 case
   - Detect `website_ad_job` type
   - Call WebsiteAdService.ProcessWebsiteAdJob()
   - Handle state transitions and errors

2. **Tests**
   - Worker processes website_ad_job correctly
   - Job transitions through states: requested → analyzing → generating → completed
   - Failure cases retry appropriately
   - Artifacts preserved for failed jobs

### Exit Criteria

- Worker successfully processes website ad jobs end-to-end
- Job state changes visible in database
- Integration tests pass

---

## Sprint 6: Frontend Pages and Components (Week 4)

### Goals

- Build UI for website ad creation and review
- Implement job polling and status display
- Create candidate gallery and selection UI

### Deliverables

1. **Type Definitions**
   - `frontend/src/types/WebsiteAd.ts` — WebsiteAdJob, ArticleTheme, CreativeCandidate, etc.

2. **API Client**
   - `frontend/src/services/websiteAdsApi.ts`
   - Methods: create, getJob, selectCreative, regenerate, download, list

3. **Pages**
   - `frontend/src/pages/WebsiteAdCreatePage.tsx` — Form to submit article + product
   - `frontend/src/pages/WebsiteAdDetailPage.tsx` — Display job status, candidates, selection UI
   - `frontend/src/pages/WebsiteAdsListPage.tsx` — Dashboard of all jobs

4. **Components**
   - `frontend/src/components/WebsiteAdCreateForm.tsx` — Article source selector, product picker
   - `frontend/src/components/CreativeGallery.tsx` — Display 3 candidate banners
   - `frontend/src/components/CreativeOptionCard.tsx` — Individual candidate card
   - `frontend/src/components/VariantSelector.tsx` — Choose formats and download
   - `frontend/src/components/WebsiteAdStatusCard.tsx` — Job status + progress indicator

5. **Custom Hooks**
   - `frontend/src/hooks/useWebsiteAdJob.ts` — Poll job status (5s interval)
   - `frontend/src/hooks/useWebsiteAdCreate.ts` — Create job with input validation

6. **Routing**
   - Update `frontend/src/App.tsx` to add Phase 5 routes

7. **Tests**
   - Component tests: form submission, gallery rendering, selection
   - API client tests: mocked responses
   - Integration test: create job → poll → select → download

### Exit Criteria

- Candidate gallery displays 3 banner designs correctly
- Selection UI allows picking one design
- Download button produces valid ZIP file
- Status polling works (updates every 5s)
- All tests pass

---

## Sprint 7: Variant Rendering (Week 4)

### Goals

- Implement image resizing/variant generation
- Support multiple output formats
- Create ZIP export

### Deliverables

1. **Variant Rendering**
   - `backend/internal/services/image_resize_service.go` (or in website_ad_service.go)
   - Input: base image (1200x628) + target format (square, vertical, icon)
   - Output: PNG with correct dimensions and transparency
   - Supported formats:
     - square: 1200x628
     - vertical: 300x600
     - icon: 256x256
   - Use ffmpeg (already available) or Go image library

2. **ZIP Export**
   - Create ZIP file with selected variants + metadata.json
   - Store in `tmp/website_ads/{job_id}_export.zip`
   - Include metadata: job_id, product_name, themes, export_date

3. **Download Handler**
   - Stream ZIP to client with correct headers
   - Content-Type: application/zip
   - Content-Disposition: attachment

4. **Tests**
   - Variant rendering produces correct dimensions
   - Transparency preserved
   - ZIP file contains expected files
   - Download endpoint serves file correctly

### Exit Criteria

- Variants render at correct dimensions
- PNG transparency works
- ZIP export includes all selected formats
- Download delivery tested end-to-end

---

## Sprint 8: Quality and Hardening (Week 5)

### Goals

- Fix bugs from integration testing
- Harden error handling
- Optimize performance
- Documentation sync

### Deliverables

1. **Integration Testing**
   - Full end-to-end test: create job → analyze → generate → select → export
   - Test with real provider credentials (if available) or excellent mocks
   - Verify output quality (images are visually coherent)

2. **Error Case Testing**
   - Article fetch failure (unreachable URL)
   - Provider timeout (analysis, generation, resize)
   - Network failure recovery
   - Max regeneration limit enforcement
   - Invalid product ID handling

3. **Performance Optimization**
   - Image generation parallel (not sequential)
   - Database query optimization (verify indexes)
   - File I/O efficiency

4. **Documentation Alignment**
   - Verify all spec docs match shipped code
   - Update README with Phase 5 quickstart
   - Add demo runbook for website ad workflow

5. **Tests**
   - Full integration test suite
   - Performance test (SLA verification: < 3 min from submit to 3 candidates)
   - Error recovery tests

### Exit Criteria

- No known bugs
- Error cases handled gracefully
- Performance SLAs met (< 3 min for full workflow)
- All documentation current
- Integration tests pass

---

## Sprint 9: Demo Readiness and Polish (Week 5-6)

### Goals

- Create demo assets and runbook
- Polish UI/UX
- Prepare for handoff

### Deliverables

1. **Demo Assets**
   - Curated article for demo (historic Rome, or similar)
   - Known-good product for demo (headphones, or similar)
   - Baseline generated banners for reference

2. **Demo Runbook**
   - Step-by-step guide to running full website ad demo
   - Environment setup instructions
   - Expected results and timing
   - Troubleshooting guide

3. **UI Polish**
   - Responsive design on mobile/tablet
   - Loading states and spinners
   - Error messages user-friendly
   - Placeholder images while loading

4. **Frontend Styling**
   - Consistent with Phase 0-4 UI
   - Dark mode support (if Phase 0-4 supports it)
   - Accessibility checklist

5. **Tests**
   - All tests passing
   - Test coverage > 80% for new code
   - Demo scenario verified

### Exit Criteria

- One documented, end-to-end demo with real inputs
- UI polish complete
- All team members can run demo successfully
- Ready for user feedback/validation

---

## 5. Estimated Timeline

| Sprint | Focus | Duration | Team Size |
|--------|-------|----------|-----------|
| 1 | Database | 3-4 days | 1-2 engineers |
| 2 | Service layer | 4-5 days | 1-2 engineers |
| 3 | Image generation | 4-5 days | 1-2 engineers |
| 4 | Backend API | 4-5 days | 1-2 engineers |
| 5 | Worker integration | 2-3 days | 1 engineer |
| 6 | Frontend | 5-6 days | 1-2 engineers |
| 7 | Variant rendering | 3-4 days | 1 engineer |
| 8 | Quality & hardening | 5-6 days | 1-2 engineers |
| 9 | Demo readiness | 3-4 days | 1-2 engineers |
| **Total** | **All Phases** | **~6 weeks** | **1-2 engineers** |

---

## 6. Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Image generation quality poor | Mock excellent images during dev; use real DALL-E 3 early |
| Article analysis inaccurate | Test with diverse articles; tune LLM prompts |
| Performance SLA not met | Profile early; parallelize image generation; cache if needed |
| Provider outage | Graceful error handling; dry-run with multiple providers |
| Database migration issues | Test migrations on separate DB; version control SQL scripts |
| Frontend UI delays | Reuse Phase 0-4 component patterns; minimal custom CSS |

---

## 7. Success Criteria for Phase 5 Complete

Phase 5 is ready for release when:

- ✅ All sprints 1-9 deliverables complete
- ✅ > 80% test coverage on new code
- ✅ End-to-end demo runs successfully with real inputs
- ✅ Performance SLA met (< 3 min from submit to candidates)
- ✅ All documentation aligned with shipped code
- ✅ No known critical bugs
- ✅ Both Azure and Vultr providers tested
- ✅ Error handling graceful (no uncaught exceptions)

---

## 8. Post-Phase 5 Roadmap (Future)

Potential enhancements after Phase 5 MVP validation:

- A/B testing multiple designs (not just one selection)
- Animated banners (MP4/GIF export)
- Custom fonts and typography
- Brand safety moderation (automated)
- Programmatic placement to ad networks
- Analytics integration (conversion tracking)
- Batch job processing (multiple jobs at once)
- API key authentication
- User quotas and billing

These are explicitly deferred to allow rapid Phase 5 validation and market feedback.

---

**Next Step:** After Phase 4 completion, use this plan to guide Phase 5 implementation.
