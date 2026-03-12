# Repository Structure

## 1. Root Layout
```text
project/
├── backend/
│   ├── api/
│   ├── orchestration/
│   ├── services/
│   │   ├── scene_detection/
│   │   ├── context_analysis/
│   │   ├── slot_ranking/
│   │   ├── ad_planning/
│   │   ├── generation/
│   │   ├── fallback_composer/
│   │   └── rendering/
│   ├── workers/
│   ├── models/
│   ├── schemas/
│   ├── db/
│   ├── utils/
│   └── config/
│
├── frontend/
│   ├── app/
│   ├── components/
│   ├── features/
│   ├── lib/
│   ├── hooks/
│   ├── styles/
│   └── types/
│
├── infrastructure/
│   ├── docker/
│   ├── terraform/
│   ├── k8s/
│   ├── scripts/
│   └── environments/
│
├── shared/
│   ├── contracts/
│   ├── types/
│   ├── constants/
│   └── validators/
│
├── tests/
│   ├── unit/
│   ├── integration/
│   ├── contract/
│   ├── e2e/
│   └── fixtures/
│
├── docs/
│   ├── product/
│   ├── architecture/
│   ├── technical/
│   ├── api/
│   ├── data/
│   └── runbooks/
│
├── .github/
│   └── workflows/
│
├── tools/
├── Makefile
├── README.md
├── package.json
├── pyproject.toml
└── .env.example
```

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
