import { expect, test } from "@playwright/test";
import {
  SAMPLE_JOB,
  SAMPLE_SLOT,
  mockHealthOk,
  mockJob,
  mockJobLogs,
  mockPreviewNotFound,
  mockSlots,
} from "./helpers";

const JOB_ID = SAMPLE_JOB.id;

test.describe("Job dashboard", () => {
  test.beforeEach(async ({ page }) => {
    await mockHealthOk(page);
  });

  test("shows the job dashboard heading with job id", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText(`Job dashboard ${JOB_ID}`)).toBeVisible();
  });

  test("shows job status in the status strip", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText(SAMPLE_JOB.status, { exact: true }).first()).toBeVisible();
  });

  test("shows job metadata in the status card", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText(SAMPLE_JOB.current_stage!)).toBeVisible();
    await expect(page.getByText("0%")).toBeVisible();
  });

  test("shows 'Start analysis' button when job is queued", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    const startBtn = page.getByRole("button", { name: "Start analysis" });
    await expect(startBtn).toBeVisible();
    await expect(startBtn).not.toBeDisabled();
  });

  test("shows 'Start analysis' button disabled for non-queued job", async ({ page }) => {
    const analyzingJob = { ...SAMPLE_JOB, status: "analyzing", current_stage: "analysis_poll" };
    await mockJob(page, analyzingJob);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByRole("button", { name: "Start analysis" })).toBeDisabled();
  });

  test("shows success message after starting analysis", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.route(`**/api/jobs/${JOB_ID}/start-analysis`, (route) => {
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ message: "analysis started", job_id: JOB_ID }),
      });
    });
    await page.goto(`/jobs/${JOB_ID}`);
    await page.getByRole("button", { name: "Start analysis" }).click();
    await expect(page.getByText("analysis started")).toBeVisible();
  });

  test("shows error message when starting analysis fails", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.route(`**/api/jobs/${JOB_ID}/start-analysis`, (route) => {
      void route.fulfill({
        status: 409,
        contentType: "application/json",
        body: JSON.stringify({ error: "analysis already started", error_code: "CONFLICT" }),
      });
    });
    await page.goto(`/jobs/${JOB_ID}`);
    await page.getByRole("button", { name: "Start analysis" }).click();
    await expect(page.getByText("analysis already started")).toBeVisible();
  });

  test("shows 'No slots available yet' when slots list is empty", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText("No slots available yet.")).toBeVisible();
  });

  test("renders slot cards when slots are available", async ({ page }) => {
    const proposedSlot = { ...SAMPLE_SLOT, status: "proposed" };
    const analyzingJob = { ...SAMPLE_JOB, status: "analyzing", current_stage: "slot_selection" };
    await mockJob(page, analyzingJob);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, [proposedSlot]);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText(proposedSlot.reasoning)).toBeVisible();
    await expect(page.getByText(`Rank ${proposedSlot.rank}`)).toBeVisible();
    await expect(page.getByText(proposedSlot.score.toFixed(3))).toBeVisible();
  });

  test("shows suggested product line in slot card", async ({ page }) => {
    const proposedSlot = { ...SAMPLE_SLOT, status: "proposed" };
    const analyzingJob = { ...SAMPLE_JOB, status: "analyzing", current_stage: "slot_selection" };
    await mockJob(page, analyzingJob);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, [proposedSlot]);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText(proposedSlot.suggested_product_line!)).toBeVisible();
  });

  test("select and reject buttons are enabled for proposed slots during slot_selection stage", async ({ page }) => {
    const proposedSlot = { ...SAMPLE_SLOT, status: "proposed" };
    const analyzingJob = { ...SAMPLE_JOB, status: "analyzing", current_stage: "slot_selection" };
    await mockJob(page, analyzingJob);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, [proposedSlot]);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByRole("button", { name: "Select" })).not.toBeDisabled();
    await expect(page.getByRole("button", { name: "Reject" })).not.toBeDisabled();
  });

  test("shows 'No job logs yet' when no logs exist", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText("No job logs yet.")).toBeVisible();
  });

  test("renders log entries when logs are available", async ({ page }) => {
    const logs = [
      {
        timestamp: "2026-03-14T08:00:00Z",
        event_type: "job_created",
        stage_name: "intake",
        message: "Job was created successfully",
      },
      {
        timestamp: "2026-03-14T08:01:00Z",
        event_type: "analysis_started",
        stage_name: "analysis_submission",
        message: "Analysis submission queued",
      },
    ];
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, logs);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText("job_created")).toBeVisible();
    await expect(page.getByText("Job was created successfully")).toBeVisible();
    await expect(page.getByText("analysis_started")).toBeVisible();
    await expect(page.getByText("Analysis submission queued")).toBeVisible();
  });

  test("shows slot count in the status strip", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, [SAMPLE_SLOT]);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText("1 slot(s)")).toBeVisible();
  });

  test("shows log count in the status strip", async ({ page }) => {
    const logs = [
      { timestamp: "2026-03-14T08:00:00Z", event_type: "job_created", message: "created" },
    ];
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, logs);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText("1 log entry(ies)")).toBeVisible();
  });

  test("re-pick button is disabled when not all slots are rejected", async ({ page }) => {
    const proposedSlot = { ...SAMPLE_SLOT, status: "proposed" };
    const analyzingJob = { ...SAMPLE_JOB, status: "analyzing", current_stage: "slot_selection" };
    await mockJob(page, analyzingJob);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, [proposedSlot]);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByRole("button", { name: "Re-pick slots" })).toBeDisabled();
  });

  test("re-pick button is enabled when all slots are rejected", async ({ page }) => {
    const rejectedSlot = { ...SAMPLE_SLOT, status: "rejected" };
    const analyzingJob = { ...SAMPLE_JOB, status: "analyzing", current_stage: "slot_selection" };
    await mockJob(page, analyzingJob);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, [rejectedSlot]);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByRole("button", { name: "Re-pick slots" })).not.toBeDisabled();
  });

  test("shows preview render section", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByText("Preview render", { exact: true })).toBeVisible();
    await expect(page.getByRole("button", { name: "Render preview" })).toBeVisible();
  });

  test("preview render button is disabled without a generated slot", async ({ page }) => {
    await mockJob(page, SAMPLE_JOB);
    await mockJobLogs(page, JOB_ID, []);
    await mockSlots(page, JOB_ID, []);
    await mockPreviewNotFound(page, JOB_ID);
    await page.goto(`/jobs/${JOB_ID}`);
    await expect(page.getByRole("button", { name: "Render preview" })).toBeDisabled();
  });
});
