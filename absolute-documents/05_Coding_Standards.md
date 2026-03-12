# Coding Standards (Go Backend + React Frontend)

## 1. Purpose
Define style, naming, patterns, and testing expectations for consistent code during 1-week MVP hackathon sprint.

## 2. General Principles
- **Clarity over cleverness** — readable code is better than smart code
- **Fail fast** — validate inputs early, return descriptive errors
- **No panics** — handle errors gracefully in production code
- **Type safety** — use strict types (Go interfaces, TypeScript strict mode)
- **Keep it simple** — skip complexity until needed

---

## Part 1: Go Backend

### 3.1 Naming Conventions

| Entity | Convention | Example |
|--------|------------|---------|
| Files | snake_case | `campaign_service.go`, `slot_ranker.go` |
| Types | PascalCase | `type Campaign struct {}` |
| Functions | camelCase | `func (sr *SlotRanker) rankSlots()` |
| Methods | camelCase | `func (db *DB) GetCampaign()` |
| Constants | SCREAMING_SNAKE_CASE | `const SCENE_THRESHOLD = 50.0` |
| Variables | camelCase | `campaign := &Campaign{}` |
| Interfaces | PascalCase + "er" suffix | `type Reader interface {}` |
| Unexported | camelCase | `func (sr *SlotRanker) scoreSlot()` (package-private) |

### 3.2 Code Formatting

**Use `gofmt` strictly:**
```bash
gofmt -w ./...
```

**Line length:** ~100 characters (soft limit, don't break readability)

**Brace placement:**
```go
// Good
if len(scenes) == 0 {
  return nil, fmt.Errorf("no scenes")
}

// Bad
if len(scenes) == 0
{
  return nil, fmt.Errorf("no scenes")
}
```

**Error handling:**
```go
// Good: error context
if err != nil {
  return fmt.Errorf("failed to detect scenes in %s: %w", videoPath, err)
}

// Bad: generic error
if err != nil {
  return nil, err
}

// Bad: swallowing error
if err != nil {
  log.Printf("error: %v", err)
  // continuing without handling...
}
```

### 3.3 Database Layer Pattern

**File:** `internal/db/campaigns.go`

```go
package db

// CreateCampaign inserts a new campaign record.
func (db *DB) CreateCampaign(ctx context.Context, c *Campaign) error {
  query := `
    INSERT INTO campaigns (id, name, product_id, video_path, duration_seconds, 
                           target_ad_duration_seconds, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `
  
  _, err := db.conn.ExecContext(ctx, query,
    c.ID, c.Name, c.ProductID, c.VideoPath, c.DurationSeconds, 
    c.TargetAdDurationSeconds, time.Now(),
  )
  
  if err != nil {
    return fmt.Errorf("insert campaign: %w", err)
  }
  
  return nil
}

// GetCampaign retrieves a campaign by ID.
func (db *DB) GetCampaign(ctx context.Context, id string) (*Campaign, error) {
  query := `SELECT id, name, product_id, video_path, duration_seconds, 
                  target_ad_duration_seconds, created_at 
           FROM campaigns WHERE id = ?`
  
  row := db.conn.QueryRowContext(ctx, query, id)
  
  c := &Campaign{}
  err := row.Scan(&c.ID, &c.Name, &c.ProductID, &c.VideoPath, 
                  &c.DurationSeconds, &c.TargetAdDurationSeconds, &c.CreatedAt)
  
  if err == sql.ErrNoRows {
    return nil, fmt.Errorf("campaign not found: %s", id)
  }
  if err != nil {
    return nil, fmt.Errorf("select campaign: %w", err)
  }
  
  return c, nil
}
```

**Rules:**
- Always use parameterized queries (prevent SQL injection)
- Use `ExecContext` and `QueryRowContext` (respect context deadlines)
- Wrap errors with context: `fmt.Errorf("operation_name: %w", err)`
- Never return `sql.ErrNoRows` directly; convert to domain error

### 3.4 Service Layer Pattern

**File:** `internal/services/slot_ranker.go`

```go
package services

type SlotRanker struct {
  logger log.Logger
  db     *db.DB
}

// NewSlotRanker creates a new slot ranker with dependencies.
func NewSlotRanker(logger log.Logger, db *db.DB) *SlotRanker {
  return &SlotRanker{logger: logger, db: db}
}

// RankSlots scores and ranks candidate insertion slots.
// Returns top slots ordered by score descending.
func (sr *SlotRanker) RankSlots(ctx context.Context, scenes []*Scene, product *Product) ([]*Slot, error) {
  sr.logger.Infof("ranking slots for %d scenes", len(scenes))
  
  slots := []*Slot{}
  
  for _, scene := range scenes {
    score := sr.calculateScore(scene, product)
    reasoning := sr.generateReasoning(scene, product, score)
    
    slot := &Slot{
      ID:              generateUUID(),
      SceneID:         scene.ID,
      SceneNumber:     scene.SceneNumber,
      InsertionFrame:  scene.DialogueGapStartFrame,
      SlotType:        "dialogue_gap",
      Confidence:      score,
      Score:           score,
      Reasoning:       reasoning,
    }
    
    slots = append(slots, slot)
  }
  
  // Sort by score descending
  sort.Slice(slots, func(i, j int) bool {
    return slots[i].Score > slots[j].Score
  })
  
  // Assign ranks and return top 5
  for i, slot := range slots {
    if i >= 5 {
      break
    }
    slot.Rank = i + 1
  }
  
  sr.logger.Infof("ranked %d slots, top score: %.2f", len(slots), slots[0].Score)
  return slots[:min(5, len(slots))], nil
}

func (sr *SlotRanker) calculateScore(scene *Scene, product *Product) float64 {
  return scene.StabilityScore*0.4 + 
         (1-scene.MotionScore)*0.3 + 
         sr.dialogueGapConfidence(scene)*0.2 + 
         sr.contextRelevance(scene, product)*0.1
}

func (sr *SlotRanker) dialogueGapConfidence(scene *Scene) float64 {
  if !scene.DialoguePresent {
    return 0.0  // No dialogue = no gap
  }
  gapDuration := (scene.DialogueGapEndFrame - scene.DialogueGapStartFrame) / 24.0  // 24 fps
  return min(1.0, gapDuration/8.0)  // Max at 8+ seconds
}
```

**Rules:**
- Inject dependencies (logger, db) via constructor
- Use contextually rich error messages
- Avoid side effects (e.g., database writes) in pure scoring functions
- Document public methods

### 3.5 HTTP Handler Pattern

**File:** `internal/api/campaigns_handler.go`

```go
package api

func (r *Router) CreateCampaign(w http.ResponseWriter, req *http.Request) {
  // Parse multipart form (5GB max)
  if err := req.ParseMultipartForm(5 << 30); err != nil {
    respondWithError(w, http.StatusBadRequest, "Invalid form data")
    return
  }
  
  // Extract fields
  name := req.FormValue("name")
  productID := req.FormValue("product_id")
  targetAdDuration := req.FormValue("target_ad_duration_seconds")
  
  // Validate
  if name == "" || productID == "" {
    respondWithError(w, http.StatusBadRequest, "Missing required fields: name, product_id")
    return
  }
  
  // Get file
  file, header, err := req.FormFile("video_file")
  if err != nil {
    respondWithError(w, http.StatusBadRequest, "Missing video_file")
    return
  }
  defer file.Close()
  
  // Call service
  campaign, err := r.campaignService.CreateCampaign(req.Context(), name, productID, file)
  if err != nil {
    respondWithError(w, http.StatusInternalServerError, "Failed to create campaign")
    return
  }
  
  respondWithJSON(w, http.StatusCreated, campaign)
}

// Helper functions
func respondWithJSON(w http.ResponseWriter, code int, data interface{}) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(code)
  json.NewEncoder(w).Encode(data)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
  respondWithJSON(w, code, map[string]string{
    "error": message,
  })
}
```

**Rules:**
- Keep handlers thin (delegate to services)
- Validate input early
- Set appropriate HTTP status codes
- Always close file handles
- Use consistent error response format

### 3.6 Testing

**File:** `internal/services/slot_ranker_test.go`

```go
package services

import (
  "testing"
)

func TestSlotRankerRanksScenes(t *testing.T) {
  // Arrange
  ranker := NewSlotRanker(nil, nil)
  scenes := []*Scene{
    {ID: "s1", SceneNumber: 1, StabilityScore: 0.8, MotionScore: 0.2},
    {ID: "s2", SceneNumber: 2, StabilityScore: 0.5, MotionScore: 0.6},
  }
  product := &Product{ID: "p1", Name: "Nike"}
  
  // Act
  slots, err := ranker.RankSlots(context.Background(), scenes, product)
  
  // Assert
  if err != nil {
    t.Fatalf("expected no error, got: %v", err)
  }
  
  if len(slots) == 0 {
    t.Fatal("expected slots, got none")
  }
  
  // Verify top slot is from high-stability scene
  if slots[0].SceneNumber != 1 {
    t.Errorf("expected top slot from scene 1, got: %d", slots[0].SceneNumber)
  }
}

func TestDatabaseCreateReturnsError(t *testing.T) {
  // Mock database that returns error
  mockDB := &mockDB{err: sql.ErrConnDone}
  
  err := mockDB.CreateCampaign(context.Background(), &Campaign{})
  if err == nil {
    t.Fatal("expected error, got nil")
  }
}
```

**MVP Testing Minimum:**
- ✅ Unit tests for business logic (scoring, ranking)
- ✅ Positive path for CRUD
- ❌ Full integration tests
- ❌ Load/stress tests

---

## Part 2: React Frontend

### 4.1 Naming Conventions

| Entity | Convention | Example |
|--------|------------|---------|
| Components | PascalCase.tsx | `Header.tsx`, `SlotCard.tsx` |
| Hooks | useCamelCase.ts | `useJob.ts`, `useFetch.ts` |
| Utilities | camelCase.ts | `api_client.ts` |
| Types/Interfaces | PascalCase | `Campaign`, `JobStatus` |
| Constants | UPPER_SNAKE_CASE | `API_URL`, `POLL_INTERVAL` |
| Functions | camelCase | `formatDate()`, `handleClick()` |
| Variables | camelCase | `jobId`, `isLoading` |

### 4.2 Component Pattern

```tsx
// File: SlotCard.tsx
import React from "react";
import "./SlotCard.css";

interface SlotCardProps {
  slot: Slot;
  isSelected: boolean;
  isLoading?: boolean;
  onSelect: (slotId: string) => void;
}

export const SlotCard: React.FC<SlotCardProps> = ({
  slot,
  isSelected,
  isLoading = false,
  onSelect,
}) => {
  const handleClick = () => {
    onSelect(slot.id);
  };
  
  return (
    <div className={`slot-card ${isSelected ? "selected" : ""}`}>
      <div className="slot-header">
        <h3>Slot #{slot.rank}</h3>
        <span className="score">{(slot.score * 100).toFixed(0)}%</span>
      </div>
      <p className="reasoning">{slot.reasoning}</p>
      <button
        onClick={handleClick}
        disabled={isLoading}
        className="select-button"
      >
        {isLoading ? "Computing..." : "Select"}
      </button>
    </div>
  );
};
```

**Rules:**
- Use `React.FC<Props>` for typing
- Define `interface` for props above component
- Default optional props to sensible values
- Extract callbacks to named functions (not inline)
- No class components

### 4.3 Custom Hook Pattern

```ts
// File: useJob.ts
import { useState, useEffect } from "react";
import { getJobStatus } from "../services/jobs_api";

export function useJob(jobId: string) {
  const [job, setJob] = useState<Job | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    const fetchJob = async () => {
      try {
        const data = await getJobStatus(jobId);
        setJob(data);
        
        // Stop polling if terminal state
        if (["completed", "failed"].includes(data.status)) {
          setLoading(false);
        }
      } catch (err) {
        setError((err as Error).message);
        setLoading(false);
      }
    };
    
    fetchJob();  // Initial fetch
    
    // Poll every 3 seconds
    const interval = setInterval(fetchJob, 3000);
    
    return () => clearInterval(interval);  // Cleanup on unmount
  }, [jobId]);
  
  return { job, loading, error };
}
```

**Rules:**
- Clean up intervals/timers in return function
- Use `useState<Type | null>(null)` for optional data
- Separate initial fetch from polling
- Return object with clear names

### 4.4 API Client Layer

```ts
// File: services/api_client.ts
const API_URL = process.env.REACT_APP_API_URL || "http://localhost:8080";

export async function apiGet<T>(endpoint: string): Promise<T> {
  const response = await fetch(`${API_URL}${endpoint}`, {
    method: "GET",
    headers: { "Content-Type": "application/json" },
  });
  
  if (!response.ok) {
    throw new Error(`API ${response.status}: ${response.statusText}`);
  }
  
  return response.json();
}

export async function apiPost<T>(endpoint: string, body: any): Promise<T> {
  const response = await fetch(`${API_URL}${endpoint}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  
  if (!response.ok) {
    throw new Error(`API ${response.status}: ${response.statusText}`);
  }
  
  return response.json();
}

export async function apiUploadFile(
  endpoint: string,
  file: File,
  otherFields: Record<string, string>
): Promise<any> {
  const form = new FormData();
  form.append("video_file", file);
  Object.entries(otherFields).forEach(([k, v]) => form.append(k, v));
  
  const response = await fetch(`${API_URL}${endpoint}`, {
    method: "POST",
    body: form,
  });
  
  if (!response.ok) {
    throw new Error(`Upload failed: ${response.status}`);
  }
  
  return response.json();
}
```

**Usage Example:**

```ts
// File: services/campaigns_api.ts
import { apiGet, apiPost, apiUploadFile } from "./api_client";

export async function createCampaign(
  name: string,
  productId: string,
  videoFile: File
): Promise<Campaign> {
  return apiUploadFile("/api/campaigns", videoFile, {
    name,
    product_id: productId,
  });
}

export async function listCampaigns(): Promise<Campaign[]> {
  const response = await apiGet<{ campaigns: Campaign[] }>("/api/campaigns");
  return response.campaigns;
}

export async function getCampaignDetail(id: string): Promise<Campaign> {
  return apiGet<Campaign>(`/api/campaigns/${id}`);
}
```

### 4.5 TypeScript Strict Mode

**tsconfig.json:**
```json
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "noImplicitThis": true,
    "alwaysStrict": true
  }
}
```

**No `any` type allowed.** Use generics or union types instead:
```tsx
// Bad
function processData(data: any) { }

// Good
function processData<T>(data: T): T { }

// Good
function handleError(error: Error | string) { }
```

### 4.6 Error Handling

```tsx
export const JobPage: React.FC = () => {
  const jobId = "job_123";
  const { job, loading, error } = useJob(jobId);
  
  if (loading) {
    return <LoadingSpinner />;
  }
  
  if (error) {
    return (
      <div className="error-container">
        <h3>Error Loading Job</h3>
        <p>{error}</p>
        <button onClick={() => window.location.reload()}>Retry</button>
      </div>
    );
  }
  
  if (!job) {
    return <div>Job not found</div>;
  }
  
  return (
    <div>
      <h2>Job {job.id}</h2>
      <p>Status: {job.status}</p>
      <ProgressBar value={job.progress_percent} />
    </div>
  );
};
```

**Rules:**
- Always handle `loading`, `error`, and `null` states
- Show user-friendly error messages
- Provide "Retry" action when appropriate

### 4.7 Styling

Use plain CSS for MVP (no Tailwind, no styled-components):

```css
/* App.css */
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.button {
  padding: 10px 16px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.button:hover {
  background-color: #0056b3;
}

.button:disabled {
  background-color: #ccc;
  cursor: not-allowed;
  opacity: 0.6;
}

.error-message {
  color: #dc3545;
  background-color: #f8d7da;
  padding: 12px;
  border: 1px solid #f5c6cb;
  border-radius: 4px;
  margin: 10px 0;
}

.success-message {
  color: #155724;
  background-color: #d4edda;
  padding: 12px;
  border: 1px solid #c3e6cb;
  border-radius: 4px;
  margin: 10px 0;
}
```

---

## Part 3: Shared Standards

### 5.1 Git Commit Messages

**Format:**
```
[component] Brief imperative description

Optional longer explanation if helpful.

Related: #123
```

**Examples:**
```
[backend] Implement slot ranking with stability signals

Scores slots based on motion + stability.
Deterministic for demo reproducibility.

[frontend] Add job status polling with 3-second interval

Uses custom useJob hook to poll /api/jobs/{id}.
Stops polling on job completion.

[docs] Update API contracts with exact field formats
```

### 5.2 Self-Review Checklist

Push only after:
- [ ] Code follows naming conventions
- [ ] Error messages are descriptive + actionable
- [ ] No hardcoded values (use constants/env vars)
- [ ] Contexts properly handled (Go) and cleaned up (React)
- [ ] Types are explicit (TypeScript `noImplicitAny`, Go interfaces)
- [ ] No secrets/API keys in code
- [ ] Formatted (`gofmt` for Go)
- [ ] Tests added for new logic
- [ ] README updated if adding features

### 5.3 Performance Gotchas

**Go:**
- Avoid allocating large slices in loops
- Use context timeouts for external calls
- Close database connections properly

**React:**
- Don't create new function objects in render (extract to constants)
- Use `useCallback` for handlers in lists
- Memoize expensive re-renders with `React.memo`

---

## Summary

### Quality Targets (MVP)
- ✅ Readable, consistent style
- ✅ Graceful error handling
- ✅ Type-safe (strict TypeScript, Go interfaces)
- ✅ Basic unit tests for critical paths
- ✅ API contracts match docs

### Skip for MVP
- ❌ 100% code coverage
- ❌ E2E test suite
- ❌ Performance profiling
- ❌ Advanced logging
- ❌ CI/CD pipeline

**Remember:** Shipped is better than perfect.
- wrap external dependency failures with project-specific context
- return structured error objects from APIs
- distinguish between user errors, system errors, and transient infra errors

### Structured Error Format
```json
{
  "error": {
    "code": "RENDER_FAILED",
    "message": "Render stage failed",
    "details": {
      "job_id": "...",
      "stage": "rendering"
    }
  }
}
```

## 7. Logging Standards
### Rules
- use structured logs only
- include correlation identifiers in every service boundary log
- never log secrets or raw credentials
- log state transitions explicitly

### Required Context Fields
- `job_id`
- `request_id`
- `service`
- `stage`
- `status`
- `duration_ms` where relevant

## 8. Formatting and Linting
### Python
- formatter: `black`
- import sorting: `isort`
- linting: `ruff` or `flake8`
- typing: `mypy` where practical

### TypeScript
- formatter: `prettier`
- linting: `eslint`
- strict typing enabled

## 9. Comments and Documentation
- comment why, not what, unless the code is non-obvious
- public functions/classes should have concise docstrings where useful
- every service package should include a README if its behavior is non-trivial
- TODO comments must include owner or issue reference where possible

## 10. API Standards
- APIs must be versioned
- every request payload must be schema-validated
- every response must follow explicit contract definitions
- timestamps must use ISO-8601 UTC unless a different format is explicitly required
- pagination and filtering should use consistent query conventions

## 11. Testing Standards
### Minimum Expectations
- unit tests for core domain logic
- integration tests for repositories and service handoffs
- contract tests for API/event compatibility
- e2e tests for critical flows

### Rules
- tests must be deterministic
- avoid network dependence in unit tests
- mock third-party services at boundaries
- use fixtures for representative video/job metadata

## 12. Configuration Standards
- config must come from environment or config files, never hardcoded secrets
- feature flags must be explicit
- environment-specific behavior must be isolated in config layers

## 13. Git and Review Standards
- one concern per PR when possible
- PRs must include test evidence
- breaking contract changes require explicit review callout
- generated files should not be committed unless intentionally versioned

## 14. Security Standards
- no secrets in repository
- signed URL or token-based asset access only
- validate uploaded file types and size limits
- sanitize user-provided metadata before downstream use

## 15. Cloud / Worker Standards
- background jobs must be idempotent where possible
- retryable failures must be marked explicitly
- GPU jobs must emit usage and timing metrics
- all object storage writes must validate success before committing DB state
