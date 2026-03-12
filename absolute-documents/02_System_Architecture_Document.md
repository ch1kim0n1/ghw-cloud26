# System Architecture Document

## 1. Overview
The system is a cloud-native media processing pipeline that ingests long-form video, analyzes it for low-disruption ad insertion opportunities, generates or composes a short context-aware ad segment, and renders a stitched output.

## 2. High-Level Architecture
```text
[Web Dashboard / Client]
          |
          v
     [API Gateway]
          |
          v
   [Orchestrator Service] -------------------------------+
          |                                              |
          v                                              v
   [Object Storage]                              [Metadata Database]
          |
          v
   [Job Queue / Event Bus]
          |
  +-------+--------+------------------+------------------+
  |                |                  |                  |
  v                v                  v                  v
[Scene          [Context          [Ad Slot          [Asset / Campaign
Detection]      Analysis]         Ranking]          Validation]
  |                |                  |                  |
  +----------------+---------> [Ad Planning Engine] <----+
                                   |
                         +---------+----------+
                         |                    |
                         v                    v
               [AI Generation Worker]   [Fallback Composer]
                         |                    |
                         +---------+----------+
                                   |
                                   v
                         [Frame Stitcher / Renderer]
                                   |
                                   v
                           [Preview + Final Output]
                                   |
                                   v
                           [Playback / Review UI]
```

## 3. Architectural Style
The system uses:
- event-driven orchestration
- service-oriented boundaries
- asynchronous background processing for heavy media jobs
- cloud object storage for large binary assets
- metadata persistence in a transactional database
- isolated GPU workers for expensive generation stages

## 4. Core Components

### 4.1 Web Dashboard / Client
**Responsibilities**
- upload source video and product assets
- configure campaign parameters
- monitor job progress
- review ranked insertion moments
- preview before/after results

**Inputs**
- media assets
- campaign settings
- user commands

**Outputs**
- API requests
- asset upload requests
- review actions

---

### 4.2 API Gateway
**Responsibilities**
- authentication and authorization
- request validation
- routing to backend services
- rate limiting and API versioning

**Inputs**
- HTTPS requests from dashboard or clients

**Outputs**
- routed requests to internal services
- structured API responses

---

### 4.3 Orchestrator Service
**Responsibilities**
- create and track processing jobs
- manage workflow state transitions
- dispatch work to queues
- coordinate retries and fallbacks

**Inputs**
- new asset/campaign submissions
- service completion events
- failure events

**Outputs**
- job records
- work queue events
- status updates

---

### 4.4 Object Storage
**Responsibilities**
- store original videos
- store extracted frames, transcripts, and derived artifacts
- store generated ad segments and final renders

**Inputs**
- uploaded files
- intermediate artifacts
- rendered outputs

**Outputs**
- signed URLs or internal object references

---

### 4.5 Metadata Database
**Responsibilities**
- persist jobs, scenes, candidates, campaigns, products, outputs
- support auditability and UI retrieval
- preserve workflow state

**Inputs**
- structured metadata from all services

**Outputs**
- queryable state for services and UI

---

### 4.6 Job Queue / Event Bus
**Responsibilities**
- decouple services
- enable asynchronous processing
- support retries and fan-out

**Inputs**
- workflow tasks from orchestrator

**Outputs**
- delivery of work items to services

---

### 4.7 Scene Detection Service
**Responsibilities**
- identify shot boundaries
- cluster shots into scenes
- output scene boundary timeline

**Inputs**
- source video asset reference

**Outputs**
- scene boundary records
- shot metadata

---

### 4.8 Context Analysis Service
**Responsibilities**
- extract transcript and timing
- estimate motion, pacing, scene stability
- derive narrative/context features

**Inputs**
- source video
- scene boundaries

**Outputs**
- transcript segments
- scene context metadata
- timing and suitability features

---

### 4.9 Ad Slot Ranking Service
**Responsibilities**
- score candidate insertion moments
- rank top opportunities
- attach explanation metadata

**Inputs**
- scene boundary data
- context analysis data
- campaign constraints

**Outputs**
- ranked candidate slots
- slot explanations

---

### 4.10 Ad Planning Engine
**Responsibilities**
- choose insertion strategy
- bind campaign/product to candidate slot
- generate prompts/plans for generation or composition path

**Inputs**
- selected slot
- campaign metadata
- product assets

**Outputs**
- insertion plan
- generation/composition instructions

---

### 4.11 AI Generation Worker
**Responsibilities**
- produce short context-aware ad segment using GPU-enabled workflows
- use anchor frames, prompts, and product assets

**Inputs**
- insertion plan
- source context frames
- product references

**Outputs**
- generated clip
- quality/confidence metadata

---

### 4.12 Fallback Composer
**Responsibilities**
- produce a lower-risk ad clip when AI generation fails or is too expensive
- use overlays, environment inserts, or stylized bridge clips

**Inputs**
- insertion plan
- scene metadata
- assets

**Outputs**
- composed clip

---

### 4.13 Frame Stitcher / Renderer
**Responsibilities**
- splice generated/composed ad into timeline
- align timing
- encode preview/final output
- preserve continuity as much as possible

**Inputs**
- original video reference
- selected slot
- generated/composed clip

**Outputs**
- preview render
- final rendered asset

## 5. Data Flow
### End-to-End Flow
1. User uploads movie/show clip and campaign assets.
2. API creates job via orchestrator.
3. Assets are stored in object storage.
4. Orchestrator emits scene analysis tasks.
5. Scene Detection Service outputs scene boundaries.
6. Context Analysis Service extracts transcript/timing/context features.
7. Slot Ranking Service scores candidate moments.
8. Operator or automation selects one of the top-ranked slots.
9. Ad Planning Engine chooses insertion mode.
10. AI Generation Worker or Fallback Composer creates ad clip.
11. Frame Stitcher inserts clip and renders output.
12. Results are stored and exposed in dashboard.

## 6. Service Boundaries
### Control Plane
- API Gateway
- Orchestrator Service
- Dashboard
- Metadata DB

### Data Plane
- Object Storage
- Queue / Event Bus
- Scene Detection
- Context Analysis
- Slot Ranking
- Generation / Composition
- Rendering

This separation prevents UI/API concerns from being tightly coupled to compute-heavy processing.

## 7. API Surface (High-Level)
- asset upload APIs
- campaign creation APIs
- job management APIs
- slot review APIs
- generation/render trigger APIs
- output retrieval APIs

Detailed contracts are defined in the API Contracts document.

## 8. Scalability Strategy
- horizontal scaling for scene/context analysis workers
- autoscaling queue consumers
- isolated GPU pools for generation workers
- independent retry logic for failed stages
- storage-backed artifact passing instead of large in-memory transfers

## 9. Reliability Strategy
- idempotent job stages
- stage-level retries
- persisted workflow state
- fallback composition path when generation fails
- structured error reporting per component

## 10. Security Considerations
- signed uploads/downloads
- authenticated dashboard access
- per-job access control
- encryption at rest for assets and metadata
- audit logging for processing actions

## 11. Observability
- centralized logs
- metrics per stage
- queue depth monitoring
- render success/failure rates
- latency per pipeline stage
- cost tracking for GPU jobs
