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
	backenddb "github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/models"
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

func TestPlaceholderRouteReturnsStandardEnvelope(t *testing.T) {
	handler := api.NewRouter(api.Dependencies{
		Config: testConfig(t.TempDir()),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:     &sql.DB{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/jobs/demo-job", nil)
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
		if response.Products[0].Name != "product b" {
			t.Fatalf("expected newest product first, got %q", response.Products[0].Name)
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

type apiEnv struct {
	handler  http.Handler
	database *sql.DB
}

func newAPIEnv(t *testing.T) apiEnv {
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
		Config: cfg,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:     database,
	})

	return apiEnv{
		handler:  handler,
		database: database,
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

type videoAssets struct {
	ValidVideo        string
	ShortVideo        string
	InvalidCodecVideo string
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
