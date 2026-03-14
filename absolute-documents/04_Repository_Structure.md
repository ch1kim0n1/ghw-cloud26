# Repository Structure

## 1. Purpose

Define the monorepo layout for the MVP implementation only.

## 2. Monorepo Decision

Use one repository for:

- Go backend
- React frontend
- SQLite schema and seed scripts
- MVP documentation
- demo assets

## 3. Canonical Layout

```text
ghw-cloud26/
|-- README.md
|-- .gitignore
|-- absolute-documents/
|-- backend/
|   |-- cmd/
|   |   `-- server/
|   |       `-- main.go
|   |-- internal/
|   |   |-- api/
|   |   |   |-- router.go
|   |   |   |-- errors.go
|   |   |   |-- products_handler.go
|   |   |   |-- campaigns_handler.go
|   |   |   |-- jobs_handler.go
|   |   |   |-- analysis_handler.go
|   |   |   |-- slots_handler.go
|   |   |   `-- preview_handler.go
|   |   |-- config/
|   |   |   `-- config.go
|   |   |-- db/
|   |   |   |-- sqlite.go
|   |   |   |-- products.go
|   |   |   |-- campaigns.go
|   |   |   |-- jobs.go
|   |   |   |-- scenes.go
|   |   |   |-- slots.go
|   |   |   |-- previews.go
|   |   |   `-- job_logs.go
|   |   |-- models/
|   |   |   |-- product.go
|   |   |   |-- campaign.go
|   |   |   |-- job.go
|   |   |   |-- scene.go
|   |   |   |-- slot.go
|   |   |   `-- preview.go
|   |   |-- services/
|   |   |   |-- product_service.go
|   |   |   |-- campaign_service.go
|   |   |   |-- analysis_client.go
|   |   |   |-- azure_video_indexer_client.go
|   |   |   |-- azure_openai_client.go
|   |   |   |-- azure_ml_client.go
|   |   |   |-- azure_speech_client.go
|   |   |   |-- blob_storage_client.go
|   |   |   |-- cafai_generator.go
|   |   |   |-- render_client.go
|   |   |   |-- artifact_service.go
|   |   |   `-- job_service.go
|   |   `-- worker/
|   |       `-- job_processor.go
|   |-- scripts/
|   |   |-- init_db.sql
|   |   |-- seed_products.sql
|   |   `-- migrations/
|   `-- tests/
|       |-- api_test.go
|       |-- db_test.go
|       `-- services_test.go
|-- frontend/
|   |-- package.json
|   |-- tsconfig.json
|   |-- public/
|   `-- src/
|       |-- App.tsx
|       |-- main.tsx
|       |-- pages/
|       |   |-- ProductsPage.tsx
|       |   |-- CreateCampaignPage.tsx
|       |   |-- JobPage.tsx
|       |   `-- PreviewPage.tsx
|       |-- components/
|       |   |-- ProductForm.tsx
|       |   |-- CampaignForm.tsx
|       |   |-- JobStatusCard.tsx
|       |   |-- SlotCard.tsx
|       |   |-- ProductLineEditor.tsx
|       |   `-- PreviewPlayer.tsx
|       |-- hooks/
|       |   |-- useJob.ts
|       |   |-- useSlots.ts
|       |   `-- usePreview.ts
|       |-- services/
|       |   |-- apiClient.ts
|       |   |-- productsApi.ts
|       |   |-- campaignsApi.ts
|       |   |-- jobsApi.ts
|       |   |-- analysisApi.ts
|       |   |-- slotsApi.ts
|       |   `-- previewApi.ts
|       |-- types/
|       |   |-- Product.ts
|       |   |-- Campaign.ts
|       |   |-- Job.ts
|       |   |-- Slot.ts
|       |   `-- Preview.ts
|       `-- styles/
|           |-- app.css
|           `-- components.css
|-- demo/
|   `-- mvp_baseline_scene.mp4
|-- infra/
|   `-- azure/
|   |   `-- notes.md
|   `-- vultr/
|       `-- notes.md
`-- tmp/
    |-- uploads/
    |-- artifacts/
    `-- previews/
```

## 4. Placement Rules

- HTTP handlers live in `backend/internal/api/`
- SQL access lives in `backend/internal/db/`
- domain structs live in `backend/internal/models/`
- provider integration code lives in `backend/internal/services/`
- the polling worker lives in `backend/internal/worker/`
- UI pages live in `frontend/src/pages/`
- shared API calls from the frontend live in `frontend/src/services/`
- demo validation assets live in `demo/`

## 5. Directory Intent

### Backend

The backend is a thin control plane:

- save files locally
- write SQLite rows
- call provider services
- expose JSON APIs

Backend local prerequisites:

- `ffprobe` must be available on `PATH` for runtime media inspection
- `ffmpeg` must be available on `PATH` for runtime anchor-frame extraction and the backend media-oriented test suite
- local development is expected to work on both Windows and macOS

### Frontend

The frontend is an operator dashboard for:

- product creation
- campaign creation
- explicit analysis start
- slot review
- product line review
- CAFAI generation status
- preview render start
- preview playback
- preview download

For the currently implemented Phase 0-4 workflow, the primary live operator path is:

- product creation
- campaign creation
- explicit analysis start
- slot review with select, reject, and re-pick
- product line review and CAFAI generation start
- preview render start
- preview playback and download

### tmp/

The temp directory is expected in MVP because local paths are part of debugging and demo visibility.
