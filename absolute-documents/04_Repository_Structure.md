# Repository Structure & Development Setup

## Document Purpose
Define the monorepo folder layout, build process, development environment setup, and quick-start guide for a solo Go developer in a 1-week hackathon.

---

## 1. Monorepo Decision

**Decision:** **Monorepo** (single Git repo, backend + frontend folders)

**Rationale:**
- Solo developer: easier single repo management
- Hackathon sprint: simpler deployment
- Shared documentation
- One SQLite database serves both

---

## 2. Repository Structure (Go Backend + React Frontend)

```
ghw-cloud26/
├── README.md
├── .gitignore
├── Makefile                            # Top-level build targets
│
├── absolute-documents/                 # Planning docs (unchanged)
│   ├── 01_Product_Design_Document.md
│   ├── 02_System_Architecture_Document.md
│   ├── 03_Technical_Specifications.md
│   ├── 04_Repository_Structure.md (THIS FILE)
│   ├── 05_Coding_Standards.md
│   ├── 06_API_Contracts.md
│   ├── 07_Data_Schema_Definitions.md
│   └── 08_Task_Decomposition_Plan.md
│
├── backend/                            # Go REST API server
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── Makefile
│   │
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                 # HTTP server entry
│   │
│   ├── internal/
│   │   ├── models/                     # Data structures (match DB schema)
│   │   │   ├── campaign.go
│   │   │   ├── job.go
│   │   │   ├── scene.go
│   │   │   ├── slot.go
│   │   │   ├── render.go
│   │   │   └── product.go
│   │   │
│   │   ├── db/                         # SQLite layer
│   │   │   ├── sqlite.go               # Connection setup
│   │   │   ├── campaigns.go            # CRUD operations
│   │   │   ├── jobs.go
│   │   │   ├── scenes.go
│   │   │   ├── slots.go
│   │   │   └── renders.go
│   │   │
│   │   ├── api/                        # HTTP handlers
│   │   │   ├── router.go               # Gorilla mux setup
│   │   │   ├── middleware.go           # CORS, logging
│   │   │   ├── campaigns_handler.go
│   │   │   ├── jobs_handler.go
│   │   │   ├── slots_handler.go
│   │   │   ├── renders_handler.go
│   │   │   └── errors.go
│   │   │
│   │   ├── services/                   # Business logic
│   │   │   ├── campaign_service.go
│   │   │   ├── job_service.go
│   │   │   ├── scene_detector.go       # OpenCV integration
│   │   │   ├── context_analyzer.go     # Whisper STT
│   │   │   ├── slot_ranker.go
│   │   │   ├── replicate_client.go     # RIFE API client
│   │   │   └── video_stitcher.go       # ffmpeg wrapper
│   │   │
│   │   ├── worker/
│   │   │   └── job_processor.go        # Async job loop
│   │   │
│   │   └── config/
│   │       └── config.go               # ENV vars + settings
│   │
│   ├── scripts/
│   │   ├── init_db.sql                 # SQLite schema
│   │   ├── seed_products.sql           # Demo products
│   │   └── migrations/
│   │       ├── 001_initial_schema.sql
│   │       └── 002_add_json.sql
│   │
│   └── tests/
│       ├── db_test.go
│       ├── services_test.go
│       └── api_test.go
│
├── frontend/                           # React dashboard
│   ├── package.json
│   ├── package-lock.json
│   ├── tsconfig.json
│   ├── .env.example
│   │
│   ├── public/
│   │   ├── index.html
│   │   └── favicon.ico
│   │
│   └── src/
│       ├── index.tsx
│       ├── App.tsx
│       │
│       ├── pages/
│       │   ├── CampaignsPage.tsx       # List campaigns
│       │   ├── CreateCampaignPage.tsx  # Upload + create
│       │   ├── JobDetailPage.tsx       # Job progress
│       │   └── ResultsPage.tsx         # View output
│       │
│       ├── components/
│       │   ├── Header.tsx
│       │   ├── JobStatusCard.tsx
│       │   ├── SlotCard.tsx
│       │   ├── VideoPreview.tsx
│       │   ├── UploadForm.tsx
│       │   └── LoadingSpinner.tsx
│       │
│       ├── hooks/
│       │   ├── useFetch.ts
│       │   ├── useJob.ts               # Poll job status
│       │   └── useSlots.ts
│       │
│       ├── services/
│       │   ├── api_client.ts
│       │   ├── campaigns_api.ts
│       │   ├── jobs_api.ts
│       │   ├── slots_api.ts
│       │   └── renders_api.ts
│       │
│       ├── types/
│       │   ├── Campaign.ts
│       │   ├── Job.ts
│       │   ├── Slot.ts
│       │   └── API.ts
│       │
│       ├── styles/
│       │   ├── App.css
│       │   └── components.css
│       │
│       └── utils/
│           ├── formatters.ts
│           ├── validators.ts
│           └── constants.ts            # API URLs
│
├── infra/
│   ├── docker/
│   │   ├── Dockerfile.backend
│   │   ├── Dockerfile.frontend
│   │   └── docker-compose.yml
│   │
│   └── azure/
│       └── deployment_guide.md
│
├── .env.example
├── docker-compose.yml                  # Optional local dev
└── ghw_cloud26.db                      # SQLite (git-ignored)
```

---

## 3. Backend Structure Details

### `backend/internal/models/`
Go structs matching database schema:
```go
// campaign.go
type Campaign struct {
  ID                       string
  Name                     string
  ProductID                string
  VideoPath                string
  TargetAdDurationSeconds  int
  CreatedAt                time.Time
}

// job.go
type Job struct {
  ID              string
  CampaignID      string
  Status          string  // queued, analyzing, generating, stitching, completed, failed
  ProgressPercent int
  CurrentStage    string
  MetadataJSON    string
  CreatedAt       time.Time
}
```

### `backend/internal/db/`
Direct SQL operations (no ORM):
```go
// campaigns.go
func (db *DB) CreateCampaign(ctx context.Context, c *Campaign) error
func (db *DB) GetCampaign(ctx context.Context, id string) (*Campaign, error)
func (db *DB) ListCampaigns(ctx context.Context, limit, offset int) ([]*Campaign, error)

// jobs.go
func (db *DB) CreateJob(ctx context.Context, j *Job) error
func (db *DB) UpdateJobStatus(ctx context.Context, jobID, status string) error
func (db *DB) GetJob(ctx context.Context, id string) (*Job, error)
```

### `backend/internal/api/`
HTTP handlers exposing REST endpoints:
```go
// campaigns_handler.go
func (r *Router) CreateCampaign(w http.ResponseWriter, req *http.Request)
func (r *Router) GetCampaign(w http.ResponseWriter, req *http.Request)
func (r *Router) ListCampaigns(w http.ResponseWriter, req *http.Request)

// slots_handler.go
func (r *Router) ListSlots(w http.ResponseWriter, req *http.Request)
func (r *Router) SelectSlot(w http.ResponseWriter, req *http.Request)
```

### `backend/internal/services/`
Business logic + integrations:
```go
// scene_detector.go
func (s *SceneDetector) DetectScenes(videoPath string) ([]*Scene, error)
  // Uses OpenCV

// context_analyzer.go
func (ca *ContextAnalyzer) Analyze(videoPath string) (*ContextInfo, error)
  // Uses Whisper STT

// slot_ranker.go
func (sr *SlotRanker) RankSlots(scenes []*Scene, context *ContextInfo) ([]*Slot, error)
  // Scoring logic

// replicate_client.go
func (rc *ReplicateClient) InterpolateFrames(startFrame, endFrame string) (string, error)
  // Calls Replicate RIFE API

// video_stitcher.go
func (vs *VideoStitcher) StitchVideo(originalPath, adPath string, insertFrame int) (string, error)
  // Calls ffmpeg
```

### `backend/internal/worker/`
Async job processing:
```go
// job_processor.go
func (jp *JobProcessor) Start()
  // Polls for queued jobs
  // Runs: scene detection → context analysis → slot ranking
  // Updates job status in database
```

---

## 4. Frontend Structure Details

### `frontend/src/pages/`
Route-level components:
```tsx
// CampaignsPage.tsx
- List all campaigns
- "New Campaign" button

// CreateCampaignPage.tsx
- Video upload form
- Product selection dropdown
- Submit → POST /api/campaigns

// JobDetailPage.tsx
- Poll GET /api/jobs/{id}
- Display progress bar
- Show ranked slots once analysis complete
- Allow slot selection

// ResultsPage.tsx
- Display selected slot info
- Show generated ad preview (after RIFE completes)
- Show final stitched video
- Download button
```

### `frontend/src/services/`
API client layer (wraps fetch):
```ts
// api_client.ts
async function post<T>(endpoint: string, data: any): Promise<T>
async function get<T>(endpoint: string): Promise<T>
async function uploadFile(endpoint: string, file: File, formData: any): Promise<any>

// campaigns_api.ts
export async function createCampaign(name, productId, videoFile)
export async function listCampaigns()
export async function getCampaign(id)

// jobs_api.ts
export async function getJob(jobId)
export async function getJobLogs(jobId)

// slots_api.ts
export async function listSlots(jobId)
export async function selectSlot(jobId, slotId)
```

### `frontend/src/hooks/`
Custom React hooks for common patterns:
```ts
// useJob.ts
- Polls job status every 3 seconds
- Returns: job, loading, error

// useSlots.ts
- Fetches + caches slots
- Returns: slots, selectedSlot, selectSlot()

// useFetch.ts
- Generic data fetching hook
- Handles loading, error states
```

---

## 5. Database

**File Location:** `ghw_cloud26.db` in repo root

**Initialization:**
```bash
sqlite3 ghw_cloud26.db < backend/scripts/init_db.sql
sqlite3 ghw_cloud26.db < backend/scripts/seed_products.sql
```

**Tables:** Campaign, Product, Job, Scene, Slot, Render, JobLog

---

## 6. Build Targets

### Top-level Makefile:
```makefile
.PHONY: setup dev build test clean

setup:
	sqlite3 ghw_cloud26.db < backend/scripts/init_db.sql
	cd frontend && npm install

dev:
	# Start backend + frontend in parallel
	(cd backend && go run cmd/server/main.go) & \
	(cd frontend && npm start)

build:
	cd backend && go build -o bin/server cmd/server/main.go
	cd frontend && npm run build

test:
	cd backend && go test ./...
	cd frontend && npm run test

clean:
	rm -f ghw_cloud26.db
	cd backend && go clean
	cd frontend && rm -rf node_modules
```

---

## 7. Environment Variables

### Backend (.env or shell export):
```bash
PORT=8080
DATABASE_PATH=ghw_cloud26.db
REPLICATE_API_KEY=<key>
UPLOAD_DIR=/tmp/uploads
OUTPUT_DIR=/tmp/outputs
LOG_LEVEL=info
```

### Frontend (.env.local):
```bash
REACT_APP_API_URL=http://localhost:8080
REACT_APP_POLL_INTERVAL=3000
```

---

## 8. Local Development

```bash
# 1. Clone
git clone <repo>

# 2. Setup database
make setup

# 3. Start backend (terminal 1)
cd backend && go run cmd/server/main.go

# 4. Start frontend (terminal 2)
cd frontend && npm install && npm start

# 5. Open browser
http://localhost:3000
```

---

## 9. Git Ignore

```
ghw_cloud26.db
/backend/bin/
/frontend/node_modules/
/frontend/build/
.env.local
*.log
/tmp/
```

---

## 10. Deployment (Not for MVP, but keep in mind)

**Future:** Azure Container Instances or App Service with Postgres + Blob Storage.

For MVP: Just run locally + demo.

---

## 11. Key Decisions Locked

✅ Monorepo (single Git repo)
✅ Go + React
✅ SQLite (local database)
✅ Local filesystem (videos in /tmp/)
✅ Replicate API (RIFE inference)
✅ ffmpeg (video stitching)
✅ Whisper (speech-to-text, local)
✅ OpenCV (scene detection)

## 2. Backend Directory Breakdown
### `backend/api/`
Contains:
- route handlers
- controllers
- request/response mapping
- authentication middleware

### `backend/orchestration/`
Contains:
- workflow state logic
- queue dispatchers
- retry policies
- job lifecycle handlers

### `backend/services/`
Contains business logic grouped by bounded component:
- `scene_detection/`
- `context_analysis/`
- `slot_ranking/`
- `ad_planning/`
- `generation/`
- `fallback_composer/`
- `rendering/`

Each service should contain:
- service logic
- adapters
- interfaces
- tests close to implementation where useful

### `backend/workers/`
Contains:
- async job consumers
- GPU worker entrypoints
- queue-triggered processors

### `backend/models/`
Contains:
- internal domain models
- DTOs if not shared globally

### `backend/schemas/`
Contains:
- request validation schemas
- event payload schemas
- persistence schemas when needed

### `backend/db/`
Contains:
- migrations
- seed data
- repositories
- database session/configuration

### `backend/utils/`
Contains:
- logging helpers
- media utilities
- time helpers
- file helpers

### `backend/config/`
Contains:
- environment config loaders
- service config objects
- feature flags

## 3. Frontend Directory Breakdown
### `frontend/app/`
App/router shell, route entry points, top-level layouts.

### `frontend/components/`
Reusable UI primitives and shared widgets.

### `frontend/features/`
Feature-specific UI and state:
- uploads
- jobs
- slot review
- render preview
- campaign config

### `frontend/lib/`
API clients, auth utilities, player integrations.

### `frontend/hooks/`
Reusable hooks for fetching, polling, state handling.

### `frontend/styles/`
Global styles, tokens, theme files.

### `frontend/types/`
Client-side types not owned by shared package.

## 4. Infrastructure Directory Breakdown
### `infrastructure/docker/`
Dockerfiles and local compose setup.

### `infrastructure/terraform/`
Cloud resources:
- storage
- compute
- queues
- databases
- networking

### `infrastructure/k8s/`
Kubernetes manifests or Helm charts if used.

### `infrastructure/scripts/`
Deployment helpers, migration scripts, bootstrap tools.

### `infrastructure/environments/`
Environment overlays/configs for:
- local
- dev
- staging
- prod

## 5. Shared Package
### `shared/contracts/`
Canonical API/event contracts used across backend and frontend.

### `shared/types/`
Shared TypeScript/Python-generated types if applicable.

### `shared/constants/`
Shared enums, limits, and fixed values.

### `shared/validators/`
Reusable schema validators.

## 6. Tests Layout
### `tests/unit/`
Fast isolated tests for functions/classes.

### `tests/integration/`
Cross-module tests with real DB/storage/queues where practical.

### `tests/contract/`
API and event schema compatibility tests.

### `tests/e2e/`
Upload → process → preview full-path tests.

### `tests/fixtures/`
Sample clips, metadata fixtures, fake products, mock events.

## 7. Placement Rules
- route definitions belong in `backend/api/`
- long-running processors belong in `backend/workers/`
- domain logic belongs in `backend/services/`
- persistence logic belongs in `backend/db/`
- shared schemas/types must not be duplicated across frontend/backend
- infrastructure code must not be mixed into application logic folders
- docs must stay under `docs/`, not buried in source folders

## 8. Branching/Scalability Benefit
This layout allows the project to scale from hackathon MVP to a real product because it separates:
- product-facing UI
- stable API boundaries
- async processing logic
- GPU-heavy workers
- infrastructure definitions
- test coverage layers
