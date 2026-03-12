# Task Decomposition Plan

## 1. Rule of Execution
Finish the MVP first. Post-MVP work starts only after the MVP is running end to end.

## 2. Locked MVP Decisions
- strategy name: CAFAI
- input video for the full MVP path: 10-20 minute H.264 MP4
- output: one downloadable preview MP4
- top 3 slots are proposed automatically when possible
- the operator can select, reject, and re-pick
- slot targets are anchor-frame pairs inside scenes
- product line review is part of MVP
- output runtime increases because the CAFAI clip is inserted
- job states stay coarse
- no auth in MVP
- no fallback generation path in MVP
- local storage and SQLite are used in MVP control flow
- Azure Blob Storage is temporary artifact storage only
- heavy analysis, generation, audio, and rendering use Azure services
- async processing uses a polling worker or goroutine model

## 3. MVP Goal
Deliver one demoable system that:
- analyzes a source clip
- proposes valid insertion slots
- lets the operator choose and optionally edit the product line
- generates one CAFAI clip for the selected slot
- inserts that clip with basic audio continuity
- exports a preview the user can watch and download

## 4. Ordered MVP Build Plan
### Phase 0: Foundation
Deliverables:
- repo skeleton
- executable SQLite schema
- local upload directories
- backend and frontend bootstrap
- environment configuration
- Azure service integration placeholders

Exit criteria:
- app starts locally
- database initializes cleanly
- files can be saved and read back

### Phase 1: Product and Campaign Ingest
Tasks:
- implement `POST /api/products`
- implement `GET /api/products`
- implement `POST /api/campaigns`
- validate video codec and duration
- support inline product creation during campaign setup
- ensure campaign creation leaves the job in `ready_for_analysis`

Exit criteria:
- a product can be created
- a campaign can be created
- a queued job is written to the database without auto-starting analysis

### Phase 2: Analysis and Slot Proposal
Tasks:
- implement `POST /api/jobs/{job_id}/start-analysis`
- add polling worker
- submit video to Azure Video Indexer and Azure OpenAI
- persist scenes
- persist proposed slots
- expose slots in the dashboard
- implement reject and re-pick flow
- enforce exclusion thresholds

Exit criteria:
- one video produces valid proposed slots when they exist
- fewer-than-3-slot cases return available valid slots
- rejected slots are excluded on re-pick

### Phase 3: Product Line Review and CAFAI Generation
Tasks:
- implement slot selection endpoint
- generate suggested product line
- implement product line review UI
- implement generation request with `product_line_mode`
- submit generation request to Azure Machine Learning and Azure OpenAI
- store generated clip metadata
- expose generation progress and failure state

Exit criteria:
- selecting a slot prepares a suggested line
- the operator can accept, edit, or disable the line
- generation success updates slot state
- generation failure fails the job clearly

### Phase 4: Preview Rendering
Tasks:
- submit preview render request
- push generation artifact to Azure Blob Storage
- run render in Azure Container Apps using ffmpeg
- apply Azure AI Speech output or simple crossfade smoothing
- copy final preview back to local storage
- store one preview output
- expose preview status and download route

Exit criteria:
- a completed job has one playable preview MP4
- preview download works from the dashboard
- render-failure path preserves generated artifacts for retry

### Phase 5: Demo Hardening
Tasks:
- improve logs and job visibility
- verify progress reporting weights
- test success and failure paths
- validate the baseline demo profile
- prepare one reliable main demo clip and product

Exit criteria:
- end-to-end demo is repeatable
- the seamlessness story and Azure compute story are both understandable

## 5. Baseline Validation Profile
### Name
`MVP_BASELINE_TEST_PROFILE`

### Purpose
Use this smaller test case first to validate the implementation before moving to the full 10-20 minute MVP target.

### Baseline Video
- duration: 40-60 seconds
- scene count: 3
- camera: static
- motion: low
- dialogue: light conversation
- cuts: 4 or fewer

### Example Scene
Two people sitting at a kitchen table talking.

### Example Product
- product_name: sparkling water
- category: beverage

### Expected Insertion
- duration: 6 seconds
- interaction: pick up bottle -> sip -> set down
- dialogue: optional short line

### Expected Pipeline Result
- analysis returns 2-3 valid slots
- one slot is chosen
- generation creates a simple interaction
- render produces a preview clip

### Demo Asset Recommendation
`/demo/mvp_baseline_scene.mp4`

## 6. MVP Acceptance Checklist
- product creation works
- campaign creation works
- explicit analysis start works
- analysis returns valid proposed slots
- reject and re-pick works
- product line review works
- one selected slot generates a CAFAI clip
- preview render completes
- final preview is downloadable
- the result demonstrates seamless insertion better than a hard cut

## 7. Explicitly Deferred Until After MVP
- fallback composition path
- production auth and access control
- multi-tenant support
- permanent cloud object storage migration
- PostgreSQL migration
- queue and event bus architecture
- advanced photorealism work
- streaming platform integrations
- billing, quotas, and analytics

## 8. Post-MVP Backlog
Only after MVP completion:
- add fallback generation or composition
- improve narrative-fit scoring
- improve audio realism
- improve character-product interaction quality
- migrate storage and database to production services
- add auth, RBAC, and operational controls

## 9. Demo Success Definition
The hackathon demo is successful if viewers can clearly understand:
- the system found the insertion moment automatically
- the inserted clip feels like it belongs inside the scene
- the Azure-backed cloud pipeline is doing the heavy lifting that makes the result possible
