# ghw-cloud26
```
   /\_/\  
  ( o.o )   ghw-cloud26
  > ^ <    Cloud-Assisted Contextual Ad Insertion
```

A cloud-assisted contextual ad insertion system built during **MLH Global Hack Week**.

---

# ♡ Development

```
   (\_/)
   ( •_•)   Getting started
  / >🍪
```

### Backend

- **Go version:** `1.25+`
- **Runtime dependencies:** `ffprobe`, `ffmpeg` on `PATH`
- **Tests require:** `ffmpeg`, `ffprobe` on `PATH`

Install ffmpeg on macOS:

```bash
brew install ffmpeg
```

Run backend server from repo root:

```bash
go run ./backend/cmd/server
```

Run backend tests:

```bash
cd backend
go test ./...
```

---

### Frontend

- **Node.js + npm required**

Install dependencies:

```bash
cd frontend
npm install
```

Run development server:

```bash
npm run dev
```

Run tests:

```bash
npm run test
```

Production build:

```bash
npm run build
```

---

# ♡ Overview

```
 /\_/\ 
( •.• )   What is this?
 > ^ <
```

This repository contains the **MVP documentation and implementation plan** for a cloud-assisted contextual ad insertion system.

### Product Strategy

**Context-Aware Fused Ad Insertion (CAFAI)**

### MVP Goal

The system will:

- analyze a provided **H.264 MP4**
- automatically propose valid **insertion slots**
- allow the operator to **select or edit the product line**
- generate a **short contextual bridge clip**
- stitch that clip into the original video
- export a **downloadable preview MP4**

---

# ♡ MVP Summary

```
  /\_/\  
 ( •ω• )   MVP Contract
  > ^ <
```

Engineering expectations:

- supported videos: **10–20 minute H.264 MP4**
- system proposes **top 3 insertion slots**
- operator may reject slots and request **up to 2 re-picks**
- CAFAI generation produces **5–8 second bridge clip**
- final preview is served from **local storage**
- **Azure Blob Storage** used for temporary generation artifacts

### Job State Machine

```
queued
  ↓
analyzing
  ↓
generating
  ↓
stitching
  ↓
completed | failed
```

---

# ♡ Current State

```
  (\_/)
  ( •_•)   Implementation Status
 / >🛠
```

### Implemented

- **Phase 0:** foundation and runtime bootstrap
- **Phase 1:** product and campaign ingest
- **Phase 2:** analysis + slot proposal
- **Phase 3:** CAFAI generation workflow
- **Phase 4:** preview rendering pipeline

### Runtime Requirements

Phases **2–4 require provider configuration** before backend startup.

Additional requirements:

- Phase 3 requires local `ffmpeg`
- Phase 4 requires **Blob/Object Storage** + render service
- local **SQLite**, uploads, and runtime directories remain the MVP control plane

### Deferred

```
Phase 5 → Demo hardening and production reliability
```

---

# ♡ Documentation

```
  /\_/\  
 ( •.• )   Docs
  > ^ <
```

Core engineering docs live in:

```
absolute-documents/
```

### Recommended Reading Order

1. `01_Product_Design_Document.md`
2. `02_System_Architecture_Document.md`
3. `03_Technical_Specifications.md`
4. `06_API_Contracts.md`
5. `07_Data_Schema_Definitions.md`
6. `08_Task_Decomposition_Plan.md`
7. `10_Phase_4_Gap_Assessment.md`

### Document Purposes

| Document | Purpose |
|--------|--------|
| 01_Product_Design_Document | Canonical MVP product behavior |
| 02_System_Architecture_Document | Control plane and job lifecycle |
| 03_Technical_Specifications | Implementation rules |
| 04_Repository_Structure | Monorepo layout |
| 05_Coding_Standards | Naming, API, testing |
| 06_API_Contracts | REST payload definitions |
| 07_Data_Schema_Definitions | SQLite schema |
| 08_Task_Decomposition_Plan | MVP build phases |
| 10_Phase_4_Gap_Assessment | Current Phase 4 status |

---

# ♡ Architecture At A Glance

```
   /\_/\  
  ( •.• )   System Overview
   > ^ <
```

### Core Stack

- **React dashboard** for operator workflow
- **Go REST API** as local control plane
- **SQLite** metadata store
- **Local filesystem** for uploads + preview download
- **Polling worker** for async jobs
- **Azure services** for AI + rendering

### Azure Services

- **Azure Video Indexer + Azure OpenAI** → analysis
- **Azure Machine Learning + Azure OpenAI** → CAFAI generation
- **Azure AI Speech** → audio generation
- **Azure Container Apps** → final rendering
- **Azure Blob Storage** → temporary artifacts

---

# ♡ Canonical Job Flow

```
(1) create product
        ↓
(2) create campaign + upload video
        ↓
(3) start analysis
        ↓
(4) review insertion slots
        ↓
(5) select slot + edit product line
        ↓
(6) generate CAFAI clip
        ↓
(7) render preview
        ↓
(8) download preview
```

---

# ♡ Repository Layout

```
 /\_/\  
( •.• )   Monorepo structure
 > ^ <
```

Important directories:

```
backend/cmd/server        → Go server entrypoint
backend/internal/api      → HTTP handlers
backend/internal/db       → SQLite + migrations
backend/internal/models   → domain models
backend/internal/services → Azure service clients
backend/internal/worker   → async worker

backend/scripts/migrations → SQL migrations

frontend/src/pages        → dashboard pages
frontend/src/services     → API client
frontend/src/types        → type contracts

tmp/                      → runtime uploads + previews
```

---

# ♡ API Surface

Base path:

```
/api
```

### Live Endpoints

- `GET /api/health`

### Implemented Routes

- products
- campaigns
- jobs
- analysis start
- slot review
- slot select
- slot reject
- slot re-pick
- slot generation
- preview render
- preview status
- preview streaming
- preview download

### Standard Error Envelope

```json
{
  "error": "",
  "error_code": "",
  "http_status": "",
  "details": "",
  "timestamp": ""
}
```

---

# ♡ Data Model

```
  /\_/\  
 ( •.• )   SQLite Schema
  > ^ <
```

Core tables:

- `products`
- `campaigns`
- `jobs`
- `scenes`
- `slots`
- `job_previews`
- `job_logs`

Key schema rules:

- IDs are **application-generated**
- local file paths intentionally stored
- provider request IDs kept internal
- slot anchors use **source video FPS**
- render failures preserve artifacts

---

# ♡ MVP Build Phases

```
Phase 0 → Foundation
Phase 1 → Product + Campaign Ingest
Phase 2 → Analysis + Slot Proposal
Phase 3 → CAFAI Generation
Phase 4 → Preview Rendering
Phase 5 → Demo Hardening
```

Rule:

> Complete the **end-to-end MVP** before any post-MVP work.

---

# ♡ Artifact Flow

```
generation output
      │
      ▼
Azure Blob Storage (temporary)
      │
      ▼
render worker pulls artifact
      │
      ▼
final preview written to Blob
      │
      ▼
preview copied to local storage
      │
      ▼
download + debugging access
```

---

# ♡ Phase Status

```
 /\_/\  
( •.• )   Progress
 > ^ <
```

### Phase 0

- runnable Go backend
- SQLite migrations
- Azure client wiring
- polling worker
- React + TypeScript dashboard
- `/api/health`

### Phase 1

- reusable product creation
- campaign creation
- video validation
- inline product creation
- job creation

### Phase 2

- explicit analysis start
- worker polling
- scene + slot persistence
- slot ranking
- dashboard slot review

### Phase 3

- slot selection
- anchor frame extraction
- product-line editing modes
- CAFAI generation
- artifact metadata persistence

### Phase 4

- preview rendering
- cloud artifact upload
- render polling
- preview download + playback
- dashboard preview actions

Remaining work tracked in:

```
absolute-documents/10_Phase_4_Gap_Assessment.md
```

---

```
   /\_/\  
  ( •.• )  thanks for reading
   > ^ <
```

