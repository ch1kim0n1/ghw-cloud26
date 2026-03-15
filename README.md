# ghw-cloud26

Context-Aware Fused Ad Insertion (CAFAI) for MLH Global Hack Week.

This project analyzes a source video, finds a plausible insertion moment, generates a short branded bridge clip, and stitches that clip back into the original footage as a downloadable preview.

## What It Does

CAFAI is an operator-driven ad insertion workflow:

1. upload a product and source video
2. analyze the video for candidate insertion windows
3. select an automatic slot or manually override it
4. generate a short bridge clip around the chosen moment
5. stitch the generated clip into the source video
6. export a preview MP4

The current implementation uses:
- Azure Video Indexer for scene analysis
- Azure OpenAI for slot ranking, product line generation, and generation briefs
- Higgsfield Kling as the primary Phase 3 generator
- Azure ML as fallback generation wiring
- Azure Blob Storage for temporary render artifacts
- a local Go control plane, SQLite metadata store, and React dashboard

## Screenshots

The frontend is a React dashboard with a voxel-inspired, pink-accented UI.

**Home** — hero with featured demo video and proof wall CTA

![Home](readme-assets/home.png)

**Upload** — campaign name, brand, and MP4 dropzone

![Upload](readme-assets/upload.png)

**Gallery** — all three demo examples with video playback and proof strip

![Gallery](readme-assets/gallery.png)

**About** — team profiles (Vlad & Monika)

![About](readme-assets/about.png)

**Proof room** — four receipts (original frame, insert window, generated bridge, final cut) for each demo

![Proof room](readme-assets/proof-room.png)

## Demo Results & Videos

The app showcases three polished examples. Each has a final stitched video, generated bridge preview, and anchor frames.

### Featured: Pixel Pop Energy

Streamer close-up with an early energy-drink insert. The branded moment lands right after the opening beat.

| Final stitched preview | Generated bridge |
|------------------------|------------------|
| [example3-final.mp4](frontend/public/demo/example3-final.mp4) | ![example3-generated](frontend/public/demo/example3-generated.gif) |

- **Source duration:** 82.2s → **Preview:** 88.5s
- **Insert window:** 7.9s → 8.6s
- **Anchor frames:** 237 → 259

### Bike Bloom Reveal

Outdoor bicycle sequence with a late-scene handoff.

| Final stitched preview | Generated bridge |
|------------------------|------------------|
| [example1-final.mp4](frontend/public/demo/example1-final.mp4) | ![example1-generated](frontend/public/demo/example1-generated.gif) |

- **Source duration:** 59.5s → **Preview:** 64.5s
- **Insert window:** 41.7s → 43.4s
- **Anchor frames:** 1250 → 1300

### Desk Darling Bridge

Desk-side talking head with a seamless branded bridge.

| Final stitched preview | Generated bridge |
|------------------------|------------------|
| [example2-final.mp4](frontend/public/demo/example2-final.mp4) | ![example2-generated](frontend/public/demo/example2-generated.gif) |

- **Source duration:** 59.0s → **Preview:** 65.5s
- **Insert window:** 20.5s → 21.0s
- **Anchor frames:** 615 → 630

## Hackathon Demo Flow

The strongest current demo path is:

1. run analysis on the short validation clip
2. select or manually define one insertion window
3. generate the bridge clip
4. if provider generation is blocked, import a manually generated MP4 through the new manual-import path
5. stitch the final preview

This repo now supports that manual recovery path explicitly, so a generated clip can be treated as if it came back from the API and continue through the normal preview flow.

## Validation Assets

The repo includes a concrete validation package under [phase4-validation](phase4-validation).

### Structure

Each example lives under `phase4-validation/input/ExampleN/` and `phase4-validation/output/ExampleN/`:

| Example | Product | Source video | Generated bridge | Final preview |
|---------|---------|---------------|------------------|---------------|
| Example 1 | [product.jpg](phase4-validation/input/Example1/product/product.jpg) | [phase4_test_59s.mp4](phase4-validation/input/Example1/video/phase4_test_59s.mp4) | [hf_...mp4](phase4-validation/output/Example1/video/hf_20260314_191119_ba726ac9-6ed2-4ac1-b9e1-696055d5e81f.mp4) | [manual_import_preview_api.mp4](phase4-validation/output/Example1/manual_import_preview_api.mp4) |
| Example 2 | [product.jpg](phase4-validation/input/Example2/product/product.jpg) | [example2_59s.mp4](phase4-validation/input/Example2/video/example2_59s.mp4) | [hf_...mp4](phase4-validation/output/Example2/video/hf_20260314_200445_73beb1a5-504f-4ac7-b174-ca69a5bc66a2.mp4) | [manual_import_preview_api.mp4](phase4-validation/output/Example2/manual_import_preview_api.mp4) |
| Example 3 | [product.jpg](phase4-validation/input/Example3/product/product.jpg) | [videoplayback (1).mp4](phase4-validation/input/Example3/video/videoplayback%20(1).mp4) | [hf_...mp4](phase4-validation/output/Example3/video/hf_20260315_053749_592fecd7-ff36-4beb-acba-170ce0f16107.mp4) | — |

### Anchor frames

- Example 1: [start-frame.png](phase4-validation/output/Example1/start-stop-frames/start-frame.png), [stop-frame.png](phase4-validation/output/Example1/start-stop-frames/stop-frame.png)
- Example 2: [start-frame.png](phase4-validation/output/Example2/start-stop-frames/start-frame.png), [stop-frame.png](phase4-validation/output/Example2/start-stop-frames/stop-frame.png)
- Example 3: [1.png](phase4-validation/output/Example3/start-stop-frames/1.png), [2.png](phase4-validation/output/Example3/start-stop-frames/2.png)

### Product metadata

All examples use [metadata.json](phase4-validation/input/Example1/product/metadata.json) (Pepsi Cola).

## Current Status

Implemented now:
- Phase 0: foundation and runtime bootstrap
- Phase 1: product and campaign ingest
- Phase 2: explicit analysis start, worker polling, scene persistence, slot review, reject, and re-pick
- Phase 3: slot selection, manual slot override, Russian-aware generation inputs, provider-aware caching, Higgsfield-primary generation, Azure ML fallback wiring, and manual generated-clip import
- Phase 4: preview render start, render polling, preview persistence, streaming, and download

Practical current state:
- the local control plane works
- automated backend and frontend tests pass
- the short validation flow works through analysis and slot selection
- manual generated MP4 import works and is now part of the backend API
- final preview stitching can be completed locally
- live cloud render still needs provider-side stabilization

Intentionally deferred:
- Phase 5: demo hardening, broader validation, and production-grade reliability work

## Architecture At A Glance

- React dashboard for the operator workflow
- Go REST API as the local control plane
- SQLite as the metadata store
- local filesystem for uploads, debug artifacts, cache, and preview outputs
- polling worker as the MVP async mechanism
- Azure Video Indexer + Azure OpenAI for analysis and ranking
- Higgsfield Kling as primary media generation
- Azure ML retained as fallback media generation path
- Azure Blob Storage + render service for cloud stitching

Canonical job flow:

1. create or select a product
2. create a campaign and upload the source video
3. explicitly start analysis
4. review proposed insertion slots
5. select a proposed slot or manually enter a slot window
6. review or edit the product line
7. generate a CAFAI bridge clip
8. render one preview MP4
9. download the final preview from local storage

Manual recovery flow now supported:

1. generate the bridge clip outside the app
2. call `POST /api/jobs/{job_id}/slots/manual-import`
3. mark that slot as `generated`
4. continue through the normal preview render flow

## API Surface

Base path: `/api`

Live route groups:
- health: `GET /api/health`
- products: `POST /api/products`, `GET /api/products`
- campaigns: `POST /api/campaigns`, `GET /api/campaigns/{campaign_id}`
- jobs: `GET /api/jobs/{job_id}`, `GET /api/jobs/{job_id}/logs`
- analysis: `POST /api/jobs/{job_id}/start-analysis`
- slots: list, detail, select, manual-select, manual-import, reject, re-pick, and generate under `/api/jobs/{job_id}/slots`
- preview: render, status, stream, and download under `/api/jobs/{job_id}/preview`

Standard API errors use one envelope:
- `error`
- `error_code`
- `http_status`
- `details`
- `timestamp`

## Repository Layout

Important paths:
- `backend/cmd/server`: Go server entrypoint
- `backend/internal/api`: HTTP handlers and router
- `backend/internal/db`: SQLite access and migration bootstrap
- `backend/internal/models`: domain structs mirroring API and schema concepts
- `backend/internal/services`: provider clients, artifact helpers, generation, render, and manual import logic
- `backend/internal/worker`: polling worker
- `backend/scripts/migrations`: executable SQL migrations
- `frontend/src/pages`: dashboard pages
- `frontend/src/content/demoContent.ts`: demo examples metadata (synced with frontend)
- `frontend/public/demo`: demo videos, GIFs, posters, frames
- `phase4-validation/`: demo assets, provider outputs, and stitched results
- `tmp/`: local runtime databases, artifacts, cache, previews, and debug output

## Development

Prerequisites:
- Go `1.25+`
- `ffmpeg` and `ffprobe` on `PATH`
- Node.js + npm

macOS install:
```bash
brew install ffmpeg
```

Backend:
```bash
go run ./backend/cmd/server
```

Backend tests:
```bash
cd backend
go test ./...
```

Frontend:
```bash
cd frontend
npm install
npm run dev
```

Frontend tests:
```bash
cd frontend
npm run test
```

Frontend production build:
```bash
cd frontend
npm run build
```

## Runtime Requirements

- Phase 2, 3, and 4 are part of the shipped app and require provider configuration before backend startup
- local `ffmpeg` is required for anchor extraction and local stitch fallback
- when `HIGGSFIELD_API_KEY` and `HIGGSFIELD_API_SECRET` are set, Phase 3 generation tries Higgsfield first
- Azure ML remains the fallback generation provider
- preview rendering still expects blob/render provider configuration
- local SQLite, local uploads, and local output folders remain part of the MVP control plane

## Documentation

Core engineering docs live in [absolute-documents/01_Product_Design_Document.md](absolute-documents/01_Product_Design_Document.md) through [absolute-documents/10_BEFORE_PHASE5_TODO.md](absolute-documents/10_BEFORE_PHASE5_TODO.md).

Recommended reading order:
1. [01_Product_Design_Document.md](absolute-documents/01_Product_Design_Document.md)
2. [02_System_Architecture_Document.md](absolute-documents/02_System_Architecture_Document.md)
3. [03_Technical_Specifications.md](absolute-documents/03_Technical_Specifications.md)
4. [06_API_Contracts.md](absolute-documents/06_API_Contracts.md)
5. [07_Data_Schema_Definitions.md](absolute-documents/07_Data_Schema_Definitions.md)
6. [08_Task_Decomposition_Plan.md](absolute-documents/08_Task_Decomposition_Plan.md)
7. [10_BEFORE_PHASE5_TODO.md](absolute-documents/10_BEFORE_PHASE5_TODO.md)

Document purposes:
- [01_Product_Design_Document.md](absolute-documents/01_Product_Design_Document.md): canonical MVP behavior and success criteria
- [02_System_Architecture_Document.md](absolute-documents/02_System_Architecture_Document.md): local control plane, cloud provider split, and lifecycle
- [03_Technical_Specifications.md](absolute-documents/03_Technical_Specifications.md): backend, frontend, storage, and media rules
- [06_API_Contracts.md](absolute-documents/06_API_Contracts.md): REST payloads and error shapes
- [07_Data_Schema_Definitions.md](absolute-documents/07_Data_Schema_Definitions.md): SQLite schema and metadata rules
- [08_Task_Decomposition_Plan.md](absolute-documents/08_Task_Decomposition_Plan.md): ordered MVP build phases
- [10_BEFORE_PHASE5_TODO.md](absolute-documents/10_BEFORE_PHASE5_TODO.md): current Phase 4 and pre-Phase 5 gap tracking
