# Technical Specifications (MVP Implementation Details)

## Document Purpose
Define specific tools, libraries, APIs, and implementation details for each component. This is the "how to implement" guide for the solo Go developer.

---

## 1. Scene Detection Service

### Purpose
Detect shot boundaries and segment video into scenes.

### Input
- `video_path` (string): Local filesystem path to MP4
- `job_id` (string): For logging + database updates

### Output
- List of `Scene` objects inserted into `scenes` table:
  ```
  scene_number: 1-N
  start_frame: 0
  end_frame: 300
  duration_seconds: 12.5
  motion_score: 0.0-1.0 (calculated later)
  ```

### Implementation Details

**Library:** OpenCV (C++ bindings for Go)

```go
import "gocv.io/x/gocv"

func (sd *SceneDetector) DetectScenes(videoPath string) ([]*Scene, error) {
  video, err := gocv.VideoCaptureFile(videoPath)
  defer video.Close()
  
  fps := video.Get(gocv.VideoCaptureFPS)
  totalFrames := int(video.Get(gocv.VideoCaptureFrameCount))
  
  // Simple boundary detection: detect cuts via histogram difference
  // More sophisticated: use pre-trained ML model if available
  
  scenes := []*Scene{}
  prevHist := gocv.NewMat()
  sceneStart := 0
  
  for frameNum := 0; frameNum < totalFrames; frameNum++ {
    mat := gocv.NewMat()
    if !video.Read(&mat) { break }
    
    // Convert to HSV for robustness
    hsv := gocv.NewMat()
    gocv.CvtColor(mat, &hsv, gocv.ColorBGRtoHSV)
    
    // Compute histogram
    hist := gocv.NewMat()
    gocv.CalcHist([]gocv.Mat{hsv}, []int{0}, gocv.NewMat(), &hist, []int{256}, []float32{0, 256}, false)
    
    // Compare with previous frame
    if frameNum > 0 {
      diff := gocv.CompareHist(prevHist, &hist, gocv.HistCompChiSqrt)
      if diff > SCENE_THRESHOLD {  // Detect cut
        // End current scene
        scenes = append(scenes, &Scene{
          SceneNumber:    len(scenes) + 1,
          StartFrame:     sceneStart,
          EndFrame:       frameNum - 1,
          DurationSeconds: float64(frameNum-sceneStart) / fps,
        })
        sceneStart = frameNum
      }
    }
    prevHist = hist.Clone()
    mat.Close()
    hsv.Close()
    hist.Close()
  }
  
  return scenes, nil
}
```

**Constants:**
```go
const SCENE_THRESHOLD = 50.0  // Tune based on test videos
```

### Failure Modes
- Video codec not supported: Return error, job status → failed
- Corrupt video: Return error
- Out of memory for long videos: Process frame-by-frame (already done above)

### Database Insert
After detection, insert all scenes:
```go
for _, scene := range scenes {
  db.InsertScene(ctx, scene)
}
db.UpdateJobStatus(ctx, jobID, "analyzing")  // Move to next stage
```

---

## 2. Context Analysis Service

### Purpose
Extract dialogue timing, motion intensity, and scene descriptions.

### Output
For each scene, populate:
- `motion_score` (0.0-1.0): How much motion in scene
- `stability_score` (0.0-1.0): How stable/static
- `dialogue_present` (bool): Speech detected
- `dialogue_gap_start_frame`, `dialogue_gap_end_frame`: Quiet moments
- `scene_description` (string): Text summary

### Implementation Details

#### 2.1 Speech-to-Text (Dialogue Detection)

**Tool:** Whisper (OpenAI open-source)

**Go Library:**
```go
import "github.com/go-echarts/go-echarts/v2/opts"
// Actually, for Whisper: use subprocess call or Python binding
```

**Implementation** (subprocess):
```go
import "os/exec"

func (ca *ContextAnalyzer) ExtractAudio(videoPath string) (string, error) {
  // Use ffmpeg to extract audio
  audioPath := "/tmp/audio_temp.wav"
  cmd := exec.Command("ffmpeg",
    "-i", videoPath,
    "-q:a", "9",
    "-n",
    audioPath,
  )
  err := cmd.Run()
  return audioPath, err
}

func (ca *ContextAnalyzer) TranscribeWithWhisper(audioPath string) (*Transcript, error) {
  // Call Whisper command-line tool
  // whisper audio.wav --output_format json --model base
  cmd := exec.Command("whisper",
    audioPath,
    "--output_format", "json",
    "--model", "base",  // or "small" for accuracy
    "--output_dir", "/tmp/transcripts",
  )
  err := cmd.Run()
  
  // Parse JSON output
  jsonPath := "/tmp/transcripts/audio.json"
  result := &Transcript{}
  // Unmarshal JSON
  return result, nil
}

type Transcript struct {
  Segments []TranscriptSegment `json:"segments"`
}

type TranscriptSegment struct {
  ID    int     `json:"id"`
  Start float64 `json:"start"`  // seconds
  End   float64 `json:"end"`
  Text  string  `json:"text"`
}
```

**Alternative:** Use Azure Cognitive Services API if Whisper setup is problematic.

#### 2.2 Motion & Stability Scoring

```go
func (ca *ContextAnalyzer) AnalyzeMotion(videoPath string, scene *Scene) error {
  video, _ := gocv.VideoCaptureFile(videoPath)
  defer video.Close()
  
  fps := video.Get(gocv.VideoCaptureFPS)
  
  var opticalFlows []float64
  var prevGray gocv.Mat
  
  for frameNum := scene.StartFrame; frameNum <= scene.EndFrame; frameNum++ {
    video.Set(gocv.VideoCaptureFrameID, float64(frameNum))
    frame := gocv.NewMat()
    video.Read(&frame)
    
    gray := gocv.NewMat()
    gocv.CvtColor(frame, &gray, gocv.ColorBGRtoGray)
    
    if frameNum > scene.StartFrame {
      // Calculate optical flow (motion)
      flow := gocv.NewMat()
      gocv.CalcOpticalFlowFarneback(prevGray, gray, &flow, 0.5, 3, 15, 3, 5, 1.2, 0)
      
      // Average magnitude of flow vectors
      avgMotion := calculateAverageFlow(&flow)
      opticalFlows = append(opticalFlows, avgMotion)
      flow.Close()
    }
    
    prevGray = gray.Clone()
    frame.Close()
  }
  
  // Normalize scores 0-1
  scene.MotionScore = normalizeScore(opticalFlows, 0.0, 30.0)
  scene.StabilityScore = 1.0 - scene.MotionScore
  
  return nil
}
```

#### 2.3 Dialogue Gap Detection

Using Whisper transcript:
```go
func (ca *ContextAnalyzer) FindDialogueGaps(scene *Scene, transcript *Transcript) error {
  fps := 24.0  // Standard
  
  // Find gaps in transcript within scene time
  sceneStartSec := float64(scene.StartFrame) / fps
  sceneEndSec := float64(scene.EndFrame) / fps
  
  sceneSegments := filterSegments(transcript.Segments, sceneStartSec, sceneEndSec)
  
  if len(sceneSegments) == 0 {
    // No dialogue, entire scene is a dialogue gap
    scene.DialoguePresent = false
    scene.DialogueGapStartFrame = scene.StartFrame
    scene.DialogueGapEndFrame = scene.EndFrame
    return nil
  }
  
  // Find longest gap between dialogue segments
  maxGapStart := 0.0
  maxGapEnd := 0.0
  maxGapDuration := 0.0
  
  for i := 0; i < len(sceneSegments)-1; i++ {
    gapStart := sceneSegments[i].End
    gapEnd := sceneSegments[i+1].Start
    duration := gapEnd - gapStart
    if duration > maxGapDuration {
      maxGapStart = gapStart
      maxGapEnd = gapEnd
      maxGapDuration = duration
    }
  }
  
  scene.DialoguePresent = true
  scene.DialogueGapStartFrame = int(maxGapStart * fps)
  scene.DialogueGapEndFrame = int(maxGapEnd * fps)
  
  return nil
}
```

---

## 3. Ad Slot Ranking Service

### Purpose
Score candidate insertion moments and rank top 3-5.

### Algorithm

**Scoring Formula:**
```
slot_score = (
  stability_score * 0.4 +           // High stability wins
  (1 - motion_score) * 0.3 +        // Low motion wins
  dialogue_gap_confidence * 0.2 +   // Dialogue gaps preferred
  context_relevance * 0.1            // Product context match
)
```

**Implementation:**
```go
func (sr *SlotRanker) RankSlots(scenes []*Scene, product *Product) ([]*Slot, error) {
  slots := []*Slot{}
  
  for _, scene := range scenes {
    score := sr.calculateSlotScore(scene, product)
    
    slot := &Slot{
      ID:               generateUUID(),
      SceneID:          scene.ID,
      SceneNumber:      scene.SceneNumber,
      InsertionFrame:   scene.DialogueGapStartFrame,
      SlotType:         "dialogue_gap",
      Confidence:       score,
      Score:            score,
      Reasoning:        sr.generateReasoning(scene, product, score),
    }
    
    slots = append(slots, slot)
  }
  
  // Sort by score descending
  sort.Slice(slots, func(i, j int) bool {
    return slots[i].Score > slots[j].Score
  })
  
  // Rank and take top 5
  for i, slot := range slots[:min(5, len(slots))] {
    slot.Rank = i + 1
  }
  
  return slots[:min(5, len(slots))], nil
}

func (sr *SlotRanker) generateReasoning(scene *Scene, product *Product, score float64) string {
  reasons := []string{}
  
  if scene.StabilityScore > 0.7 {
    reasons = append(reasons, "high stability")
  }
  if scene.MotionScore < 0.3 {
    reasons = append(reasons, "low motion")
  }
  if scene.DialoguePresent && scene.DialogueGapEndFrame-scene.DialogueGapStartFrame > 150 {
    reasons = append(reasons, "sufficient dialogue gap (>6 sec)")
  }
  
  // Check if product matches scene context
  if matchesContext(scene.Description, product.ContextKeywords) {
    reasons = append(reasons, fmt.Sprintf("context matches %s", product.Name))
  }
  
  return strings.Join(reasons, ", ")
}
```

---

## 4. Ad Generation (RIFE Frame Interpolation)

### Purpose
Generate smooth 5-8 second ad segment by interpolating between start and end frames.

### Tool: RIFE (Real-Time Intermediate Flow Estimation)

**Provider:** Replicate API (free tier suitable for MVP)

**Replicate Model:** `deforum-research/rife:latest`

### Implementation

```go
import "github.com/replicate/replicate-go"

func (rc *ReplicateClient) InterpolateFrames(
  jobID string,
  videoPath string,
  startFrame, endFrame int,
  multiplier int,  // How many frames to generate (e.g., 4x = multiply frames)
) (string, error) {
  
  // Extract start and end frames as PNG
  startFramePath := rc.extractFrame(videoPath, startFrame)
  endFramePath := rc.extractFrame(videoPath, endFrame)
  
  // Read as base64
  startBase64 := readFileAsBase64(startFramePath)
  endBase64 := readFileAsBase64(endFramePath)
  
  // Call Replicate API
  client := replicate.NewClient(replicate.WithToken(os.Getenv("REPLICATE_API_KEY")))
  
  prediction, err := client.CreatePrediction(ctx, "deforum-research/rife:latest", replicate.PredictionInput{
    "start_frame": startBase64,
    "end_frame":   endBase64,
    "multiplier":  multiplier,  // 4 or 8
  }, nil)
  
  if err != nil {
    return "", err
  }
  
  // Poll for completion (with timeout)
  for i := 0; i < 120; i++ {  // 10 minutes max
    time.Sleep(5 * time.Second)
    prediction, _ := client.GetPrediction(ctx, prediction.ID)
    
    if prediction.Status == "succeeded" {
      // Extract video from output
      outputVideo := prediction.Output.(map[string]interface{})["video"].(string)
      return downloadVideo(outputVideo, jobID)
    } else if prediction.Status == "failed" {
      return "", fmt.Errorf("RIFE failed: %v", prediction.Error)
    }
  }
  
  return "", fmt.Errorf("RIFE timeout after 10 minutes")
}

func (rc *ReplicateClient) extractFrame(videoPath string, frameNum int) string {
  outputPath := fmt.Sprintf("/tmp/frames/frame_%d.png", frameNum)
  
  // ffmpeg command
  cmd := exec.Command("ffmpeg",
    "-i", videoPath,
    "-vf", fmt.Sprintf("select=eq(n\\,%d)", frameNum),
    "-vsync", "vfr",
    outputPath,
  )
  cmd.Run()
  
  return outputPath
}
```

**Error Handling:**
- Replicate API timeout: Retry or fallback to simple frame copy
- Invalid frames: Return error, slot generation fails gracefully

---

## 5. Video Stitching (ffmpeg)

### Purpose
Insert generated ad segment seamlessly into original video.

### Tool: ffmpeg (subprocess)

```go
func (vs *VideoStitcher) StitchVideo(
  originalPath string,
  adPath string,
  insertionFrame int,
  outputPath string,
) error {
  fps := 24.0
  insertionSec := float64(insertionFrame) / fps
  
  // ffmpeg complex filter:
  // 1. Split original: [0:v]trim=0:INSERTION_TIME[before] + trim=INSERTION_TIME:...[after]
  // 2. Concat: [before][ad][after]concat=n=3[out]
  
  filterComplex := fmt.Sprintf(
    "[0:v]trim=0:%f[before];[0:v]trim=%f[after];[before][1:v][after]concat=n=3[out]",
    insertionSec, insertionSec,
  )
  
  cmd := exec.Command("ffmpeg",
    "-i", originalPath,
    "-i", adPath,
    "-filter_complex", filterComplex,
    "-map", "[out]",
    "-map", "0:a",
    "-c:v", "libx264",
    "-preset", "fast",
    "-crf", "18",
    "-c:a", "aac",
    "-y",  // Overwrite
    outputPath,
  )
  
  err := cmd.Run()
  if err != nil {
    return fmt.Errorf("ffmpeg failed: %w", err)
  }
  
  return nil
}
```

**Fallback:** If smooth interpolation fails, use simple hard-cut with 0.5-second cross-fade.

---

## 6. Job Orchestration (Worker Loop)

### Purpose
Async background processing of jobs.

```go
func (jp *JobProcessor) Start() {
  for {
    // Poll for queued jobs every 5 seconds
    job, err := jp.db.GetNextQueuedJob(context.Background())
    if err != nil || job == nil {
      time.Sleep(5 * time.Second)
      continue
    }
    
    jp.processJob(job)
  }
}

func (jp *JobProcessor) processJob(job *Job) {
  jobID := job.ID
  
  // Stage 1: Scene Detection
  jp.updateJobStatus(jobID, "analyzing", "scene_detection")
  campaign, _ := jp.db.GetCampaign(context.Background(), job.CampaignID)
  scenes, err := jp.sceneDetector.DetectScenes(campaign.VideoPath)
  if err != nil {
    jp.failJob(jobID, "Scene detection failed: "+err.Error())
    return
  }
  for _, scene := range scenes {
    jp.db.InsertScene(context.Background(), scene)
  }
  
  // Stage 2: Context Analysis
  jp.updateJobStatus(jobID, "analyzing", "context_analysis")
  for i := range scenes {
    jp.contextAnalyzer.AnalyzeMotion(campaign.VideoPath, &scenes[i])
    transcript, _ := jp.contextAnalyzer.ExtractTranscript(campaign.VideoPath)
    jp.contextAnalyzer.FindDialogueGaps(&scenes[i], transcript)
    jp.db.UpdateScene(context.Background(), &scenes[i])
  }
  
  // Stage 3: Slot Ranking
  jp.updateJobStatus(jobID, "analyzing", "slot_ranking")
  product, _ := jp.db.GetProduct(context.Background(), campaign.ProductID)
  slots, err := jp.slotRanker.RankSlots(scenes, product)
  if err != nil {
    jp.failJob(jobID, "Slot ranking failed")
    return
  }
  for _, slot := range slots {
    slot.JobID = jobID
    jp.db.InsertSlot(context.Background(), slot)
  }
  
  // Update job: now waiting for user to select a slot
  jp.updateJobStatus(jobID, "analyzing", "complete")
  jp.db.UpdateJobMetadata(jobID, map[string]interface{}{
    "top_3_slots": slots[:min(3, len(slots))],
  })
}

func (jp *JobProcessor) updateJobStatus(jobID, status, stage string) {
  jp.db.UpdateJobStatus(context.Background(), jobID, status)
  jp.db.UpdateJobStage(context.Background(), jobID, stage)
}

func (jp *JobProcessor) failJob(jobID, errorMsg string) {
  jp.db.UpdateJobStatus(context.Background(), jobID, "failed")
  jp.db.UpdateJobError(context.Background(), jobID, errorMsg)
}
```

---

## 7. Configuration & Deployment Notes

### Environment Variables
```bash
REPLICATE_API_KEY=<key>
UPLOAD_DIR=/tmp/uploads
OUTPUT_DIR=/tmp/outputs
DATABASE_PATH=ghw_cloud26.db
LOG_LEVEL=info
```

### System Dependencies
```
ffmpeg (with libx264)
OpenCV 4.5+ (with video support)
Go 1.19+
SQLite3
Whisper (OpenAI CLI tool)
```

### Performance Targets
- Scene detection: <30 seconds for 20-min video
- Context analysis: <2 minutes
- Slot ranking: <5 seconds
- Ad generation (RIFE): <10 minutes (bottleneck, Replicate-limited)
- Stitching: <3 minutes

### Testing
Use test video: Public domain film clip (10-20 min, MP4)
- Test all services independently
- Test async job loop
- Load test with multiple campaigns (not required for MVP)



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
