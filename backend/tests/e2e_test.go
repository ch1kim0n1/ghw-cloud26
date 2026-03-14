package tests

// TestFullJobLifecycleE2E validates the complete job workflow from product
// creation through preview download, exercising every HTTP API endpoint in
// sequence. This test acts as the primary end-to-end regression guard for the
// four-phase CAFAI pipeline.
//
// Phase 1  – product creation and campaign intake (requires FFmpeg for video validation)
// Phase 2  – analysis submission, polling, slot ranking, reject/re-pick
// Phase 3  – slot selection, product-line generation, CAFAI clip generation
// Phase 4  – preview render submission, polling, download, and stream

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
)

func TestFullJobLifecycleE2E(t *testing.T) {
	requireMediaToolchain(t)

	// --- shared fake clients ----------------------------------------------
	const (
		analysisRequestID = "req_e2e_lifecycle"
		genRequestID      = "gen_e2e_lifecycle"
		renderRequestID   = "render_e2e_lifecycle"
	)

	analysisClient := &fakeAnalysisClient{
		submitRequestID: analysisRequestID,
		pollResponses: []services.AnalysisPollResponse{
			{RequestID: analysisRequestID, Status: "pending"},
			{RequestID: analysisRequestID, Status: "completed", Scenes: validAnalysisScenes()},
		},
	}

	// openAI responses will be set after campaign creation so we can embed the real job ID
	openAIClient := &fakeOpenAIClient{}

	mlClient := &fakeMLClient{
		submitResponse: services.GenerationResponse{RequestID: genRequestID, Status: "submitted"},
		pollResponses: []services.GenerationResponse{
			{
				RequestID: genRequestID,
				Status:    "completed",
				// paths are resolved relative to the test working directory
				GeneratedClipPath:  "tmp/artifacts/job_e2e/slot_e2e.mp4",
				GeneratedAudioPath: "tmp/artifacts/job_e2e/slot_e2e.wav",
				Metadata:           models.Metadata{"duration_seconds": 6.0},
			},
		},
	}

	blobClient := &fakeBlobStorageClient{}

	renderClient := &fakeRenderClient{
		submitResponse: services.RenderResponse{RequestID: renderRequestID, Status: "submitted"},
		pollResponses: []services.RenderResponse{
			{RequestID: renderRequestID, Status: "pending"},
			{
				RequestID:       renderRequestID,
				Status:          "completed",
				PreviewBlobURI:  "https://blob.example.com/renders/e2e_preview.mp4",
				DurationSeconds: 55.0,
			},
		},
	}

	env := newAPIEnvWithPreviewClients(
		t,
		analysisClient,
		openAIClient,
		mlClient,
		&fakeAnchorFrameExtractor{},
		blobClient,
		renderClient,
	)

	// Spin up a real HTTP test server so the smoke path exercises the full
	// middleware stack (CORS, logging, recovery) in addition to business logic.
	server := httptest.NewServer(env.handler)
	t.Cleanup(server.Close)

	do := func(method, path, contentType string, body []byte) *http.Response {
		t.Helper()
		var r io.Reader
		if body != nil {
			r = bytes.NewReader(body)
		}
		req, err := http.NewRequest(method, server.URL+path, r)
		if err != nil {
			t.Fatalf("NewRequest(%s %s): %v", method, path, err)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatalf("Do(%s %s): %v", method, path, err)
		}
		return resp
	}

	readJSON := func(resp *http.Response, target any) {
		t.Helper()
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		if err := json.Unmarshal(body, target); err != nil {
			t.Fatalf("Unmarshal: %v\nbody: %s", err, string(body))
		}
	}

	assets := phaseOneVideoAssets(t)

	// -----------------------------------------------------------------------
	// Phase 1 – product creation
	// -----------------------------------------------------------------------
	t.Log("Phase 1: creating product via HTTP API")

	productBody, productCT := multipartRequest(t, map[string]string{
		"name":       "e2e water",
		"source_url": "https://example.com/e2e-water",
	}, nil)
	productResp := do(http.MethodPost, "/api/products", productCT, productBody)
	if productResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(productResp.Body)
		productResp.Body.Close()
		t.Fatalf("product create: expected 200, got %d: %s", productResp.StatusCode, body)
	}
	var product models.Product
	readJSON(productResp, &product)
	if product.ID == "" {
		t.Fatal("product create: expected non-empty product ID")
	}
	if product.Name != "e2e water" {
		t.Fatalf("product create: expected name %q, got %q", "e2e water", product.Name)
	}

	// Verify the product appears in the catalog
	listResp := do(http.MethodGet, "/api/products", "", nil)
	if listResp.StatusCode != http.StatusOK {
		listResp.Body.Close()
		t.Fatalf("product list: expected 200, got %d", listResp.StatusCode)
	}
	var catalog struct {
		Products []models.Product `json:"products"`
	}
	readJSON(listResp, &catalog)
	found := false
	for _, p := range catalog.Products {
		if p.ID == product.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("product list: new product %s not found in catalog", product.ID)
	}

	// -----------------------------------------------------------------------
	// Phase 1 – campaign creation
	// -----------------------------------------------------------------------
	t.Log("Phase 1: creating campaign via HTTP API")

	campaignBody, campaignCT := multipartRequest(t, map[string]string{
		"name":       "e2e lifecycle campaign",
		"product_id": product.ID,
	}, map[string]uploadFile{
		"video_file": {
			Filename: "valid.mp4",
			Content:  mustReadFile(t, assets.ValidVideo),
		},
	})
	campaignResp := do(http.MethodPost, "/api/campaigns", campaignCT, campaignBody)
	if campaignResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(campaignResp.Body)
		campaignResp.Body.Close()
		t.Fatalf("campaign create: expected 200, got %d: %s", campaignResp.StatusCode, body)
	}
	var campaign models.Campaign
	readJSON(campaignResp, &campaign)
	if campaign.JobID == "" {
		t.Fatal("campaign create: expected non-empty job ID")
	}
	if campaign.Status != constants.JobStatusQueued {
		t.Fatalf("campaign create: expected queued status, got %s", campaign.Status)
	}
	if campaign.CurrentStage != constants.StageReadyForAnalysis {
		t.Fatalf("campaign create: expected ready_for_analysis stage, got %s", campaign.CurrentStage)
	}

	// Verify the campaign is retrievable by ID
	getResp := do(http.MethodGet, "/api/campaigns/"+campaign.ID, "", nil)
	if getResp.StatusCode != http.StatusOK {
		getResp.Body.Close()
		t.Fatalf("campaign get: expected 200, got %d", getResp.StatusCode)
	}
	var fetchedCampaign models.Campaign
	readJSON(getResp, &fetchedCampaign)
	if fetchedCampaign.ID != campaign.ID {
		t.Fatalf("campaign get: ID mismatch %s vs %s", fetchedCampaign.ID, campaign.ID)
	}

	jobID := campaign.JobID

	// Configure OpenAI responses now that we know the real job ID so that the
	// slot ranking JSON embeds the correct slot IDs.
	openAIClient.responses = []string{
		slotRankingContentForJobID(jobID),
		`{"suggested_product_line":"Grab a sparkling e2e water and refresh yourself."}`,
		`{"generation_brief":"Natural pause into e2e water reveal."}`,
	}

	// -----------------------------------------------------------------------
	// Phase 2 – start analysis
	// -----------------------------------------------------------------------
	t.Log("Phase 2: starting analysis")

	startResp := do(http.MethodPost, "/api/jobs/"+jobID+"/start-analysis", "", nil)
	if startResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(startResp.Body)
		startResp.Body.Close()
		t.Fatalf("start-analysis: expected 200, got %d: %s", startResp.StatusCode, body)
	}
	startResp.Body.Close()

	// Verify job is now in analysis_submission stage
	jobResp := do(http.MethodGet, "/api/jobs/"+jobID, "", nil)
	var inFlightJob models.Job
	readJSON(jobResp, &inFlightJob)
	if inFlightJob.Status != constants.JobStatusAnalyzing {
		t.Fatalf("after start-analysis: expected analyzing status, got %s", inFlightJob.Status)
	}

	// Worker iteration 1: submit analysis
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis (submission): %v", err)
	}

	// Worker iteration 2: poll analysis (pending → completed)
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis (poll): %v", err)
	}

	// Verify job has advanced to slot_selection
	jobResp2 := do(http.MethodGet, "/api/jobs/"+jobID, "", nil)
	var analyzedJob models.Job
	readJSON(jobResp2, &analyzedJob)
	if analyzedJob.CurrentStage != constants.StageSlotSelection {
		t.Fatalf("after analysis: expected slot_selection stage, got %s", analyzedJob.CurrentStage)
	}
	if _, exposed := analyzedJob.Metadata["analysis_request_id"]; exposed {
		t.Fatal("analysis_request_id must not be exposed in job metadata")
	}

	// Verify slots are available
	slotsResp := do(http.MethodGet, "/api/jobs/"+jobID+"/slots", "", nil)
	if slotsResp.StatusCode != http.StatusOK {
		slotsResp.Body.Close()
		t.Fatalf("slots list: expected 200, got %d", slotsResp.StatusCode)
	}
	var slotsPayload struct {
		JobID string        `json:"job_id"`
		Slots []models.Slot `json:"slots"`
	}
	readJSON(slotsResp, &slotsPayload)
	if len(slotsPayload.Slots) == 0 {
		t.Fatal("slots list: expected at least one slot after analysis")
	}

	// Verify job logs contain analysis events
	logsResp := do(http.MethodGet, "/api/jobs/"+jobID+"/logs", "", nil)
	if logsResp.StatusCode != http.StatusOK {
		logsResp.Body.Close()
		t.Fatalf("job logs: expected 200, got %d", logsResp.StatusCode)
	}
	var logsPayload struct {
		JobID string          `json:"job_id"`
		Logs  []models.JobLog `json:"logs"`
	}
	readJSON(logsResp, &logsPayload)
	if len(logsPayload.Logs) == 0 {
		t.Fatal("job logs: expected log entries after analysis")
	}

	// -----------------------------------------------------------------------
	// Phase 2 – slot rejection and re-pick
	// -----------------------------------------------------------------------
	t.Log("Phase 2: rejecting all slots and requesting re-pick")

	// Reset openAI for repick
	openAIClient.responses = append(openAIClient.responses,
		repickSlotRankingContentForJobID(jobID),
	)

	for _, slot := range slotsPayload.Slots {
		rejectBody := []byte(`{"note":"e2e rejection"}`)
		rejectResp := do(http.MethodPost, "/api/jobs/"+jobID+"/slots/"+slot.ID+"/reject", "application/json", rejectBody)
		if rejectResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(rejectResp.Body)
			rejectResp.Body.Close()
			t.Fatalf("slot reject %s: expected 200, got %d: %s", slot.ID, rejectResp.StatusCode, body)
		}
		rejectResp.Body.Close()
	}

	repickResp := do(http.MethodPost, "/api/jobs/"+jobID+"/slots/re-pick", "", nil)
	if repickResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(repickResp.Body)
		repickResp.Body.Close()
		t.Fatalf("re-pick: expected 200, got %d: %s", repickResp.StatusCode, body)
	}
	repickResp.Body.Close()

	// Worker: process re-pick (uses remaining OpenAI response)
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis (repick): %v", err)
	}

	// Verify new slots are different from the original rejected ones
	repickSlotsResp := do(http.MethodGet, "/api/jobs/"+jobID+"/slots", "", nil)
	var repickSlotsPayload struct {
		Slots []models.Slot `json:"slots"`
	}
	readJSON(repickSlotsResp, &repickSlotsPayload)
	if len(repickSlotsPayload.Slots) == 0 {
		t.Fatal("re-pick: expected new slots after re-pick")
	}
	rejectedIDs := make(map[string]bool, len(slotsPayload.Slots))
	for _, s := range slotsPayload.Slots {
		rejectedIDs[s.ID] = true
	}
	for _, s := range repickSlotsPayload.Slots {
		if rejectedIDs[s.ID] {
			t.Fatalf("re-pick: repicked slot %s reused a rejected slot ID", s.ID)
		}
	}

	// -----------------------------------------------------------------------
	// Phase 3 – select a slot and generate CAFAI clip
	// -----------------------------------------------------------------------
	t.Log("Phase 3: selecting slot and starting CAFAI generation")

	// Reset OpenAI for phase 3
	openAIClient.responses = []string{
		`{"suggested_product_line":"Grab a sparkling e2e water and refresh yourself."}`,
		`{"generation_brief":"Natural pause into e2e water reveal."}`,
	}

	// Find the first proposed slot in the repicked set
	var selectedSlot models.Slot
	for _, s := range repickSlotsPayload.Slots {
		if s.Status == constants.SlotStatusProposed {
			selectedSlot = s
			break
		}
	}
	if selectedSlot.ID == "" {
		t.Fatal("phase 3: no proposed slot available for selection")
	}

	selectResp := do(http.MethodPost, "/api/jobs/"+jobID+"/slots/"+selectedSlot.ID+"/select", "", nil)
	if selectResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(selectResp.Body)
		selectResp.Body.Close()
		t.Fatalf("slot select: expected 200, got %d: %s", selectResp.StatusCode, body)
	}
	selectResp.Body.Close()

	// Verify slot details are retrievable
	slotDetailResp := do(http.MethodGet, "/api/jobs/"+jobID+"/slots/"+selectedSlot.ID, "", nil)
	if slotDetailResp.StatusCode != http.StatusOK {
		slotDetailResp.Body.Close()
		t.Fatalf("slot detail: expected 200, got %d", slotDetailResp.StatusCode)
	}
	var slotDetail models.Slot
	readJSON(slotDetailResp, &slotDetail)
	if slotDetail.Status != constants.SlotStatusSelected {
		t.Fatalf("slot detail: expected selected status, got %s", slotDetail.Status)
	}

	generateBody := []byte(`{"product_line_mode":"auto"}`)
	generateResp := do(http.MethodPost, "/api/jobs/"+jobID+"/slots/"+selectedSlot.ID+"/generate", "application/json", generateBody)
	if generateResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(generateResp.Body)
		generateResp.Body.Close()
		t.Fatalf("generate: expected 200, got %d: %s", generateResp.StatusCode, body)
	}
	generateResp.Body.Close()

	// Worker: process generation submission
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis (generation submission): %v", err)
	}

	// Create the fake generated artifacts that the blob client upload step expects
	artifactsDir := filepath.Join("tmp", "artifacts", "job_e2e")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", artifactsDir, err)
	}
	if err := os.WriteFile(filepath.Join(artifactsDir, "slot_e2e.mp4"), []byte("fake e2e clip"), 0o644); err != nil {
		t.Fatalf("WriteFile clip: %v", err)
	}
	if err := os.WriteFile(filepath.Join(artifactsDir, "slot_e2e.wav"), []byte("fake e2e audio"), 0o644); err != nil {
		t.Fatalf("WriteFile audio: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll("tmp") })

	// Worker: poll generation (pending → completed)
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis (generation poll): %v", err)
	}

	// Verify slot is now in generated state
	generatedSlotResp := do(http.MethodGet, "/api/jobs/"+jobID+"/slots/"+selectedSlot.ID, "", nil)
	var generatedSlot models.Slot
	readJSON(generatedSlotResp, &generatedSlot)
	if generatedSlot.Status != constants.SlotStatusGenerated {
		errStr := "<nil>"
		if generatedSlot.GenerationError != nil {
			errStr = *generatedSlot.GenerationError
		}
		t.Fatalf("after generation: expected generated slot status, got %s (err: %s)", generatedSlot.Status, errStr)
	}

	// -----------------------------------------------------------------------
	// Phase 4 – preview rendering
	// -----------------------------------------------------------------------
	t.Log("Phase 4: rendering preview")

	renderBody := []byte(`{"slot_id":"` + selectedSlot.ID + `"}`)
	renderStartResp := do(http.MethodPost, "/api/jobs/"+jobID+"/preview/render", "application/json", renderBody)
	if renderStartResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(renderStartResp.Body)
		renderStartResp.Body.Close()
		t.Fatalf("preview render: expected 200, got %d: %s", renderStartResp.StatusCode, body)
	}
	var renderPayload map[string]any
	readJSON(renderStartResp, &renderPayload)
	if renderPayload["current_stage"] != constants.StageRenderSubmit {
		t.Fatalf("preview render: expected %s stage, got %#v", constants.StageRenderSubmit, renderPayload["current_stage"])
	}

	// Worker: submit render to remote provider
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis (render submission): %v", err)
	}

	// Set up the fake blob download so the preview artifact gets persisted locally
	if blobClient.uploads == nil {
		blobClient.uploads = map[string][]byte{}
	}
	blobClient.uploads["https://blob.example.com/renders/e2e_preview.mp4"] = []byte("fake e2e preview mp4 bytes")

	// Worker: poll render (pending → completed)
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis (render poll pending): %v", err)
	}
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis (render poll completed): %v", err)
	}

	// Verify preview is completed
	previewResp := do(http.MethodGet, "/api/jobs/"+jobID+"/preview", "", nil)
	if previewResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(previewResp.Body)
		previewResp.Body.Close()
		t.Fatalf("preview status: expected 200, got %d: %s", previewResp.StatusCode, body)
	}
	var preview models.Preview
	readJSON(previewResp, &preview)
	if preview.Status != "completed" {
		errStr := "<nil>"
		if preview.ErrorMessage != nil {
			errStr = *preview.ErrorMessage
		}
		t.Fatalf("preview: expected completed status, got %s (%v)", preview.Status, errStr)
	}
	if preview.DownloadPath == "" {
		t.Fatal("preview: expected non-empty download path on completed preview")
	}

	// Verify overall job is completed
	finalJobResp := do(http.MethodGet, "/api/jobs/"+jobID, "", nil)
	var finalJob models.Job
	readJSON(finalJobResp, &finalJob)
	if finalJob.Status != constants.JobStatusCompleted {
		t.Fatalf("final job: expected completed status, got %s", finalJob.Status)
	}
	if _, exposed := finalJob.Metadata["render_request_id"]; exposed {
		t.Fatal("render_request_id must not be exposed in final job metadata")
	}

	// Verify preview file download
	downloadResp := do(http.MethodGet, "/api/jobs/"+jobID+"/preview/download", "", nil)
	if downloadResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(downloadResp.Body)
		downloadResp.Body.Close()
		t.Fatalf("download: expected 200, got %d: %s", downloadResp.StatusCode, body)
	}
	downloadResp.Body.Close()
	if ct := downloadResp.Header.Get("Content-Type"); !strings.Contains(ct, "video/mp4") {
		t.Fatalf("download: expected video/mp4 Content-Type, got %q", ct)
	}
	if cd := downloadResp.Header.Get("Content-Disposition"); !strings.Contains(cd, "attachment;") {
		t.Fatalf("download: expected attachment Content-Disposition, got %q", cd)
	}

	// Verify preview stream endpoint
	streamResp := do(http.MethodGet, "/api/jobs/"+jobID+"/preview/stream", "", nil)
	if streamResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(streamResp.Body)
		streamResp.Body.Close()
		t.Fatalf("stream: expected 200, got %d: %s", streamResp.StatusCode, body)
	}
	streamResp.Body.Close()
	if ct := streamResp.Header.Get("Content-Type"); !strings.Contains(ct, "video/mp4") {
		t.Fatalf("stream: expected video/mp4 Content-Type, got %q", ct)
	}
	if cd := streamResp.Header.Get("Content-Disposition"); !strings.Contains(cd, "inline;") {
		t.Fatalf("stream: expected inline Content-Disposition, got %q", cd)
	}

	t.Logf("✓ Full job lifecycle E2E completed for job %s", jobID)
}

// TestJobWorkflowWithoutMediaToolchain validates all API endpoints that do not
// require FFmpeg or FFprobe, providing a fast E2E path for environments where
// the media toolchain is unavailable (e.g., pure Go CI runners).
func TestJobWorkflowWithoutMediaToolchain(t *testing.T) {
	env := newAPIEnv(t)

	// Health check
	healthRec := env.serve(http.MethodGet, "/api/health", nil, "")
	if healthRec.Code != http.StatusOK {
		t.Fatalf("health: expected 200, got %d", healthRec.Code)
	}
	var healthPayload map[string]any
	decodeJSON(t, healthRec.Body.Bytes(), &healthPayload)
	if healthPayload["status"] != "healthy" {
		t.Fatalf("health: expected healthy status, got %#v", healthPayload["status"])
	}

	// Product creation
	productRec := env.serve(http.MethodPost, "/api/products", marshalJSON(t, map[string]string{
		"name":       "no-media product",
		"source_url": "https://example.com/no-media",
	}), "application/json")
	// Products endpoint uses multipart so fall back to multipart
	productBody, productCT := multipartRequest(t, map[string]string{
		"name":       "no-media product",
		"source_url": "https://example.com/no-media",
	}, nil)
	productRec = env.serve(http.MethodPost, "/api/products", productBody, productCT)
	if productRec.Code != http.StatusOK {
		t.Fatalf("product create: expected 200, got %d: %s", productRec.Code, productRec.Body.String())
	}
	var product models.Product
	decodeJSON(t, productRec.Body.Bytes(), &product)
	if product.ID == "" {
		t.Fatal("product create: expected non-empty product ID")
	}

	// Product list
	listRec := env.serve(http.MethodGet, "/api/products", nil, "")
	if listRec.Code != http.StatusOK {
		t.Fatalf("product list: expected 200, got %d", listRec.Code)
	}
	var catalog struct {
		Products []models.Product `json:"products"`
	}
	decodeJSON(t, listRec.Body.Bytes(), &catalog)
	if len(catalog.Products) == 0 {
		t.Fatal("product list: expected at least one product after creation")
	}

	// Insert a campaign+job fixture directly (bypasses FFmpeg validation)
	_, job := insertCampaignJobFixture(t, env.database, product.ID, "no_media")

	// Job status
	jobRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID, nil, "")
	if jobRec.Code != http.StatusOK {
		t.Fatalf("job status: expected 200, got %d", jobRec.Code)
	}
	var fetchedJob models.Job
	decodeJSON(t, jobRec.Body.Bytes(), &fetchedJob)
	if fetchedJob.Status != constants.JobStatusQueued {
		t.Fatalf("job status: expected queued, got %s", fetchedJob.Status)
	}
	if fetchedJob.CurrentStage != constants.StageReadyForAnalysis {
		t.Fatalf("job status: expected ready_for_analysis, got %s", fetchedJob.CurrentStage)
	}

	// Job logs (empty initially)
	logsRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID+"/logs", nil, "")
	if logsRec.Code != http.StatusOK {
		t.Fatalf("job logs: expected 200, got %d", logsRec.Code)
	}
	var logsPayload struct {
		JobID string          `json:"job_id"`
		Logs  []models.JobLog `json:"logs"`
	}
	decodeJSON(t, logsRec.Body.Bytes(), &logsPayload)
	if logsPayload.JobID != job.ID {
		t.Fatalf("job logs: expected job_id %s, got %s", job.ID, logsPayload.JobID)
	}

	// Slots (empty before analysis)
	slotsRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID+"/slots", nil, "")
	if slotsRec.Code != http.StatusOK {
		t.Fatalf("slots list: expected 200, got %d", slotsRec.Code)
	}
	var slotsPayload struct {
		Slots []models.Slot `json:"slots"`
	}
	decodeJSON(t, slotsRec.Body.Bytes(), &slotsPayload)
	if len(slotsPayload.Slots) != 0 {
		t.Fatalf("slots list: expected 0 slots before analysis, got %d", len(slotsPayload.Slots))
	}

	// Preview (not yet started – 404)
	previewRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID+"/preview", nil, "")
	if previewRec.Code != http.StatusNotFound {
		t.Fatalf("preview status: expected 404 before render, got %d", previewRec.Code)
	}

	// CORS preflight on a jobs endpoint
	preflightReq := httptest.NewRequest(http.MethodOptions, "/api/jobs/"+job.ID, nil)
	preflightReq.Header.Set("Origin", "http://localhost:5173")
	preflightReq.Header.Set("Access-Control-Request-Method", http.MethodGet)
	preflightRec := httptest.NewRecorder()
	env.handler.ServeHTTP(preflightRec, preflightReq)
	if preflightRec.Code != http.StatusNoContent {
		t.Fatalf("CORS preflight: expected 204, got %d", preflightRec.Code)
	}
	if got := preflightRec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("CORS preflight: expected origin header, got %q", got)
	}
}

// marshalJSON is a test helper that marshals a value to JSON.
func marshalJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshalJSON: %v", err)
	}
	return data
}
