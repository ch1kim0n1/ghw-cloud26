import { expect, test } from "@playwright/test";
import { mockHealthOk, mockPreviewNotFound } from "./helpers";

const JOB_ID = "job_e2e_preview";

const SAMPLE_PREVIEW_PENDING = {
  id: "prev_e2e_1",
  job_id: JOB_ID,
  slot_id: "slot_e2e_1",
  status: "pending",
  duration_seconds: null,
  render_retry_count: 0,
  error_message: null,
  created_at: "2026-03-14T08:00:00Z",
};

const SAMPLE_PREVIEW_COMPLETED = {
  id: "prev_e2e_1",
  job_id: JOB_ID,
  slot_id: "slot_e2e_1",
  status: "completed",
  output_video_path: "/tmp/previews/output.mp4",
  duration_seconds: 12.5,
  render_retry_count: 0,
  error_message: null,
  created_at: "2026-03-14T08:00:00Z",
};

const SAMPLE_PREVIEW_FAILED = {
  id: "prev_e2e_1",
  job_id: JOB_ID,
  slot_id: "slot_e2e_1",
  status: "failed",
  duration_seconds: null,
  render_retry_count: 2,
  error_message: "render process exited with code 1",
  created_at: "2026-03-14T08:00:00Z",
};

test.describe("Preview page", () => {
  test.beforeEach(async ({ page }) => {
    await mockHealthOk(page);
  });

  test("shows preview heading with job id", async ({ page }) => {
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}/preview`);
    await expect(page.getByRole("heading", { name: `Preview status ${JOB_ID}` })).toBeVisible();
  });

  test("shows 'No preview has been started yet' when no preview exists", async ({ page }) => {
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}/preview`);
    await expect(page.getByText("No preview has been started yet.")).toBeVisible();
  });

  test("shows pending status summary when preview is queued", async ({ page }) => {
    await page.route(`**/api/jobs/${JOB_ID}/preview`, (route) => {
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(SAMPLE_PREVIEW_PENDING),
      });
    });
    await page.goto(`/jobs/${JOB_ID}/preview`);
    await expect(page.getByText("Preview render is queued.")).toBeVisible();
    await expect(page.getByText("pending")).toBeVisible();
    await expect(page.getByText(SAMPLE_PREVIEW_PENDING.slot_id)).toBeVisible();
  });

  test("shows completion summary and metadata when preview completes", async ({ page }) => {
    await page.route(`**/api/jobs/${JOB_ID}/preview`, (route) => {
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(SAMPLE_PREVIEW_COMPLETED),
      });
    });
    await page.goto(`/jobs/${JOB_ID}/preview`);
    await expect(page.getByText("Preview render is complete.")).toBeVisible();
    await expect(page.getByText("completed")).toBeVisible();
    await expect(page.getByText(`${SAMPLE_PREVIEW_COMPLETED.duration_seconds.toFixed(1)}s`)).toBeVisible();
    await expect(page.getByText(SAMPLE_PREVIEW_COMPLETED.output_video_path)).toBeVisible();
  });

  test("shows download link when preview is completed", async ({ page }) => {
    await page.route(`**/api/jobs/${JOB_ID}/preview`, (route) => {
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(SAMPLE_PREVIEW_COMPLETED),
      });
    });
    await page.goto(`/jobs/${JOB_ID}/preview`);
    await expect(page.getByRole("link", { name: "Download preview" })).toBeVisible();
  });

  test("shows error message and retry count when preview fails", async ({ page }) => {
    await page.route(`**/api/jobs/${JOB_ID}/preview`, (route) => {
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(SAMPLE_PREVIEW_FAILED),
      });
    });
    await page.goto(`/jobs/${JOB_ID}/preview`);
    await expect(page.getByText("Preview render failed.")).toBeVisible();
    await expect(page.getByText(SAMPLE_PREVIEW_FAILED.error_message)).toBeVisible();
    await expect(page.getByText("2", { exact: true })).toBeVisible();
  });

  test("shows 'Back to job' link", async ({ page }) => {
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}/preview`);
    await expect(page.getByRole("link", { name: "Back to job" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Back to job" })).toHaveAttribute("href", `/jobs/${JOB_ID}`);
  });

  test("clicking 'Back to job' navigates to the job page", async ({ page }) => {
    await mockPreviewNotFound(page, JOB_ID);
    // Mock job page requests to avoid errors
    await page.route(`**/api/jobs/${JOB_ID}`, (route) => {
      void route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ error: "not found", error_code: "RESOURCE_NOT_FOUND" }),
      });
    });
    await page.route(`**/api/jobs/${JOB_ID}/logs`, (route) => {
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ job_id: JOB_ID, logs: [] }),
      });
    });
    await page.route(`**/api/jobs/${JOB_ID}/slots`, (route) => {
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ job_id: JOB_ID, slots: [] }),
      });
    });
    await page.goto(`/jobs/${JOB_ID}/preview`);
    await page.getByRole("link", { name: "Back to job" }).click();
    await expect(page).toHaveURL(new RegExp(`/jobs/${JOB_ID}$`));
  });

  test("shows retry count in the metadata table", async ({ page }) => {
    await page.route(`**/api/jobs/${JOB_ID}/preview`, (route) => {
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ ...SAMPLE_PREVIEW_PENDING, render_retry_count: 3 }),
      });
    });
    await page.goto(`/jobs/${JOB_ID}/preview`);
    await expect(page.getByText("3", { exact: true })).toBeVisible();
  });
});
