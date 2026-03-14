# System Architecture Document

## 1. Scope
This document describes the actual MVP architecture only.

## 2. Canonical MVP Architecture
The MVP uses a local control plane and local storage for operator access, while off-loading analysis, generation, audio work, and final rendering to cloud-backed compute. The current generation path is hybrid: Higgsfield is the primary Phase 3 media generator and Azure ML remains the fallback path.

```text
[React Dashboard]
        |
        v
   [Go REST API]
        |
        +--------------------+
        |                    |
        v                    v
[SQLite Metadata DB]   [Local Asset Filesystem]
        |                    |
        |                    +-----------------------------+
        v                                                  |
[Polling Job Worker]                                       |
        |                                                  |
        +-----------------------+--------------------------+-----------------------+
        |                       |                                                  |
        v                       v                                                  v
[Azure Video Indexer +   [Higgsfield Kling + Azure OpenAI                 [Azure Container Apps
 Azure OpenAI]            (primary) + Azure ML (fallback)]                + Azure AI Speech
  Analysis]               CAFAI Generation                                + Azure Blob Storage]
        |                       |                                                  |
        +-----------------------+--------------------------+-----------------------+
                                                               |
                                                               v
                                            [Preview copied back to local storage]
                                                               |
                                                               v
                                                 [Dashboard Preview / Download]
```

## 3. Design Principles
- keep the control plane simple and local for the hackathon
- push heavy compute to Azure services
- keep job states coarse and understandable
- keep the output model simple: one downloadable preview per job
- preserve generated artifacts when render fails
- do not mix future production-only systems into the MVP design

## 4. Core Components
### 4.1 React Dashboard
Responsibilities:
- create products
- create campaigns
- upload source video and product assets
- explicitly start analysis
- poll job state
- show proposed slots
- allow select, reject, and re-pick actions
- allow suggested product line review and editing
- preview and download the final MP4

### 4.2 Go REST API
Responsibilities:
- validate incoming requests
- save uploaded files locally
- create and read SQLite records
- expose the MVP API surface
- start and monitor background jobs

### 4.3 SQLite Metadata Database
Responsibilities:
- store products, campaigns, jobs, scenes, slots, previews, and job logs
- act as the source of truth for job state and UI queries

### 4.4 Local Asset Filesystem
Responsibilities:
- store uploaded source videos
- store uploaded product assets
- store extracted anchor images and debug artifacts
- store the final downloadable preview MP4

### 4.5 Polling Job Worker
Responsibilities:
- poll for queued jobs and next actions
- submit heavy tasks to Azure services
- persist state transitions
- fail jobs immediately when a required stage fails
- preserve generated artifacts when render fails so render can be retried

This worker is the MVP async mechanism. There is no queue or event bus in MVP.

### 4.6 Analysis Layer
Services:
- Azure Video Indexer
- Azure OpenAI

Responsibilities:
- analyze the full video
- segment scenes
- detect candidate anchor-frame pairs inside scenes
- derive motion, dialogue, narrative, and continuity features
- propose ranked slot candidates with scores and reasoning

### 4.7 CAFAI Generation Layer
Services:
- Higgsfield Kling
- Azure Machine Learning fallback
- Azure OpenAI

Responsibilities:
- generate a short CAFAI clip for one selected slot
- condition generation on start and finish anchors, scene context, and product data
- generate either a short spoken mention or a silent product interaction

### 4.8 Audio and Render Layer
Services:
- Azure AI Speech
- Azure Container Apps running ffmpeg
- Azure Blob Storage

Responsibilities:
- generate or align clip dialogue when needed
- smooth boundary audio with simple crossfade behavior
- use Blob Storage as temporary artifact storage
- render the preview MP4
- copy the final preview back to local storage for download

## 5. Blob and Local Storage Policy
Canonical rule:

Azure Blob Storage is used as temporary cloud artifact storage during generation and rendering. After rendering completes, the final preview file is copied back to local storage for MVP download, inspection, and debugging.

Artifact flow:

```text
generation output
      |
      v
Azure Blob Storage
      |
      v
render worker pulls artifact
      |
      v
preview written to Blob
      |
      v
preview copied to local storage
      |
      v
download and debugging access
```

## 6. Anchor-Frame Model
The MVP does not replace a chunk of source footage. It inserts new runtime between two source frames.

Definitions:
- `anchor_start_frame`: the source frame immediately before the generated CAFAI clip
- `anchor_end_frame`: the source frame immediately after the generated CAFAI clip
- `inserted_duration`: the new 5-8 second CAFAI clip added between those anchors

Result:
- output runtime is longer than source runtime
- source story resumes on the finish anchor frame after the generated clip ends

## 7. Job State Model
The job state model is intentionally coarse:

```text
queued -> analyzing -> generating -> stitching -> completed
queued -> analyzing -> generating -> stitching -> failed
queued -> analyzing -> failed
queued -> generating -> failed
queued -> stitching -> failed
```

### Current Stage Naming
`current_stage` values are lower snake_case. Canonical stage names include:
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

## 8. Failure Mapping
Canonical failure structure:

```text
status = failed
current_stage = <stage>
error_code = <specific_error>
```

Examples:
- no suitable slot found: `status=failed`, `current_stage=slot_selection`, `error_code=NO_SUITABLE_SLOT_FOUND`
- render failure: `status=failed`, `current_stage=render`, `error_code=PREVIEW_RENDER_FAILED`

## 9. End-to-End Data Flow
1. Operator creates or selects a product.
2. Operator creates a campaign and uploads the video.
3. API stores files locally, creates a campaign row, and creates a queued job row.
4. Operator explicitly starts analysis.
5. Polling worker submits the analysis stage to Azure Video Indexer and Azure OpenAI.
6. Analysis returns scenes, candidate slots, and ranking metadata.
7. API exposes the proposed slots in the dashboard.
8. Operator selects a slot, rejects slots, or requests a re-pick.
9. After slot selection, the system prepares a suggested product line.
10. Operator accepts, edits, or disables the line.
11. Worker submits CAFAI generation to Azure Machine Learning and Azure OpenAI.
11. Worker submits CAFAI generation to Higgsfield plus Azure OpenAI, and falls back to Azure ML if the primary provider fails.
12. If generation succeeds, artifacts are written to Azure Blob Storage.
13. Worker submits audio alignment and ffmpeg rendering.
14. Final preview is written to Blob and then copied back to local storage.
15. Dashboard shows preview readiness and allows download.

## 10. Ranking and Re-Pick Rules
Candidate slots must be rejected when they violate any exclusion threshold.

The system:
- returns up to the top 3 valid slots
- returns 1-2 slots when fewer valid slots exist
- fails only when no valid slots exist
- allows up to 2 re-picks after all current proposals are rejected
- keeps ranking rules unchanged during re-picks
- excludes previously rejected slots from re-pick results

## 11. Progress Reporting
The MVP progress bar is stage-based:
- analysis: `40%`
- generation: `40%`
- stitching and rendering: `20%`

## 12. Observability
The MVP should record:
- job state transitions
- slot selection and rejection events
- generation and render durations
- failure reasons
- provider request IDs in internal logs and metadata only

Provider request IDs are not exposed in standard API responses.
