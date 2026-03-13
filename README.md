# ghw-cloud26

## Development
- Prerequisites for backend runtime: `ffprobe` and `ffmpeg` on `PATH`
- Prerequisites for backend tests: `ffmpeg` and `ffprobe` on `PATH`
- macOS install: `brew install ffmpeg`
- Backend dev server from the repo root: `go run ./backend/cmd/server`
- Backend tests from `backend/`: `go test ./...`
- Frontend install from `frontend/`: `npm install`
- Frontend dev server from `frontend/`: `npm run dev`
- Frontend tests from `frontend/`: `npm run test`
- Frontend production build from `frontend/`: `npm run build`

## Overview
This repository contains the MVP documentation and implementation plan for a cloud-assisted contextual ad insertion system built during MLH Global Hack Week.

The product strategy is:
- Context-Aware Fused Ad Insertion (CAFAI)

The MVP goal is:
- analyze a provided H.264 MP4.
- propose valid insertion slots automatically
- let the operator choose a slot and optionally edit the generated product line
- generate a short context-aware bridge clip
- stitch that clip into the source video with basic audio continuity
- export one downloadable preview MP4

## MVP Summary
The engineering-facing MVP contract is:
- supported source videos for the main MVP path are 10-20 minute H.264 MP4 files
- the system proposes the top 3 valid anchor-frame insertion slots when possible
- the operator can reject slots and request up to 2 re-picks
- CAFAI generation creates a 5-8 second bridge clip
- final preview download is served from local storage
- Azure Blob Storage is used only for temporary cloud artifacts during generation and rendering
- job states are coarse: `queued -> analyzing -> generating -> stitching -> completed|failed`

## Current State
Implemented now:
- Phase 0: foundation and runtime bootstrap
- Phase 1: product and campaign ingest
- Phase 2: explicit analysis start, worker polling, scene persistence, slot review, reject, and re-pick
- Phase 3: slot selection, product-line review, and CAFAI generation state tracking

Runtime requirements:
- Phases 2 and 3 are part of the shipped app and require Azure Video Indexer, Azure OpenAI, and Azure Machine Learning configuration before the backend will start
- Phase 3 runtime also requires local `ffmpeg` because anchor-frame images are extracted before generation submission
- local SQLite, uploads, and runtime directories remain part of the MVP control plane

Intentionally deferred:
- Phase 4: preview rendering and download

## Documentation
Core engineering docs live in [absolute-documents/01_Product_Design_Document.md](absolute-documents/01_Product_Design_Document.md) through [absolute-documents/08_Task_Decomposition_Plan.md](absolute-documents/08_Task_Decomposition_Plan.md).

Recommended reading order:
1. [01_Product_Design_Document.md](absolute-documents/01_Product_Design_Document.md)
2. [02_System_Architecture_Document.md](absolute-documents/02_System_Architecture_Document.md)
3. [03_Technical_Specifications.md](absolute-documents/03_Technical_Specifications.md)
4. [06_API_Contracts.md](absolute-documents/06_API_Contracts.md)
5. [07_Data_Schema_Definitions.md](absolute-documents/07_Data_Schema_Definitions.md)
6. [08_Task_Decomposition_Plan.md](absolute-documents/08_Task_Decomposition_Plan.md)

Document purposes:
- [01_Product_Design_Document.md](absolute-documents/01_Product_Design_Document.md): canonical MVP product behavior, constraints, and success criteria
- [02_System_Architecture_Document.md](absolute-documents/02_System_Architecture_Document.md): local control plane, Azure-backed compute flow, and job lifecycle
- [03_Technical_Specifications.md](absolute-documents/03_Technical_Specifications.md): implementation-accurate backend, frontend, storage, and media rules
- [04_Repository_Structure.md](absolute-documents/04_Repository_Structure.md): canonical monorepo layout and placement rules
- [05_Coding_Standards.md](absolute-documents/05_Coding_Standards.md): naming, state-model, API, and testing standards
- [06_API_Contracts.md](absolute-documents/06_API_Contracts.md): REST surface, payloads, and error envelope
- [07_Data_Schema_Definitions.md](absolute-documents/07_Data_Schema_Definitions.md): SQLite schema, indexes, and metadata rules
- [08_Task_Decomposition_Plan.md](absolute-documents/08_Task_Decomposition_Plan.md): ordered MVP build phases and demo acceptance checklist

## Architecture At A Glance
The documented MVP architecture is:
- React dashboard for the operator workflow
- Go REST API as the local control plane
- SQLite as the metadata store
- local filesystem for uploads, debug artifacts, and final preview download
- a polling worker as the MVP async mechanism
- Azure Video Indexer + Azure OpenAI for analysis
- Azure Machine Learning + Azure OpenAI for CAFAI generation
- Azure AI Speech + Azure Container Apps + Azure Blob Storage for audio/rendering and temporary artifacts

Canonical job flow:
1. create or select a product
2. create a campaign and upload the source video
3. explicitly start analysis
4. review proposed insertion slots
5. select a slot and review/edit the product line
6. generate a CAFAI bridge clip
7. render one preview MP4
8. download the final preview from local storage

## Repository Layout
The docs define this repo as a single monorepo for backend, frontend, schema, docs, and demo assets.

Important paths:
- `backend/cmd/server`: Go server entrypoint
- `backend/internal/api`: HTTP handlers and router
- `backend/internal/db`: SQLite access and migration bootstrap
- `backend/internal/models`: domain structs mirroring API/schema concepts
- `backend/internal/services`: Azure client interfaces, artifact helpers, and service layer
- `backend/internal/worker`: polling worker
- `backend/scripts/migrations`: executable SQL migrations
- `frontend/src/pages`: routed dashboard pages
- `frontend/src/services`: frontend API client modules
- `frontend/src/types`: frontend contract types
- `tmp/`: runtime uploads, artifacts, previews, and local DB files

## API Surface
The documented MVP API base path is `/api`, with snake_case JSON and no auth in MVP.

Current implementation status:
- `GET /api/health` is live
- products, campaigns, jobs, analysis start, slot list/detail, slot select, slot reject, slot re-pick, and slot generate are implemented
- preview routes remain intentionally unimplemented until Phase 4

Planned MVP route groups from the docs:
- products: `POST /api/products`, `GET /api/products`
- campaigns: `POST /api/campaigns`, `GET /api/campaigns/{campaign_id}`
- jobs: `GET /api/jobs/{job_id}`, `GET /api/jobs/{job_id}/logs`
- analysis: `POST /api/jobs/{job_id}/start-analysis`
- slots: list, detail, select, reject, re-pick, and generate under `/api/jobs/{job_id}/slots`
- preview: render, status, and download under `/api/jobs/{job_id}/preview`

Standard API errors follow one envelope with:
- `error`
- `error_code`
- `http_status`
- `details`
- `timestamp`

## Data Model
The documented MVP database is SQLite with SQL migrations under `backend/scripts/migrations/`.

Core tables:
- `products`: reusable advertised items
- `campaigns`: one source video plus one linked product
- `jobs`: async execution state for analysis, generation, and rendering
- `scenes`: analyzed scene data returned from cloud analysis
- `slots`: ranked candidate anchor-frame insertion pairs
- `job_previews`: the single preview output per job
- `job_logs`: operational audit trail for the dashboard

Important schema rules from the docs:
- IDs are application-generated text IDs
- local file paths are stored intentionally in MVP
- provider request IDs stay in internal metadata only
- slot anchors always use the uploaded source FPS
- render failure preserves generated artifacts for retry

## MVP Build Phases
The task decomposition doc defines the ordered build plan as:
1. Phase 0: Foundation
2. Phase 1: Product and Campaign Ingest
3. Phase 2: Analysis and Slot Proposal
4. Phase 3: Product Line Review and CAFAI Generation
5. Phase 4: Preview Rendering
6. Phase 5: Demo Hardening

The rule of execution in the docs is to finish the MVP end to end before any post-MVP work.

## Azure Service Choices
The current MVP stack assumes:
- analysis: Azure Video Indexer + Azure OpenAI
- CAFAI generation: Azure Machine Learning + Azure OpenAI
- audio generation and alignment: Azure AI Speech
- final render: Azure Container Apps running ffmpeg, with Azure Blob Storage for intermediate artifacts

## Artifact Flow
```text
generation output
      |
      v
Azure Blob Storage (temporary)
      |
      v
render worker pulls artifact
      |
      v
final preview written to Blob
      |
      v
preview copied back to local storage
      |
      v
download and debugging access
```

## Phase Status
Phase 0 foundation is implemented with:
- a runnable Go backend on `net/http`
- executable SQLite migrations and auto-created runtime directories
- Azure-backed Phase 2 and Phase 3 client wiring and a polling worker
- a React + TypeScript dashboard for the live Phase 0-3 workflow
- a live health check at `/api/health`

Phase 1 is implemented with:
- reusable product creation and listing
- campaign creation with source video validation
- inline product creation during campaign intake
- jobs created in `queued` and `ready_for_analysis`

Phase 2 is implemented with:
- explicit analysis start
- worker-driven analysis submission and polling
- scene persistence, slot persistence, and slot ranking
- dashboard slot review, rejection, and re-pick
- provider request IDs recorded internally and hidden from standard API responses

Phase 3 is implemented with:
- slot selection and suggested product-line generation
- anchor-frame image extraction for generation inputs
- dashboard product-line review with `auto`, `operator`, and `disabled` modes
- worker-driven CAFAI generation submission and polling using Azure Machine Learning plus Azure OpenAI
- generated artifact metadata persisted on the selected slot
- generation failures surfaced clearly on both job and slot state

Phase 4 is intentionally deferred:
- preview rendering, preview status, and preview download
