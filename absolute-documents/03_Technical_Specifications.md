# Technical Specifications

## 1. Purpose

This document defines the implementation-accurate MVP behavior for the Go backend, React dashboard, SQLite data layer, local filesystem storage, and provider-backed cloud processing used by CAFAI.

## 2. Canonical MVP Contract

The system must:

- accept one 10-20 minute H.264 MP4 as the main supported MVP input
- accept one advertised product with image or source URL plus metadata
- create a campaign without auto-starting analysis
- explicitly start analysis through a dedicated API call
- propose up to the top 3 valid anchor-frame insertion slots automatically
- let the operator select, reject, or re-pick slots
- generate a suggested product line after slot selection
- let the operator accept, edit, or disable that line
- generate one 5-8 second CAFAI clip for the selected slot
- insert that clip between the chosen source anchor frames
- export one downloadable preview MP4

The system must not:

- replace source runtime instead of inserting new runtime
- hide generation failure behind a fallback path
- expose provider request IDs in standard API responses

## 3. Provider Service Choices

The MVP uses `CAFAI_PROVIDER_PROFILE` with `azure` as the default and `vultr` as a supported alternative for Phases 2-4.

The MVP names the following cloud services:

- analysis: Azure Video Indexer + Azure OpenAI, or a Vultr-hosted analysis service
- CAFAI clip generation: Azure Machine Learning + Azure OpenAI, or a Vultr-hosted generation service
- audio generation and alignment: Azure AI Speech
- final render: Azure Container Apps running ffmpeg, or a Vultr-hosted render service
- temporary artifact storage: Azure Blob Storage or Vultr Object Storage

## 4. Storage Model

### 4.1 Canonical Rule

Azure Blob Storage or Vultr Object Storage is used as temporary cloud artifact storage during generation and rendering. After rendering completes, the final preview file is copied back to local storage for MVP download, inspection, and debugging.

### 4.2 Storage Roles

- Azure Blob Storage or Vultr Object Storage: temporary generation and render artifacts
- local storage: uploads, debug artifacts, and final preview download

## 5. Input and Output Constraints

### Input

- source video: H.264 MP4 only
- source duration: 10-20 minutes for the full MVP path
- product image: PNG or JPG when provided
- product metadata: name, description, optional category, optional source URL, optional context keywords

### Output

- one preview MP4 per job
- local filesystem path intentionally exposed for MVP debugging

## 6. Local Control Plane

### 6.0 Local Runtime Dependencies

The local Phase 0-4 control plane assumes these command-line tools are available on `PATH`:

- `ffprobe` for runtime media inspection during campaign intake
- `ffmpeg` for runtime anchor-frame extraction during Phase 3 generation plus backend test fixture generation and later-phase media work

Cross-platform expectation:

- Windows and macOS are both valid local development environments
- macOS setup should use Homebrew, for example `brew install ffmpeg`
- the backend should fail fast at startup when `ffprobe` is missing instead of deferring that failure to upload time

### 6.1 Go API

Responsibilities:

- multipart upload handling
- request validation
- SQLite CRUD
- explicit analysis start endpoint
- slot selection and product line review endpoints
- generation start and progress endpoints
- job creation and status endpoints
- health endpoint including the active provider profile
- preview download endpoint

### 6.2 Local Filesystem

Stores:

- uploaded source videos
- uploaded product images
- extracted anchor frames
- downloaded cloud artifacts
- final preview output

### 6.3 SQLite

Stores:

- products
- campaigns
- jobs
- scenes
- slots
- job_previews
- job_logs

### 6.4 Polling Worker

Loop behavior:

- poll for queued jobs awaiting analysis
- poll for selected slots awaiting line review completion
- poll for selected slots ready for generation
- poll for generated slots ready for preview rendering
- update coarse job states and granular current_stage fields

There is no event bus in MVP.

## 7. Lightweight Local Processing

Local code may perform only lightweight pre-processing before cloud submission:

- validate file type and codec with ffprobe
- read actual source FPS
- read source duration
- extract anchor thumbnails for the UI
- persist debug artifacts

Frame math must always use the source video FPS. Never hardcode 24 FPS.

Operational rule:

- if `ffprobe` or `ffmpeg` is not installed locally, the backend startup path should surface a clear dependency error with platform-appropriate install guidance
- if Azure Video Indexer, Azure OpenAI, or Azure Machine Learning configuration is incomplete, the backend startup path should fail fast instead of silently enabling placeholder runtime behavior

## 8. Cloud Analysis Stage

### 8.1 Purpose

Analyze the full video and return enough data to:

- propose valid insertion slots
- let the operator inspect and choose one
- feed the selected slot into CAFAI generation and stitching

### 8.2 Required Outputs

For each detected scene:

- scene_number
- start_frame
- end_frame
- start_seconds
- end_seconds
- motion_score
- stability_score
- dialogue_activity_score
- longest_quiet_window_seconds
- narrative_summary
- context_keywords
- action_intensity_score
- abrupt_cut_risk

For each candidate slot:

- scene_id
- anchor_start_frame
- anchor_end_frame
- quiet_window_seconds
- score
- reasoning
- context_relevance_score
- narrative_fit_score
- anchor_continuity_score

### 8.3 Hard Exclusion Rules

Reject a candidate slot when any of the following is true:

- normalized motion score across the candidate window is greater than `0.65`
- either anchor boundary sub-window exceeds motion score `0.75`
- action intensity score across the candidate window is greater than `0.70`
- any 1-second sub-window exceeds action intensity `0.80`
- a shot boundary is detected within `0.5` seconds of either anchor
- cut-confidence at either anchor exceeds `0.70`
- scene duration is less than `10` seconds
- no quiet window of at least `3` seconds exists

### 8.4 Candidate Definition

A slot is not just a scene. A slot is a proposed anchor-frame pair inside a scene:

- `anchor_start_frame` is the source frame immediately before insertion
- `anchor_end_frame` is the source frame immediately after insertion
- the generated CAFAI clip is inserted between those two frames

### 8.5 Ranking Rules

The ranking engine returns up to the top 3 valid slots automatically.

Reference scoring formula:

```text
slot_score =
  stability_score * 0.30 +
  quiet_window_score * 0.25 +
  context_relevance_score * 0.20 +
  narrative_fit_score * 0.15 +
  anchor_continuity_score * 0.10
```

Sub-score definitions:

- `quiet_window_score`: normalized score capped once quiet window duration reaches at least 3 seconds
- `context_relevance_score`: weighted keyword overlap between scene context tokens and product descriptor tokens, normalized by total product descriptor weight
- `narrative_fit_score`: weighted heuristic score based on dialogue activity, tone consistency, pacing, and contextual relevance
- `anchor_continuity_score`: weighted combination of color histogram similarity, structural similarity, and motion difference between the start and finish anchors

### 8.6 Fewer-Than-Three Rule

If analysis produces fewer than 3 valid slots:

- return the available 1-2 valid slots
- fail the job only if zero valid slots are found

## 9. Slot Review and Re-Pick

### 9.1 Slot Review

The dashboard must expose returned slots with:

- rank
- anchor frames
- score
- reasoning
- current slot status

### 9.2 Re-Pick Rules

The operator may request a re-pick only after all current proposed slots are rejected.

Re-pick behavior:

- exclude all previously rejected slot IDs
- keep the same ranking criteria and thresholds
- allow up to 2 re-pick attempts
- fail the job with `error_code=NO_SUITABLE_SLOT_FOUND` after the second re-pick if no acceptable slot remains

## 10. Product Line Review

### 10.1 Suggested Line

After slot selection, the system generates a suggested product line from product metadata and scene context.

### 10.2 Product Line Modes

- `auto`: use the system-generated line
- `operator`: use the operator-provided edited line
- `disabled`: generate a silent product interaction instead of spoken dialogue

### 10.3 UI Behavior

The operator may:

- accept the generated line
- edit the line
- disable dialogue entirely

## 11. CAFAI Generation Stage

### 11.1 Purpose

Generate the in-between ad moment for one selected slot.

### 11.2 Canonical Generation Method

CAFAI generates the inserted scene by conditioning a short video-generation or compositing model on the selected start and end anchor frames, the surrounding scene context, and the target product, then synthesizing a new 5-8 second bridge clip in which the character naturally interacts with the product and the clip resolves back into the original finish anchor frame.

### 11.3 Required Inputs

- source video path
- anchor_start_frame image
- anchor_end_frame image
- source FPS
- target duration, default `6` seconds and maximum `8` seconds
- product image or source reference
- product name
- product description
- product context keywords
- scene narrative summary
- `product_line_mode`
- `custom_product_line` when provided

### 11.4 Required CAFAI Behavior

The inserted scene is produced by:

- taking the chosen start anchor frame as the visual starting image
- generating a short sequence of intermediate frames that introduce and animate the product within that scene
- constraining the sequence to land near the chosen finish anchor frame
- stitching that generated bridge clip back into the original video with matched audio

The generated clip should:

- begin in visual continuity with the start anchor frame
- end in visual continuity with the finish anchor frame
- show the product directly in the scene
- depict the on-screen character naturally interacting with the product
- include either a short spoken mention or silent interaction depending on `product_line_mode`
- remain narratively plausible enough to not feel like a hard ad break

### 11.5 Output

- generated clip path or downloaded local copy
- generated audio track or muxed clip with audio
- generation duration
- quality metadata if available

### 11.6 Failure Policy

There is no fallback generation path in MVP.

If CAFAI generation fails:

- slot status becomes `failed`
- job status becomes `failed`
- `current_stage` remains in generation-related state
- the dashboard exposes the failure reason

### 11.7 RIFE and Smoothing

RIFE is not part of the canonical MVP generation path.

It may be used optionally after generation for visual interpolation or minor temporal smoothing if boundary motion looks rough, but the MVP must not depend on it for successful output.

## 12. Audio Requirements

The MVP guarantees basic audio continuity by:

- preserving the original soundtrack where possible
- optionally inserting a short synthesized product mention
- applying simple crossfade-based smoothing at clip boundaries

The generated CAFAI clip may contain either:

- a short spoken product mention
- a silent product interaction

The output does not need studio-grade audio mixing, but it must not feel obviously broken.

## 13. Preview Rendering Stage

### 13.1 Purpose

Insert the CAFAI clip into the source video and produce one preview MP4.

### 13.2 Required Rules

- insert new duration between `anchor_start_frame` and `anchor_end_frame`
- do not replace the surrounding source footage
- output runtime equals source runtime plus inserted clip duration
- preserve the finish anchor frame as the point where the source movie resumes
- use Azure Blob Storage or Vultr Object Storage for temporary render artifacts
- copy the finished preview back to local storage

### 13.3 Render Failure Behavior

If render fails after generation succeeds:

- preserve the generated clip artifact
- set `status=failed`
- set `current_stage=render`
- set `error_code=PREVIEW_RENDER_FAILED`
- allow render retry without re-running generation

### 13.4 Output

- one preview MP4
- one local output path
- render metrics such as duration and insertion frame range

## 14. API and State Requirements

### 14.1 Job Status Values

Allowed job status values:

- `queued`
- `analyzing`
- `generating`
- `stitching`
- `completed`
- `failed`

### 14.2 Current Stage Values

Canonical current_stage values:

- `ready_for_analysis`
- `analysis_submission`
- `analysis_poll`
- `slot_selection`
- `line_review`
- `generation_submission`
- `generation_poll`
- `render`
- `render_submission`
- `render_poll`

### 14.3 Slot Status Values

Allowed slot status values:

- `proposed`
- `selected`
- `rejected`
- `generating`
- `generated`
- `failed`

## 15. Progress Reporting

The MVP progress indicator is stage-based:

- analysis: `40%`
- clip generation: `40%`
- stitching and rendering: `20%`

## 16. Provider Request IDs

Provider request IDs must be recorded in internal logs and job metadata for traceability and debugging but must not be exposed in standard API responses.

## 17. Baseline Demo Test Profile

### 17.1 Name

`MVP_BASELINE_TEST_PROFILE`

### 17.2 Purpose

This is the recommended first engineering validation case. It is intentionally smaller than the full 10-20 minute MVP target so the pipeline can be verified before scaling up.

### 17.3 Baseline Clip

- duration: 40-60 seconds
- scene count: 3
- camera behavior: mostly static
- motion: low
- dialogue: light conversation
- cuts: 4 or fewer

### 17.4 Example Scene

Two people sitting at a kitchen table talking.

### 17.5 Example Product

- product_name: sparkling water
- category: beverage

### 17.6 Expected Insertion

- duration: 6 seconds
- interaction: pick up bottle -> sip -> set down
- dialogue: optional short line

### 17.7 Expected Result

- analysis should return 2-3 valid slots
- one slot should be selectable without re-pick
- generation should produce a simple interaction clip
- render should produce a preview clip

### 17.8 Demo Asset Recommendation

`/demo/mvp_baseline_scene.mp4`
