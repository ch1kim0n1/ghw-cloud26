# Technical Specifications (Per Component)

## 1. Scene Detection Service
### Purpose
Detect shot boundaries and aggregate them into scenes to create a timeline structure for candidate ad insertion analysis.

### Input
- video asset reference
- job_id

### Output
- list of shot boundaries
- scene boundary records
- scene timing metadata

### Constraints
- must support asynchronous execution
- should process long-form assets without loading entire video into memory at once
- target high boundary detection accuracy for demo-quality output

### Dependencies
- OpenCV / ffmpeg
- optional ML boundary detection model
- object storage client
- metadata database client

### Failure Modes
- corrupt/unsupported video
- missing storage object
- timeout on long assets

---

## 2. Context Analysis Service
### Purpose
Derive transcript, motion, pacing, and context features needed for slot ranking.

### Input
- video asset reference
- scene list

### Output
- transcript segments
- scene motion scores
- stability scores
- dialogue timing metadata
- contextual summaries/features

### Constraints
- should run independently of slot ranking
- must persist derived artifacts for reuse

### Dependencies
- speech-to-text system
- frame sampling pipeline
- optional vision-language or multimodal model

### Failure Modes
- speech extraction failure
- partial transcript quality issues
- missing scene metadata upstream

---

## 3. Ad Slot Ranking Service
### Purpose
Score and rank the best insertion opportunities.

### Input
- scene boundaries
- context analysis outputs
- campaign constraints

### Output
- ranked slot list
- confidence score per slot
- explanation metadata

### Constraints
- must produce deterministic output for the same input/config when required
- ranking logic must be auditable

### Dependencies
- rules engine
- scoring model or weighted heuristic engine
- metadata database

### Example Signals
- scene stability
- motion intensity
- dialogue gap confidence
- emotional intensity penalty
- product/context fit

---

## 4. Ad Planning Engine
### Purpose
Convert a selected slot and campaign data into a concrete insertion plan.

### Input
- slot_id
- product_id
- campaign config
- source scene context

### Output
- insertion strategy
- prompt package
- anchor frame references
- rendering instructions

### Constraints
- must support multiple insertion modes
- must select fallback mode if generation risk is high

### Dependencies
- slot ranking output
- product asset library
- campaign rules

---

## 5. AI Generation Worker
### Purpose
Generate a short scene-aware ad clip using anchor frames, product references, and prompts.

### Input
- insertion plan
- anchor frame references
- product assets
- scene context payload

### Output
- generated ad clip
- generation metadata
- confidence / quality metrics

### Constraints
- GPU-backed execution only
- clip duration capped by campaign and system limits
- output must be compatible with renderer ingest requirements

### Dependencies
- generative video/image models
- GPU runtime
- object storage

### Failure Modes
- GPU unavailability
- generation artifacts
- invalid output duration/format

---

## 6. Fallback Composer
### Purpose
Produce a lower-risk ad insertion clip when full generation is not viable.

### Input
- insertion plan
- scene context
- product assets

### Output
- composed clip

### Constraints
- should be faster and cheaper than full generation
- should produce a render-compatible output in all normal cases

### Dependencies
- ffmpeg / compositor tooling
- asset templates
- optional image inpainting pipeline

---

## 7. Frame Stitcher / Renderer
### Purpose
Insert the generated/composed clip into the original timeline and export playback-ready media.

### Input
- original video reference
- insertion slot timing
- ad clip reference

### Output
- preview output
- final render

### Constraints
- must preserve codec/container compatibility targets
- should avoid broken timestamps or audio/video drift

### Dependencies
- ffmpeg / media rendering pipeline
- object storage
- metadata database

### Failure Modes
- codec mismatch
- corrupted intermediate clip
- render timeout

---

## 8. Orchestrator Service
### Purpose
Manage state transitions across the pipeline.

### Input
- API commands
- stage completion events
- stage failure events

### Output
- queued tasks
- job state updates
- retry decisions

### Constraints
- idempotent state handling
- audit-friendly job lifecycle

### Dependencies
- queue/event bus
- metadata database

---

## 9. API Service
### Purpose
Expose stable interfaces to dashboard and clients.

### Input
- authenticated HTTP requests

### Output
- JSON responses
- job and asset metadata

### Constraints
- versioned endpoints
- strict schema validation
- structured errors only

### Dependencies
- auth layer
- metadata DB
- object storage signing logic

---

## 10. Dashboard Frontend
### Purpose
Provide operator-facing UI for uploads, job tracking, slot review, and result preview.

### Input
- user interactions
- API responses

### Output
- upload requests
- review/selection actions
- output preview requests

### Constraints
- clear job state visibility
- timeline visualization for candidate slots

### Dependencies
- web framework
- player component
- API client

---

## 11. Shared Cross-Cutting Requirements
### Logging
All services must emit:
- request/job identifiers
- stage transitions
- warnings
- structured errors

### Metrics
Track:
- stage latency
- success/failure counts
- queue depth
- GPU usage
- render completion rate

### Security
- authenticated access to control APIs
- signed asset access
- encrypted storage

### Testing
Each component should have:
- unit tests
- contract tests where applicable
- integration tests for pipeline handoff
