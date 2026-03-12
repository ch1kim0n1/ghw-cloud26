# Coding Standards

## 1. Purpose
Keep the MVP implementation consistent, direct, and aligned with the documented CAFAI workflow.

## 2. General Rules
- prioritize implementation accuracy over pitch language
- document the MVP that actually exists, not a future platform
- keep handlers thin and services focused
- fail fast on invalid input and failed cloud stages
- use one JSON naming style everywhere: snake_case
- preserve one source of truth for job statuses, stage names, slot statuses, and error codes

## 3. Canonical Enum Rules
### Job Status
Job statuses are lower snake_case:
- `queued`
- `analyzing`
- `generating`
- `stitching`
- `completed`
- `failed`

### Current Stage
Current stages are lower snake_case, for example:
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

### Slot Status
Slot statuses are lower snake_case:
- `proposed`
- `selected`
- `rejected`
- `generating`
- `generated`
- `failed`

### Error Code
Error codes are upper snake_case, for example:
- `NO_SUITABLE_SLOT_FOUND`
- `GENERATION_FAILED`
- `PREVIEW_RENDER_FAILED`

## 4. MVP Truth Rules
The following facts must stay consistent across code and docs:
- the MVP strategy name is CAFAI
- the system proposes up to 3 candidate slots automatically
- slot targets are anchor-frame pairs inside scenes
- campaign creation does not auto-start analysis
- the operator may select, reject, and re-pick
- the operator may accept, edit, or disable the product line
- the job state model is `queued -> analyzing -> generating -> stitching -> completed|failed`
- the output is one downloadable preview MP4
- Azure Blob Storage is temporary only; final preview download is local
- there is no fallback generation path in MVP
- provider request IDs stay internal only

## 5. Go Backend Standards
### Naming
- files: snake_case
- exported types: PascalCase
- exported methods: PascalCase
- unexported helpers: camelCase
- constants: UPPER_SNAKE_CASE

### Structure
- `api/` handles HTTP only
- `db/` handles SQL only
- `services/` contains domain logic and Azure calls
- `worker/` advances async jobs and updates state

### Data Access
- use parameterized SQL only
- wrap errors with context
- use direct SQL for MVP; do not add an ORM
- keep SQLite DDL executable as written

### State Handling
- never invent ad-hoc job states
- never hardcode frame rate assumptions
- slot rejection and re-pick behavior must be explicit in code
- preserve generated clip artifacts on render failure

## 6. React Frontend Standards
### Naming
- components: PascalCase.tsx
- hooks: useThing.ts
- service files: camelCase.ts
- types: PascalCase

### UI Behavior
- always show loading, error, and empty states
- polling must stop on terminal job states
- slot cards must clearly show rank, anchors, score, and reasoning
- the UI must make select, reject, and re-pick actions obvious
- the product line review step must appear before generation starts

## 7. API Standards
- MVP API base path is `/api`
- no auth in MVP
- JSON uses snake_case
- error responses use one standard shape
- timestamps use ISO-8601 UTC strings
- standard API responses must not expose provider request IDs

Standard error shape:

```json
{
  "error": "Preview render failed",
  "error_code": "PREVIEW_RENDER_FAILED",
  "http_status": 500,
  "details": {
    "job_id": "job_123"
  },
  "timestamp": "2026-03-12T21:00:00Z"
}
```

## 8. Azure Integration Standards
- Azure calls must be wrapped behind service interfaces or clients
- persist provider request IDs in internal logs and metadata only
- treat failed required cloud stages as terminal for the job
- do not silently swallow provider failures

## 9. Audio and Media Standards
- always use source FPS from the uploaded video
- preserve anchor-frame ordering exactly
- default generated clip duration is 6 seconds
- generated clip duration must not exceed 8 seconds
- output runtime must increase after insertion
- audio transitions must use simple crossfade smoothing at both insertion boundaries

## 10. Testing Standards
Minimum MVP coverage:
- product creation validation
- campaign creation validation
- explicit analysis start path
- job state transitions
- slot ranking and re-pick logic
- product line review modes
- preview render success path
- generation failure path
- render-failure-with-retry path

Do not spend MVP time on:
- load tests
- multi-tenant tests
- auth tests
- production cloud deployment tests

## 11. Documentation Standards
- when MVP decisions change, update all affected docs in the same pass
- remove stale sections instead of appending contradictory replacements
- future ideas belong under clearly marked post-MVP sections only
