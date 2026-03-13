package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/api"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/constants"
	backenddb "github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
)

func TestHealthEndpoint(t *testing.T) {
	handler := api.NewRouter(api.Dependencies{
		Config: testConfig(t.TempDir()),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:     &sql.DB{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if payload["status"] != "healthy" {
		t.Fatalf("expected healthy status, got %#v", payload["status"])
	}
}

func TestUnimplementedRouteReturnsStandardEnvelope(t *testing.T) {
	handler := api.NewRouter(api.Dependencies{
		Config: testConfig(t.TempDir()),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:     &sql.DB{},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/jobs/demo-job/slots/demo-slot/select", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}

	var payload api.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if payload.ErrorCode != "NOT_IMPLEMENTED" {
		t.Fatalf("expected NOT_IMPLEMENTED, got %q", payload.ErrorCode)
	}
}

func TestCORSPreflight(t *testing.T) {
	handler := api.NewRouter(api.Dependencies{
		Config: testConfig(t.TempDir()),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:     &sql.DB{},
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/products", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected allow origin header, got %q", got)
	}
}

func TestProductsAPI(t *testing.T) {
	env := newAPIEnv(t)

	t.Run("create product with source url", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name":             "sparkling water",
			"description":      "light citrus",
			"category":         "beverage",
			"context_keywords": "drink, water, refreshment",
			"source_url":       "https://example.com/sparkling-water",
		}, nil)

		rec := env.serve(http.MethodPost, "/api/products", body, contentType)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var product models.Product
		decodeJSON(t, rec.Body.Bytes(), &product)
		if product.Name != "sparkling water" {
			t.Fatalf("unexpected product name %q", product.Name)
		}
		if len(product.ContextKeywords) != 3 {
			t.Fatalf("expected 3 keywords, got %d", len(product.ContextKeywords))
		}
	})

	t.Run("create product with image only", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name": "boxed tea",
		}, map[string]uploadFile{
			"image_file": {
				Filename: "tea.png",
				Content:  pngFixture(t),
			},
		})

		rec := env.serve(http.MethodPost, "/api/products", body, contentType)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var product models.Product
		decodeJSON(t, rec.Body.Bytes(), &product)
		if product.ImagePath == "" {
			t.Fatal("expected image_path to be set")
		}
		if _, err := os.Stat(product.ImagePath); err != nil {
			t.Fatalf("expected saved image at %s: %v", product.ImagePath, err)
		}
	})

	t.Run("reject missing source and image", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name": "invalid product",
		}, nil)

		rec := env.serve(http.MethodPost, "/api/products", body, contentType)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("reject invalid image type", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name": "invalid image",
		}, map[string]uploadFile{
			"image_file": {
				Filename: "image.gif",
				Content:  []byte("gif"),
			},
		})

		rec := env.serve(http.MethodPost, "/api/products", body, contentType)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("list products newest first", func(t *testing.T) {
		bodyA, contentTypeA := multipartRequest(t, map[string]string{
			"name":       "product a",
			"source_url": "https://example.com/a",
		}, nil)
		env.serve(http.MethodPost, "/api/products", bodyA, contentTypeA)

		bodyB, contentTypeB := multipartRequest(t, map[string]string{
			"name":       "product b",
			"source_url": "https://example.com/b",
		}, nil)
		env.serve(http.MethodPost, "/api/products", bodyB, contentTypeB)

		rec := env.serve(http.MethodGet, "/api/products", nil, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response struct {
			Products []models.Product `json:"products"`
		}
		decodeJSON(t, rec.Body.Bytes(), &response)
		if len(response.Products) < 2 {
			t.Fatalf("expected at least 2 products, got %d", len(response.Products))
		}
		indexA := -1
		indexB := -1
		for index, product := range response.Products {
			switch product.Name {
			case "product a":
				indexA = index
			case "product b":
				indexB = index
			}
		}
		if indexA == -1 || indexB == -1 {
			t.Fatalf("expected both seeded products in response: %#v", response.Products)
		}
		if indexB >= indexA {
			t.Fatalf("expected product b before product a, got indexes %d and %d", indexB, indexA)
		}
	})
}

func TestCampaignsAPI(t *testing.T) {
	env := newAPIEnv(t)
	assets := phaseOneVideoAssets(t)

	existingProduct := insertProductFixture(t, env.database, "existing product")

	t.Run("create campaign with existing product", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name":                       "existing product campaign",
			"product_id":                 existingProduct.ID,
			"target_ad_duration_seconds": "6",
		}, map[string]uploadFile{
			"video_file": {
				Filename: "valid.mp4",
				Content:  mustReadFile(t, assets.ValidVideo),
			},
		})

		rec := env.serve(http.MethodPost, "/api/campaigns", body, contentType)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var campaign models.Campaign
		decodeJSON(t, rec.Body.Bytes(), &campaign)
		if campaign.Status != "queued" || campaign.CurrentStage != "ready_for_analysis" {
			t.Fatalf("unexpected campaign job state: %s / %s", campaign.Status, campaign.CurrentStage)
		}
		if campaign.JobID == "" {
			t.Fatal("expected job_id in campaign response")
		}
	})

	t.Run("create campaign with inline product", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name":                     "inline product campaign",
			"product_name":             "inline soda",
			"product_context_keywords": "drink, kitchen",
			"product_source_url":       "https://example.com/inline-soda",
		}, map[string]uploadFile{
			"video_file": {
				Filename: "valid.mp4",
				Content:  mustReadFile(t, assets.ValidVideo),
			},
		})

		rec := env.serve(http.MethodPost, "/api/campaigns", body, contentType)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var campaign models.Campaign
		decodeJSON(t, rec.Body.Bytes(), &campaign)
		if campaign.ProductID == "" {
			t.Fatal("expected inline product to be created")
		}
	})

	t.Run("reject both existing and inline product input", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name":               "invalid campaign",
			"product_id":         existingProduct.ID,
			"product_name":       "inline",
			"product_source_url": "https://example.com/inline",
		}, map[string]uploadFile{
			"video_file": {
				Filename: "valid.mp4",
				Content:  mustReadFile(t, assets.ValidVideo),
			},
		})

		rec := env.serve(http.MethodPost, "/api/campaigns", body, contentType)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("reject missing product input", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name": "missing product campaign",
		}, map[string]uploadFile{
			"video_file": {
				Filename: "valid.mp4",
				Content:  mustReadFile(t, assets.ValidVideo),
			},
		})

		rec := env.serve(http.MethodPost, "/api/campaigns", body, contentType)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("reject unknown product id", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name":       "unknown product campaign",
			"product_id": "prod_missing",
		}, map[string]uploadFile{
			"video_file": {
				Filename: "valid.mp4",
				Content:  mustReadFile(t, assets.ValidVideo),
			},
		})

		rec := env.serve(http.MethodPost, "/api/campaigns", body, contentType)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("reject invalid codec", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name":       "invalid codec campaign",
			"product_id": existingProduct.ID,
		}, map[string]uploadFile{
			"video_file": {
				Filename: "invalid_codec.mp4",
				Content:  mustReadFile(t, assets.InvalidCodecVideo),
			},
		})

		rec := env.serve(http.MethodPost, "/api/campaigns", body, contentType)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("reject invalid duration", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name":       "invalid duration campaign",
			"product_id": existingProduct.ID,
		}, map[string]uploadFile{
			"video_file": {
				Filename: "short.mp4",
				Content:  mustReadFile(t, assets.ShortVideo),
			},
		})

		rec := env.serve(http.MethodPost, "/api/campaigns", body, contentType)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("get campaign by id", func(t *testing.T) {
		body, contentType := multipartRequest(t, map[string]string{
			"name":       "lookup campaign",
			"product_id": existingProduct.ID,
		}, map[string]uploadFile{
			"video_file": {
				Filename: "valid.mp4",
				Content:  mustReadFile(t, assets.ValidVideo),
			},
		})

		createRec := env.serve(http.MethodPost, "/api/campaigns", body, contentType)
		if createRec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", createRec.Code, createRec.Body.String())
		}

		var created models.Campaign
		decodeJSON(t, createRec.Body.Bytes(), &created)

		rec := env.serve(http.MethodGet, "/api/campaigns/"+created.ID, nil, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var fetched models.Campaign
		decodeJSON(t, rec.Body.Bytes(), &fetched)
		if fetched.JobID != created.JobID {
			t.Fatalf("expected job_id %s, got %s", created.JobID, fetched.JobID)
		}
	})
}

func TestPhaseTwoAnalysisWorkflow(t *testing.T) {
	client := &fakeAnalysisClient{
		submitRequestID: "req_123",
		pollResponses: []services.AnalysisPollResponse{
			{RequestID: "req_123", Status: "pending"},
			{RequestID: "req_123", Status: "completed", Scenes: validAnalysisScenes()},
		},
	}
	openAIClient := &fakeOpenAIClient{responseContent: slotRankingContentForSuffix("phase2")}
	env := newAPIEnvWithPhaseTwoClients(t, client, openAIClient)

	product := insertProductFixture(t, env.database, "sparkling water")
	_, job := insertCampaignJobFixture(t, env.database, product.ID, "phase2")

	startRec := env.serve(http.MethodPost, "/api/jobs/"+job.ID+"/start-analysis", nil, "")
	if startRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", startRec.Code, startRec.Body.String())
	}

	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis() error = %v", err)
	}

	jobRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID, nil, "")
	if jobRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", jobRec.Code, jobRec.Body.String())
	}
	var inFlight models.Job
	decodeJSON(t, jobRec.Body.Bytes(), &inFlight)
	if inFlight.CurrentStage != constants.StageAnalysisPoll {
		t.Fatalf("expected analysis_poll after first tick, got %s", inFlight.CurrentStage)
	}
	if _, ok := inFlight.Metadata["analysis_request_id"]; ok {
		t.Fatal("analysis_request_id should not be exposed in job metadata")
	}

	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis() second call error = %v", err)
	}

	finalJobRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID, nil, "")
	if finalJobRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", finalJobRec.Code, finalJobRec.Body.String())
	}
	var analyzed models.Job
	decodeJSON(t, finalJobRec.Body.Bytes(), &analyzed)
	if analyzed.CurrentStage != constants.StageSlotSelection || analyzed.ProgressPercent != 40 {
		t.Fatalf("unexpected analyzed job state: %#v", analyzed)
	}
	if _, ok := analyzed.Metadata["analysis_request_id"]; ok {
		t.Fatal("analysis_request_id should not be exposed after completion")
	}
	if _, ok := analyzed.Metadata["slot_ranking_request_id"]; ok {
		t.Fatal("slot_ranking_request_id should not be exposed after completion")
	}

	slotsRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID+"/slots", nil, "")
	if slotsRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", slotsRec.Code, slotsRec.Body.String())
	}
	var slotsPayload struct {
		JobID string        `json:"job_id"`
		Slots []models.Slot `json:"slots"`
	}
	decodeJSON(t, slotsRec.Body.Bytes(), &slotsPayload)
	if len(slotsPayload.Slots) != 3 {
		t.Fatalf("expected 3 slots, got %d", len(slotsPayload.Slots))
	}
	if slotsPayload.Slots[0].SourceFPS != 24 {
		t.Fatalf("expected source fps to be enriched, got %f", slotsPayload.Slots[0].SourceFPS)
	}

	var sceneCount int
	if err := env.database.QueryRow(`SELECT COUNT(*) FROM scenes WHERE job_id = ?`, job.ID).Scan(&sceneCount); err != nil {
		t.Fatalf("count persisted scenes error = %v", err)
	}
	if sceneCount != 5 {
		t.Fatalf("expected 5 persisted scenes, got %d", sceneCount)
	}

	logsRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID+"/logs", nil, "")
	if logsRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", logsRec.Code, logsRec.Body.String())
	}
	var logsPayload struct {
		Logs []models.JobLog `json:"logs"`
	}
	decodeJSON(t, logsRec.Body.Bytes(), &logsPayload)
	if len(logsPayload.Logs) < 3 {
		t.Fatalf("expected stage logs to be persisted, got %d", len(logsPayload.Logs))
	}
	if openAIClient.calls != 1 {
		t.Fatalf("expected exactly one openai ranking call, got %d", openAIClient.calls)
	}
}

func TestPhaseTwoRejectAndRepick(t *testing.T) {
	client := &fakeAnalysisClient{
		submitRequestID: "req_456",
		pollResponses: []services.AnalysisPollResponse{
			{RequestID: "req_456", Status: "completed", Scenes: validAnalysisScenes()},
		},
	}
	openAIClient := &fakeOpenAIClient{
		responseContent: slotRankingContentForSuffix("repick"),
		responses:       []string{slotRankingContentForSuffix("repick"), repickSlotRankingContent()},
	}
	env := newAPIEnvWithPhaseTwoClients(t, client, openAIClient)
	product := insertProductFixture(t, env.database, "sparkling water repick")
	_, job := insertCampaignJobFixture(t, env.database, product.ID, "repick")

	startRec := env.serve(http.MethodPost, "/api/jobs/"+job.ID+"/start-analysis", nil, "")
	if startRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", startRec.Code, startRec.Body.String())
	}
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis() error = %v", err)
	}

	slotsRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID+"/slots", nil, "")
	var slotsPayload struct {
		Slots []models.Slot `json:"slots"`
	}
	decodeJSON(t, slotsRec.Body.Bytes(), &slotsPayload)
	if len(slotsPayload.Slots) != 3 {
		t.Fatalf("expected 3 initial slots, got %d", len(slotsPayload.Slots))
	}

	rejectRec := env.serve(http.MethodPost, "/api/jobs/"+job.ID+"/slots/"+slotsPayload.Slots[0].ID+"/reject", []byte(`{"note":"too close to dialogue"}`), "application/json")
	if rejectRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rejectRec.Code, rejectRec.Body.String())
	}

	repickTooSoon := env.serve(http.MethodPost, "/api/jobs/"+job.ID+"/slots/re-pick", nil, "")
	if repickTooSoon.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", repickTooSoon.Code, repickTooSoon.Body.String())
	}

	for _, slot := range slotsPayload.Slots[1:] {
		rec := env.serve(http.MethodPost, "/api/jobs/"+job.ID+"/slots/"+slot.ID+"/reject", nil, "application/json")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	}

	repickRec := env.serve(http.MethodPost, "/api/jobs/"+job.ID+"/slots/re-pick", nil, "")
	if repickRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", repickRec.Code, repickRec.Body.String())
	}

	newSlotsRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID+"/slots", nil, "")
	if newSlotsRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", newSlotsRec.Code, newSlotsRec.Body.String())
	}
	var newSlotsPayload struct {
		Slots []models.Slot `json:"slots"`
	}
	decodeJSON(t, newSlotsRec.Body.Bytes(), &newSlotsPayload)
	if len(newSlotsPayload.Slots) != 2 {
		t.Fatalf("expected 2 replacement slots, got %d", len(newSlotsPayload.Slots))
	}
	for _, replacement := range newSlotsPayload.Slots {
		for _, rejected := range slotsPayload.Slots {
			if replacement.ID == rejected.ID {
				t.Fatalf("replacement slot %s reused a rejected candidate", replacement.ID)
			}
		}
	}

	jobRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID, nil, "")
	var repicked models.Job
	decodeJSON(t, jobRec.Body.Bytes(), &repicked)
	if got := repicked.Metadata["repick_count"]; got != float64(1) && got != 1 {
		t.Fatalf("expected repick_count 1, got %#v", got)
	}
}

func TestPhaseTwoHTTPSmokePath(t *testing.T) {
	client := &fakeAnalysisClient{
		submitRequestID: "req_smoke",
		pollResponses: []services.AnalysisPollResponse{
			{RequestID: "req_smoke", Status: "pending"},
			{RequestID: "req_smoke", Status: "completed", Scenes: validAnalysisScenes()},
		},
	}
	openAIClient := &fakeOpenAIClient{
		responses: []string{},
	}
	env := newAPIEnvWithPhaseTwoClients(t, client, openAIClient)
	server := httptest.NewServer(env.handler)
	defer server.Close()

	httpClient := server.Client()
	assets := phaseOneVideoAssets(t)

	doRequest := func(method, path, contentType string, body []byte) *http.Response {
		t.Helper()

		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}

		req, err := http.NewRequest(method, server.URL+path, reader)
		if err != nil {
			t.Fatalf("NewRequest(%s %s) error = %v", method, path, err)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatalf("Do(%s %s) error = %v", method, path, err)
		}
		return resp
	}

	productBody, productContentType := multipartRequest(t, map[string]string{
		"name":       "smoke product",
		"source_url": "https://example.com/smoke-product",
	}, nil)
	productResp := doRequest(http.MethodPost, "/api/products", productContentType, productBody)
	defer productResp.Body.Close()
	if productResp.StatusCode != http.StatusOK {
		t.Fatalf("expected product create 200, got %d", productResp.StatusCode)
	}

	var product models.Product
	productPayload, err := io.ReadAll(productResp.Body)
	if err != nil {
		t.Fatalf("read product response error = %v", err)
	}
	decodeJSON(t, productPayload, &product)

	campaignBody, campaignContentType := multipartRequest(t, map[string]string{
		"name":       "smoke campaign",
		"product_id": product.ID,
	}, map[string]uploadFile{
		"video_file": {
			Filename: "valid.mp4",
			Content:  mustReadFile(t, assets.ValidVideo),
		},
	})
	campaignResp := doRequest(http.MethodPost, "/api/campaigns", campaignContentType, campaignBody)
	defer campaignResp.Body.Close()
	if campaignResp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(campaignResp.Body)
		t.Fatalf("expected campaign create 200, got %d: %s", campaignResp.StatusCode, string(payload))
	}

	var campaign models.Campaign
	campaignPayload, err := io.ReadAll(campaignResp.Body)
	if err != nil {
		t.Fatalf("read campaign response error = %v", err)
	}
	decodeJSON(t, campaignPayload, &campaign)
	openAIClient.responses = []string{
		slotRankingContentForJobID(campaign.JobID),
		repickSlotRankingContentForJobID(campaign.JobID),
	}

	startResp := doRequest(http.MethodPost, "/api/jobs/"+campaign.JobID+"/start-analysis", "", nil)
	defer startResp.Body.Close()
	if startResp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(startResp.Body)
		t.Fatalf("expected start-analysis 200, got %d: %s", startResp.StatusCode, string(payload))
	}

	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis() submission error = %v", err)
	}
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis() completion error = %v", err)
	}

	jobResp := doRequest(http.MethodGet, "/api/jobs/"+campaign.JobID, "", nil)
	defer jobResp.Body.Close()
	if jobResp.StatusCode != http.StatusOK {
		t.Fatalf("expected job status 200, got %d", jobResp.StatusCode)
	}
	var job models.Job
	jobPayload, err := io.ReadAll(jobResp.Body)
	if err != nil {
		t.Fatalf("read job response error = %v", err)
	}
	decodeJSON(t, jobPayload, &job)
	if job.CurrentStage != constants.StageSlotSelection {
		t.Fatalf("expected slot_selection stage, got %s", job.CurrentStage)
	}
	if _, ok := job.Metadata["analysis_request_id"]; ok {
		t.Fatal("analysis_request_id should not be exposed in smoke response")
	}

	slotsResp := doRequest(http.MethodGet, "/api/jobs/"+campaign.JobID+"/slots", "", nil)
	defer slotsResp.Body.Close()
	if slotsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected slots 200, got %d", slotsResp.StatusCode)
	}
	var slotsPayload struct {
		Slots []models.Slot `json:"slots"`
	}
	slotsBodyPayload, err := io.ReadAll(slotsResp.Body)
	if err != nil {
		t.Fatalf("read slots response error = %v", err)
	}
	decodeJSON(t, slotsBodyPayload, &slotsPayload)
	if len(slotsPayload.Slots) != 3 {
		t.Fatalf("expected 3 initial slots, got %d", len(slotsPayload.Slots))
	}

	for _, slot := range slotsPayload.Slots {
		rejectResp := doRequest(http.MethodPost, "/api/jobs/"+campaign.JobID+"/slots/"+slot.ID+"/reject", "application/json", []byte(`{"note":"smoke reject"}`))
		rejectBody, _ := io.ReadAll(rejectResp.Body)
		rejectResp.Body.Close()
		if rejectResp.StatusCode != http.StatusOK {
			t.Fatalf("expected reject 200, got %d: %s", rejectResp.StatusCode, string(rejectBody))
		}
	}

	repickResp := doRequest(http.MethodPost, "/api/jobs/"+campaign.JobID+"/slots/re-pick", "", nil)
	defer repickResp.Body.Close()
	if repickResp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(repickResp.Body)
		t.Fatalf("expected repick 200, got %d: %s", repickResp.StatusCode, string(payload))
	}

	repickSlotsResp := doRequest(http.MethodGet, "/api/jobs/"+campaign.JobID+"/slots", "", nil)
	defer repickSlotsResp.Body.Close()
	if repickSlotsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected replacement slots 200, got %d", repickSlotsResp.StatusCode)
	}
	var repickSlotsPayload struct {
		Slots []models.Slot `json:"slots"`
	}
	repickSlotsBody, err := io.ReadAll(repickSlotsResp.Body)
	if err != nil {
		t.Fatalf("read replacement slots error = %v", err)
	}
	decodeJSON(t, repickSlotsBody, &repickSlotsPayload)
	if len(repickSlotsPayload.Slots) != 2 {
		t.Fatalf("expected 2 replacement slots, got %d", len(repickSlotsPayload.Slots))
	}
	for _, replacement := range repickSlotsPayload.Slots {
		for _, rejected := range slotsPayload.Slots {
			if replacement.ID == rejected.ID {
				t.Fatalf("replacement slot %s reused rejected slot", replacement.ID)
			}
		}
	}
}

func TestPhaseTwoProviderFailureMarksJobFailed(t *testing.T) {
	client := &fakeAnalysisClient{
		submitRequestID: "req_failure",
		pollResponses: []services.AnalysisPollResponse{
			{RequestID: "req_failure", Status: "failed"},
		},
	}
	env := newAPIEnvWithPhaseTwoClients(t, client, &fakeOpenAIClient{responseContent: slotRankingContentForSuffix("failure")})
	product := insertProductFixture(t, env.database, "sparkling water failure")
	_, job := insertCampaignJobFixture(t, env.database, product.ID, "failure")

	startRec := env.serve(http.MethodPost, "/api/jobs/"+job.ID+"/start-analysis", nil, "")
	if startRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", startRec.Code, startRec.Body.String())
	}
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis() error = %v", err)
	}

	jobRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID, nil, "")
	if jobRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", jobRec.Code, jobRec.Body.String())
	}

	var failed models.Job
	decodeJSON(t, jobRec.Body.Bytes(), &failed)
	if failed.Status != constants.JobStatusFailed {
		t.Fatalf("expected failed job, got %s", failed.Status)
	}
	if failed.ErrorCode == nil || *failed.ErrorCode != constants.ErrorCodeAnalysisFailed {
		t.Fatalf("expected analysis failure code, got %#v", failed.ErrorCode)
	}
}

func TestPhaseTwoNoSuitableSlotsMarksJobFailed(t *testing.T) {
	client := &fakeAnalysisClient{
		submitRequestID: "req_none",
		pollResponses: []services.AnalysisPollResponse{
			{RequestID: "req_none", Status: "completed", Scenes: validAnalysisScenes()},
		},
	}
	openAIClient := &fakeOpenAIClient{responseContent: invalidSlotRankingContent()}
	env := newAPIEnvWithPhaseTwoClients(t, client, openAIClient)
	product := insertProductFixture(t, env.database, "sparkling water none")
	_, job := insertCampaignJobFixture(t, env.database, product.ID, "none")

	startRec := env.serve(http.MethodPost, "/api/jobs/"+job.ID+"/start-analysis", nil, "")
	if startRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", startRec.Code, startRec.Body.String())
	}
	if err := env.jobService.ProcessPendingAnalysis(context.Background()); err != nil {
		t.Fatalf("ProcessPendingAnalysis() error = %v", err)
	}

	jobRec := env.serve(http.MethodGet, "/api/jobs/"+job.ID, nil, "")
	if jobRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", jobRec.Code, jobRec.Body.String())
	}
	var failed models.Job
	decodeJSON(t, jobRec.Body.Bytes(), &failed)
	if failed.Status != constants.JobStatusFailed {
		t.Fatalf("expected failed job, got %s", failed.Status)
	}
	if failed.ErrorCode == nil || *failed.ErrorCode != constants.ErrorCodeNoSuitableSlot {
		t.Fatalf("expected no suitable slot code, got %#v", failed.ErrorCode)
	}
}

type apiEnv struct {
	handler    http.Handler
	database   *sql.DB
	jobService *services.JobService
}

func newAPIEnv(t *testing.T) apiEnv {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return newAPIEnvWithPhaseTwoClients(t, services.NewNoopAnalysisClient(logger), services.NewNoopOpenAIClient(logger))
}

func newAPIEnvWithAnalysisClient(t *testing.T, analysisClient services.AnalysisClient) apiEnv {
	t.Helper()
	return newAPIEnvWithPhaseTwoClients(t, analysisClient, services.NewNoopOpenAIClient(slog.New(slog.NewTextHandler(io.Discard, nil))))
}

func newAPIEnvWithPhaseTwoClients(t *testing.T, analysisClient services.AnalysisClient, openAIClient services.OpenAIClient) apiEnv {
	t.Helper()

	root := t.TempDir()
	cfg := testConfig(root)
	if err := os.MkdirAll(filepath.Join(root, "tmp"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	database, err := backenddb.Open(context.Background(), filepath.Join(root, "tmp", "phase1.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := backenddb.ApplyMigrations(context.Background(), database, filepath.Join("..", "scripts", "migrations")); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	handler := api.NewRouter(api.Dependencies{
		Config:         cfg,
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:             database,
		AnalysisClient: analysisClient,
		OpenAIClient:   openAIClient,
	})

	jobService := services.NewJobService(
		database,
		backenddb.NewJobsRepository(database),
		backenddb.NewCampaignsRepository(database),
		backenddb.NewProductsRepository(database),
		backenddb.NewJobLogsRepository(database),
		backenddb.NewScenesRepository(database),
		backenddb.NewSlotsRepository(database),
		analysisClient,
		openAIClient,
	)

	return apiEnv{
		handler:    handler,
		database:   database,
		jobService: jobService,
	}
}

func (e apiEnv) serve(method, path string, body []byte, contentType string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	rec := httptest.NewRecorder()
	e.handler.ServeHTTP(rec, req)
	return rec
}

type uploadFile struct {
	Filename string
	Content  []byte
}

func multipartRequest(t *testing.T, fields map[string]string, files map[string]uploadFile) ([]byte, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s) error = %v", key, err)
		}
	}

	for field, file := range files {
		part, err := writer.CreateFormFile(field, file.Filename)
		if err != nil {
			t.Fatalf("CreateFormFile(%s) error = %v", field, err)
		}
		if _, err := part.Write(file.Content); err != nil {
			t.Fatalf("Write(%s) error = %v", field, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	return body.Bytes(), writer.FormDataContentType()
}

func decodeJSON(t *testing.T, payload []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("Unmarshal() error = %v\npayload: %s", err, string(payload))
	}
}

func insertProductFixture(t *testing.T, database *sql.DB, name string) models.Product {
	t.Helper()

	product := models.Product{
		ID:              "prod_fixture_" + strings.ReplaceAll(strings.ToLower(name), " ", "_"),
		Name:            name,
		ContextKeywords: []string{"drink"},
		SourceURL:       "https://example.com/" + strings.ReplaceAll(strings.ToLower(name), " ", "-"),
		CreatedAt:       "2026-03-13T00:00:00Z",
		UpdatedAt:       "2026-03-13T00:00:00Z",
	}

	if err := backenddb.NewProductsRepository(database).Insert(context.Background(), product); err != nil {
		t.Fatalf("Insert product fixture error = %v", err)
	}
	return product
}

func insertCampaignJobFixture(t *testing.T, database *sql.DB, productID, suffix string) (models.Campaign, models.Job) {
	t.Helper()

	campaign := models.Campaign{
		ID:                      "camp_fixture_" + suffix,
		JobID:                   "job_fixture_" + suffix,
		ProductID:               productID,
		Name:                    "campaign " + suffix,
		Status:                  constants.JobStatusQueued,
		CurrentStage:            constants.StageReadyForAnalysis,
		VideoFilename:           "fixture.mp4",
		VideoPath:               filepath.Join(t.TempDir(), "fixture.mp4"),
		SourceFPS:               24,
		DurationSeconds:         900,
		TargetAdDurationSeconds: 6,
		CreatedAt:               "2026-03-13T00:00:00Z",
		UpdatedAt:               "2026-03-13T00:00:00Z",
	}
	job := models.Job{
		ID:              campaign.JobID,
		CampaignID:      campaign.ID,
		Status:          constants.JobStatusQueued,
		CurrentStage:    constants.StageReadyForAnalysis,
		ProgressPercent: 0,
		CreatedAt:       "2026-03-13T00:00:00Z",
		Metadata: models.Metadata{
			"source_fps":        24.0,
			"duration_seconds":  900.0,
			"rejected_slot_ids": []string{},
			"top_slot_ids":      []string{},
			"repick_count":      0,
		},
	}

	if err := backenddb.NewCampaignsRepository(database).Insert(context.Background(), campaign); err != nil {
		t.Fatalf("Insert campaign fixture error = %v", err)
	}
	if err := backenddb.NewJobsRepository(database).Insert(context.Background(), job); err != nil {
		t.Fatalf("Insert job fixture error = %v", err)
	}

	return campaign, job
}

type videoAssets struct {
	ValidVideo        string
	ShortVideo        string
	InvalidCodecVideo string
}

type fakeAnalysisClient struct {
	submitRequestID string
	submitErr       error
	pollErr         error
	pollResponses   []services.AnalysisPollResponse
	submitCalls     int
	pollCalls       int
}

func (c *fakeAnalysisClient) SubmitAnalysis(context.Context, services.AnalysisRequest) (services.AnalysisResponse, error) {
	c.submitCalls++
	if c.submitErr != nil {
		return services.AnalysisResponse{}, c.submitErr
	}
	return services.AnalysisResponse{RequestID: c.submitRequestID}, nil
}

func (c *fakeAnalysisClient) PollAnalysis(context.Context, services.AnalysisPollRequest) (services.AnalysisPollResponse, error) {
	c.pollCalls++
	if c.pollErr != nil {
		return services.AnalysisPollResponse{}, c.pollErr
	}
	if len(c.pollResponses) == 0 {
		return services.AnalysisPollResponse{RequestID: c.submitRequestID, Status: "pending"}, nil
	}
	response := c.pollResponses[0]
	if len(c.pollResponses) > 1 {
		c.pollResponses = c.pollResponses[1:]
	}
	return response, nil
}

type fakeOpenAIClient struct {
	responseContent string
	responses       []string
	err             error
	calls           int
}

func (c *fakeOpenAIClient) Complete(context.Context, services.OpenAIRequest) (services.OpenAIResponse, error) {
	c.calls++
	if c.err != nil {
		return services.OpenAIResponse{}, c.err
	}
	content := c.responseContent
	if len(c.responses) > 0 {
		content = c.responses[0]
		if len(c.responses) > 1 {
			c.responses = c.responses[1:]
		}
	}
	return services.OpenAIResponse{
		RequestID: fmt.Sprintf("openai_req_%d", c.calls),
		Content:   content,
	}, nil
}

var (
	videoAssetsOnce sync.Once
	videoAssetsData videoAssets
	videoAssetsErr  error
)

func phaseOneVideoAssets(t *testing.T) videoAssets {
	t.Helper()

	videoAssetsOnce.Do(func() {
		root, err := os.MkdirTemp("", "ghw-cloud26-videos-*")
		if err != nil {
			videoAssetsErr = err
			return
		}

		videoAssetsData = videoAssets{
			ValidVideo:        filepath.Join(root, "valid.mp4"),
			ShortVideo:        filepath.Join(root, "short.mp4"),
			InvalidCodecVideo: filepath.Join(root, "invalid_codec.mp4"),
		}

		videoAssetsErr = createVideoFixture(videoAssetsData.ValidVideo, 601, "libx264")
		if videoAssetsErr != nil {
			return
		}
		videoAssetsErr = createVideoFixture(videoAssetsData.ShortVideo, 30, "libx264")
		if videoAssetsErr != nil {
			return
		}
		videoAssetsErr = createVideoFixture(videoAssetsData.InvalidCodecVideo, 601, "mpeg4")
	})

	if videoAssetsErr != nil {
		t.Fatalf("phase one video assets error = %v", videoAssetsErr)
	}
	return videoAssetsData
}

func createVideoFixture(path string, durationSeconds int, codec string) error {
	command := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", "color=c=black:s=16x16:r=1:d="+intString(durationSeconds),
		"-an",
		"-c:v", codec,
		"-pix_fmt", "yuv420p",
		"-preset", "ultrafast",
		path,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w: %s", err, string(output))
	}
	return nil
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return data
}

func pngFixture(t *testing.T) []byte {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9p2qxW8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	return data
}

func validAnalysisScenes() []models.Scene {
	return []models.Scene{
		sceneFixture(1, 0, 480, 0, 20, 0.15, 0.90, 0.15, 4.5, "calm kitchen drink setup", []string{"drink", "kitchen", "refreshment"}, 0.20, 0.10),
		sceneFixture(2, 481, 960, 20, 40, 0.18, 0.85, 0.10, 4.0, "refreshment moment at the counter", []string{"water", "drink", "counter"}, 0.25, 0.12),
		sceneFixture(3, 961, 1440, 40, 60, 0.22, 0.82, 0.18, 3.8, "quiet beverage context", []string{"drink", "beverage"}, 0.22, 0.15),
		sceneFixture(4, 1441, 1920, 60, 80, 0.25, 0.80, 0.20, 3.6, "tabletop refreshment beat", []string{"refreshment", "table", "drink"}, 0.30, 0.18),
		sceneFixture(5, 1921, 2400, 80, 100, 0.30, 0.78, 0.25, 3.3, "soft product context in room", []string{"water", "room"}, 0.28, 0.20),
	}
}

func slotRankingContentForSuffix(suffix string) string {
	return fmt.Sprintf(`{
  "candidates": [
    {
      "scene_id": "scene_job_fixture_%s_001",
      "anchor_start_frame": 220,
      "anchor_end_frame": 221,
      "quiet_window_seconds": 4.5,
      "reasoning": "low motion, strong beverage context, stable continuity anchors",
      "context_relevance_score": 0.93,
      "narrative_fit_score": 0.88,
      "anchor_continuity_score": 0.91,
      "quiet_window_score": 1,
      "motion_score": 0.15,
      "start_boundary_motion_score": 0.18,
      "end_boundary_motion_score": 0.20,
      "action_intensity_score": 0.20,
      "max_subwindow_action_intensity": 0.24,
      "shot_boundary_distance_start_seconds": 9.0,
      "shot_boundary_distance_end_seconds": 10.8,
      "start_cut_confidence": 0.10,
      "end_cut_confidence": 0.12,
      "stability_score": 0.90,
      "dialogue_activity_score": 0.15,
      "metadata": {}
    },
    {
      "scene_id": "scene_job_fixture_%s_002",
      "anchor_start_frame": 720,
      "anchor_end_frame": 721,
      "quiet_window_seconds": 4.0,
      "reasoning": "quiet counter beat with clear drink context",
      "context_relevance_score": 0.89,
      "narrative_fit_score": 0.84,
      "anchor_continuity_score": 0.88,
      "quiet_window_score": 1,
      "motion_score": 0.18,
      "start_boundary_motion_score": 0.21,
      "end_boundary_motion_score": 0.22,
      "action_intensity_score": 0.25,
      "max_subwindow_action_intensity": 0.30,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.12,
      "end_cut_confidence": 0.14,
      "stability_score": 0.85,
      "dialogue_activity_score": 0.10,
      "metadata": {}
    },
    {
      "scene_id": "scene_job_fixture_%s_003",
      "anchor_start_frame": 1200,
      "anchor_end_frame": 1201,
      "quiet_window_seconds": 3.8,
      "reasoning": "quiet beverage context with low disruption risk",
      "context_relevance_score": 0.86,
      "narrative_fit_score": 0.80,
      "anchor_continuity_score": 0.84,
      "quiet_window_score": 1,
      "motion_score": 0.22,
      "start_boundary_motion_score": 0.24,
      "end_boundary_motion_score": 0.26,
      "action_intensity_score": 0.22,
      "max_subwindow_action_intensity": 0.28,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.15,
      "end_cut_confidence": 0.16,
      "stability_score": 0.82,
      "dialogue_activity_score": 0.18,
      "metadata": {}
    },
    {
      "scene_id": "scene_job_fixture_%s_004",
      "anchor_start_frame": 1680,
      "anchor_end_frame": 1681,
      "quiet_window_seconds": 3.6,
      "reasoning": "secondary tabletop beat for repick",
      "context_relevance_score": 0.80,
      "narrative_fit_score": 0.76,
      "anchor_continuity_score": 0.82,
      "quiet_window_score": 1,
      "motion_score": 0.25,
      "start_boundary_motion_score": 0.28,
      "end_boundary_motion_score": 0.30,
      "action_intensity_score": 0.30,
      "max_subwindow_action_intensity": 0.34,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.18,
      "end_cut_confidence": 0.18,
      "stability_score": 0.80,
      "dialogue_activity_score": 0.20,
      "metadata": {}
    },
    {
      "scene_id": "scene_job_fixture_%s_005",
      "anchor_start_frame": 2160,
      "anchor_end_frame": 2161,
      "quiet_window_seconds": 3.3,
      "reasoning": "fallback room beat for repick",
      "context_relevance_score": 0.74,
      "narrative_fit_score": 0.72,
      "anchor_continuity_score": 0.78,
      "quiet_window_score": 1,
      "motion_score": 0.30,
      "start_boundary_motion_score": 0.33,
      "end_boundary_motion_score": 0.34,
      "action_intensity_score": 0.28,
      "max_subwindow_action_intensity": 0.31,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.20,
      "end_cut_confidence": 0.22,
      "stability_score": 0.78,
      "dialogue_activity_score": 0.25,
      "metadata": {}
    }
  ]
}`, suffix, suffix, suffix, suffix, suffix)
}

func slotRankingContentForJobID(jobID string) string {
	return fmt.Sprintf(`{
  "candidates": [
    {
      "scene_id": "scene_%s_001",
      "anchor_start_frame": 220,
      "anchor_end_frame": 221,
      "quiet_window_seconds": 4.5,
      "reasoning": "low motion, strong beverage context, stable continuity anchors",
      "context_relevance_score": 0.93,
      "narrative_fit_score": 0.88,
      "anchor_continuity_score": 0.91,
      "quiet_window_score": 1,
      "motion_score": 0.15,
      "start_boundary_motion_score": 0.18,
      "end_boundary_motion_score": 0.20,
      "action_intensity_score": 0.20,
      "max_subwindow_action_intensity": 0.24,
      "shot_boundary_distance_start_seconds": 9.0,
      "shot_boundary_distance_end_seconds": 10.8,
      "start_cut_confidence": 0.10,
      "end_cut_confidence": 0.12,
      "stability_score": 0.90,
      "dialogue_activity_score": 0.15,
      "metadata": {}
    },
    {
      "scene_id": "scene_%s_002",
      "anchor_start_frame": 720,
      "anchor_end_frame": 721,
      "quiet_window_seconds": 4.0,
      "reasoning": "quiet counter beat with clear drink context",
      "context_relevance_score": 0.89,
      "narrative_fit_score": 0.84,
      "anchor_continuity_score": 0.88,
      "quiet_window_score": 1,
      "motion_score": 0.18,
      "start_boundary_motion_score": 0.21,
      "end_boundary_motion_score": 0.22,
      "action_intensity_score": 0.25,
      "max_subwindow_action_intensity": 0.30,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.12,
      "end_cut_confidence": 0.14,
      "stability_score": 0.85,
      "dialogue_activity_score": 0.10,
      "metadata": {}
    },
    {
      "scene_id": "scene_%s_003",
      "anchor_start_frame": 1200,
      "anchor_end_frame": 1201,
      "quiet_window_seconds": 3.8,
      "reasoning": "quiet beverage context with low disruption risk",
      "context_relevance_score": 0.86,
      "narrative_fit_score": 0.80,
      "anchor_continuity_score": 0.84,
      "quiet_window_score": 1,
      "motion_score": 0.22,
      "start_boundary_motion_score": 0.24,
      "end_boundary_motion_score": 0.26,
      "action_intensity_score": 0.22,
      "max_subwindow_action_intensity": 0.28,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.15,
      "end_cut_confidence": 0.16,
      "stability_score": 0.82,
      "dialogue_activity_score": 0.18,
      "metadata": {}
    },
    {
      "scene_id": "scene_%s_004",
      "anchor_start_frame": 1680,
      "anchor_end_frame": 1681,
      "quiet_window_seconds": 3.6,
      "reasoning": "secondary tabletop beat for repick",
      "context_relevance_score": 0.80,
      "narrative_fit_score": 0.76,
      "anchor_continuity_score": 0.82,
      "quiet_window_score": 1,
      "motion_score": 0.25,
      "start_boundary_motion_score": 0.28,
      "end_boundary_motion_score": 0.30,
      "action_intensity_score": 0.30,
      "max_subwindow_action_intensity": 0.34,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.18,
      "end_cut_confidence": 0.18,
      "stability_score": 0.80,
      "dialogue_activity_score": 0.20,
      "metadata": {}
    },
    {
      "scene_id": "scene_%s_005",
      "anchor_start_frame": 2160,
      "anchor_end_frame": 2161,
      "quiet_window_seconds": 3.3,
      "reasoning": "fallback room beat for repick",
      "context_relevance_score": 0.74,
      "narrative_fit_score": 0.72,
      "anchor_continuity_score": 0.78,
      "quiet_window_score": 1,
      "motion_score": 0.30,
      "start_boundary_motion_score": 0.33,
      "end_boundary_motion_score": 0.34,
      "action_intensity_score": 0.28,
      "max_subwindow_action_intensity": 0.31,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.20,
      "end_cut_confidence": 0.22,
      "stability_score": 0.78,
      "dialogue_activity_score": 0.25,
      "metadata": {}
    }
  ]
}`, jobID, jobID, jobID, jobID, jobID)
}

func repickSlotRankingContent() string {
	return `{
  "candidates": [
    {
      "scene_id": "scene_job_fixture_repick_004",
      "anchor_start_frame": 1680,
      "anchor_end_frame": 1681,
      "quiet_window_seconds": 3.6,
      "reasoning": "secondary tabletop beat for repick",
      "context_relevance_score": 0.80,
      "narrative_fit_score": 0.76,
      "anchor_continuity_score": 0.82,
      "quiet_window_score": 1,
      "motion_score": 0.25,
      "start_boundary_motion_score": 0.28,
      "end_boundary_motion_score": 0.30,
      "action_intensity_score": 0.30,
      "max_subwindow_action_intensity": 0.34,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.18,
      "end_cut_confidence": 0.18,
      "stability_score": 0.80,
      "dialogue_activity_score": 0.20,
      "metadata": {}
    },
    {
      "scene_id": "scene_job_fixture_repick_005",
      "anchor_start_frame": 2160,
      "anchor_end_frame": 2161,
      "quiet_window_seconds": 3.3,
      "reasoning": "fallback room beat for repick",
      "context_relevance_score": 0.74,
      "narrative_fit_score": 0.72,
      "anchor_continuity_score": 0.78,
      "quiet_window_score": 1,
      "motion_score": 0.30,
      "start_boundary_motion_score": 0.33,
      "end_boundary_motion_score": 0.34,
      "action_intensity_score": 0.28,
      "max_subwindow_action_intensity": 0.31,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.20,
      "end_cut_confidence": 0.22,
      "stability_score": 0.78,
      "dialogue_activity_score": 0.25,
      "metadata": {}
    }
  ]
}`
}

func repickSlotRankingContentForJobID(jobID string) string {
	return fmt.Sprintf(`{
  "candidates": [
    {
      "scene_id": "scene_%s_004",
      "anchor_start_frame": 1680,
      "anchor_end_frame": 1681,
      "quiet_window_seconds": 3.6,
      "reasoning": "secondary tabletop beat for repick",
      "context_relevance_score": 0.80,
      "narrative_fit_score": 0.76,
      "anchor_continuity_score": 0.82,
      "quiet_window_score": 1,
      "motion_score": 0.25,
      "start_boundary_motion_score": 0.28,
      "end_boundary_motion_score": 0.30,
      "action_intensity_score": 0.30,
      "max_subwindow_action_intensity": 0.34,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.18,
      "end_cut_confidence": 0.18,
      "stability_score": 0.80,
      "dialogue_activity_score": 0.20,
      "metadata": {}
    },
    {
      "scene_id": "scene_%s_005",
      "anchor_start_frame": 2160,
      "anchor_end_frame": 2161,
      "quiet_window_seconds": 3.3,
      "reasoning": "fallback room beat for repick",
      "context_relevance_score": 0.74,
      "narrative_fit_score": 0.72,
      "anchor_continuity_score": 0.78,
      "quiet_window_score": 1,
      "motion_score": 0.30,
      "start_boundary_motion_score": 0.33,
      "end_boundary_motion_score": 0.34,
      "action_intensity_score": 0.28,
      "max_subwindow_action_intensity": 0.31,
      "shot_boundary_distance_start_seconds": 10.0,
      "shot_boundary_distance_end_seconds": 10.0,
      "start_cut_confidence": 0.20,
      "end_cut_confidence": 0.22,
      "stability_score": 0.78,
      "dialogue_activity_score": 0.25,
      "metadata": {}
    }
  ]
}`, jobID, jobID)
}

func invalidSlotRankingContent() string {
	return `{
  "candidates": [
    {
      "scene_id": "scene_job_fixture_none_001",
      "anchor_start_frame": 220,
      "anchor_end_frame": 221,
      "quiet_window_seconds": 2.0,
      "reasoning": "too short quiet window",
      "context_relevance_score": 0.90,
      "narrative_fit_score": 0.80,
      "anchor_continuity_score": 0.80,
      "quiet_window_score": 0.66,
      "motion_score": 0.70,
      "start_boundary_motion_score": 0.80,
      "end_boundary_motion_score": 0.82,
      "action_intensity_score": 0.75,
      "max_subwindow_action_intensity": 0.85,
      "shot_boundary_distance_start_seconds": 0.2,
      "shot_boundary_distance_end_seconds": 0.2,
      "start_cut_confidence": 0.80,
      "end_cut_confidence": 0.82,
      "stability_score": 0.30,
      "dialogue_activity_score": 0.40,
      "metadata": {}
    }
  ]
}`
}

func sceneFixture(number, startFrame, endFrame int, startSeconds, endSeconds float64, motion, stability, dialogue, quiet float64, summary string, keywords []string, action, abrupt float64) models.Scene {
	return models.Scene{
		SceneNumber:               number,
		StartFrame:                startFrame,
		EndFrame:                  endFrame,
		StartSeconds:              startSeconds,
		EndSeconds:                endSeconds,
		MotionScore:               float64Pointer(motion),
		StabilityScore:            float64Pointer(stability),
		DialogueActivityScore:     float64Pointer(dialogue),
		LongestQuietWindowSeconds: float64Pointer(quiet),
		NarrativeSummary:          summary,
		ContextKeywords:           keywords,
		ActionIntensityScore:      float64Pointer(action),
		AbruptCutRisk:             float64Pointer(abrupt),
	}
}

func float64Pointer(value float64) *float64 {
	return &value
}

func intString(value int) string {
	return fmt.Sprintf("%d", value)
}

func testConfig(root string) config.Config {
	return config.Config{
		RepoRoot:           root,
		ServerAddr:         ":8080",
		DatabasePath:       filepath.Join(root, "tmp", "cafai_mvp.db"),
		MigrationsDir:      filepath.Join("..", "scripts", "migrations"),
		UploadProductsDir:  filepath.Join(root, "tmp", "uploads", "products"),
		UploadCampaignsDir: filepath.Join(root, "tmp", "uploads", "campaigns"),
		ArtifactsDir:       filepath.Join(root, "tmp", "artifacts"),
		PreviewsDir:        filepath.Join(root, "tmp", "previews"),
		AllowedOrigins:     []string{"http://localhost:5173"},
		WorkerInterval:     5,
		ShutdownTimeout:    10,
		Version:            "0.1.0-mvp",
	}
}
