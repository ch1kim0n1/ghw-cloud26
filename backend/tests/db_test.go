package tests

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
)

func TestEnsureRuntimeDirectories(t *testing.T) {
	root := t.TempDir()
	service := services.NewPathService(testConfig(root))
	if err := service.EnsureRuntimeDirectories(); err != nil {
		t.Fatalf("EnsureRuntimeDirectories() error = %v", err)
	}

	for _, path := range []string{
		filepath.Join(root, "tmp", "uploads", "products"),
		filepath.Join(root, "tmp", "uploads", "campaigns"),
		filepath.Join(root, "tmp", "artifacts"),
		filepath.Join(root, "tmp", "previews"),
		filepath.Join(root, "tmp", "website_ads"),
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", path, err)
		}
		if !info.IsDir() {
			t.Fatalf("%q is not a directory", path)
		}
	}
}

func TestApplyMigrationsCreatesSchema(t *testing.T) {
	ctx := context.Background()
	tempDB := filepath.Join(t.TempDir(), "phase0.db")
	database, err := db.Open(ctx, tempDB)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer database.Close()

	if err := db.ApplyMigrations(ctx, database, filepath.Join("..", "scripts", "migrations")); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	expectedObjects := []string{
		"campaigns",
		"idx_campaigns_product_id",
		"idx_job_logs_job_id_timestamp",
		"idx_job_previews_job_id",
		"idx_jobs_campaign_id",
		"idx_jobs_status",
		"idx_scenes_job_id",
		"idx_slots_job_id",
		"idx_slots_job_rank",
		"idx_slots_status",
		"idx_website_ad_jobs_created_at",
		"job_logs",
		"job_previews",
		"jobs",
		"products",
		"scenes",
		"schema_migrations",
		"slots",
		"website_ad_jobs",
	}

	got := schemaObjects(t, database)
	if strings.Join(got, ",") != strings.Join(expectedObjects, ",") {
		t.Fatalf("unexpected schema objects\nwant: %v\ngot:  %v", expectedObjects, got)
	}
}

func TestInitScriptMatchesMigrationSchema(t *testing.T) {
	ctx := context.Background()
	migratedDB := filepath.Join(t.TempDir(), "migrated.db")
	scriptDB := filepath.Join(t.TempDir(), "script.db")

	withMigrations, err := db.Open(ctx, migratedDB)
	if err != nil {
		t.Fatalf("Open migrated db: %v", err)
	}
	defer withMigrations.Close()

	if err := db.ApplyMigrations(ctx, withMigrations, filepath.Join("..", "scripts", "migrations")); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	withScript, err := db.Open(ctx, scriptDB)
	if err != nil {
		t.Fatalf("Open script db: %v", err)
	}
	defer withScript.Close()

	script, err := os.ReadFile(filepath.Join("..", "scripts", "init_db.sql"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if _, err := withScript.ExecContext(ctx, string(script)); err != nil {
		t.Fatalf("ExecContext() error = %v", err)
	}

	left := schemaDefinitions(t, withMigrations)
	right := schemaDefinitions(t, withScript)
	delete(left, "schema_migrations")

	if len(left) != len(right) {
		t.Fatalf("schema definition counts differ: %d vs %d", len(left), len(right))
	}

	for name, definition := range left {
		if right[name] != definition {
			t.Fatalf("schema mismatch for %s\nwant: %s\ngot:  %s", name, definition, right[name])
		}
	}
}

func schemaObjects(t *testing.T, database *sql.DB) []string {
	t.Helper()

	rows, err := database.Query(`
		SELECT name
		FROM sqlite_master
		WHERE type IN ('table', 'index')
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("rows.Scan() error = %v", err)
		}
		results = append(results, name)
	}
	return results
}

func schemaDefinitions(t *testing.T, database *sql.DB) map[string]string {
	t.Helper()

	rows, err := database.Query(`
		SELECT name, sql
		FROM sqlite_master
		WHERE type IN ('table', 'index')
		  AND name NOT LIKE 'sqlite_%'
		  AND sql IS NOT NULL
	`)
	if err != nil {
		t.Fatalf("query sqlite_master definitions: %v", err)
	}
	defer rows.Close()

	results := map[string]string{}
	for rows.Next() {
		var name string
		var definition string
		if err := rows.Scan(&name, &definition); err != nil {
			t.Fatalf("rows.Scan() error = %v", err)
		}
		results[name] = normalizeSQL(definition)
	}
	return results
}

func normalizeSQL(value string) string {
	normalized := strings.ToLower(strings.Join(strings.Fields(value), " "))
	return strings.ReplaceAll(normalized, "if not exists ", "")
}
