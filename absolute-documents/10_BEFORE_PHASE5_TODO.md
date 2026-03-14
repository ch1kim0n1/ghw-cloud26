# Phase 4 Gap Assessment

## 1. Purpose

Capture the current real Phase 4 state in the repository, identify what is already implemented, and define what still needs work before Phase 4 can be treated as complete and demo-ready.

Current provider shape:

- analysis: Azure Video Indexer + Azure OpenAI
- generation: Higgsfield Kling primary, Azure ML fallback
- render: Azure Blob Storage + Azure render service

## 2. Current Implemented Phase 4 Scope

The repository already includes the main Phase 4 control flow:

- preview render start endpoint
- preview status endpoint
- preview stream endpoint
- preview download endpoint
- worker-driven render submission and polling
- temporary artifact upload to Blob/Object Storage
- preview artifact download back to local storage
- persisted preview records and retry count handling
- frontend controls for render, open preview, and download preview
- backend tests that exercise a successful preview path

## 3. What Is Working Today

### Backend

- `POST /api/jobs/{job_id}/preview/render` starts a render for the selected generated slot
- `GET /api/jobs/{job_id}/preview` returns persisted preview state
- `GET /api/jobs/{job_id}/preview/stream` serves inline MP4 playback
- `GET /api/jobs/{job_id}/preview/download` serves the preview as a downloadable MP4
- render submission uploads the source video and generated artifacts to temporary cloud storage
- render completion downloads the final preview back to local storage
- render failure marks the job failed while preserving preview artifact metadata for retry

### Frontend

- the job page exposes a Phase 4 render card
- the dashboard shows preview status and retry count
- users can open the preview page
- users can download the completed preview
- the preview page supports inline playback via the stream route

### Tests

- backend API tests cover preview route not-found behavior
- backend API tests cover end-to-end preview render success with fake clients
- frontend tests cover preview route rendering and job-page preview actions

## 4. Remaining Phase 4 Gaps

These are the gaps between "implemented" and "finished enough to call Phase 4 complete without qualification."

### 4.1 Real Provider Validation

The control plane and API flow are implemented, but repository evidence still centers on fake-client and mocked-path validation.

Still needed:

- one verified end-to-end run against the real configured provider profile
- confirmation that uploaded source video, generated clip, and optional generated audio match the real render service contract
- one live confirmation that Higgsfield output downloads cleanly into the local artifact flow
- one live confirmation that Azure ML fallback still works when the primary generation path fails
- confirmation that the preview blob returned by the render provider is playable without manual cleanup

### 4.2 Audio Path Clarity

The current render request always asks for `crossfade`, and optional generated audio is passed through when present.

The current generation path is video-first: Higgsfield may return only a clip, while Azure ML fallback may return clip plus audio.

Still needed:

- explicitly define whether Azure AI Speech is directly part of the current shipped runtime path or only a planned provider-side responsibility
- confirm spoken-line renders for `auto` and `operator` modes
- confirm silent renders for `disabled` mode
- confirm preview quality remains acceptable when generation returns no standalone audio track

### 4.3 Media Quality Validation

The pipeline can mark a preview completed, but there is not yet repository evidence that the output has been reviewed against the product goals.

Still needed:

- visual continuity checks at both anchor boundaries
- audio continuity checks before and after insertion
- verification that output runtime increases by the inserted duration
- verification that the source resumes on the expected finish anchor frame

### 4.4 Retry and Failure Hardening

The code supports retry after render failure, but it still needs stronger behavioral validation.

Still needed:

- tests for retry after a real render failure
- confirmation that preserved artifacts are sufficient for retry without regeneration
- validation of partial-failure cases such as missing downloaded preview files or stale blob references

### 4.5 Operator and Demo Readiness

The UI exposes Phase 4, but the surrounding docs and user messaging were behind the codebase until this update.

Still needed:

- complete doc alignment across all engineering and demo-facing materials
- one documented demo runbook for the full source video to preview workflow
- a known-good demo asset and product pairing validated through the real preview path

Current staged validation assets already exist in the repo:

- `phase4-validation/input/video/phase4_test_60s.mp4`
- `phase4-validation/input/video/phase4_test.mp4`
- `phase4-validation/input/product/product.jpg`
- `phase4-validation/input/product/metadata.json`

## 5. Recommended Completion Plan

### Step 1

Keep the docs aligned with the shipped Phase 4 behavior.

### Step 2

Run the full automated backend and frontend test suites in a working local toolchain.

### Step 3

Execute one baseline end-to-end Phase 4 run with real provider configuration.

Use the staged 60-second baseline clip first:

- `phase4-validation/input/video/phase4_test_60s.mp4`
- `phase4-validation/input/product/product.jpg`
- `phase4-validation/input/product/metadata.json`

### Step 4

Review the produced preview for:

- anchor continuity
- audio continuity
- playable download
- correct duration increase

### Step 5

Exercise the render failure and retry path deliberately.

### Step 6

Write a short demo runbook with the exact environment variables, startup steps, and validation checklist.

## 6. Exit Criteria For Calling Phase 4 Complete

Phase 4 should be treated as complete when all of the following are true:

- preview render works with real provider configuration
- Higgsfield primary generation works or Azure ML fallback is proven live
- preview playback and download work from the dashboard
- output media quality has been checked against anchor and audio continuity expectations
- render failure and retry have both been verified
- docs and operator guidance match the shipped system
