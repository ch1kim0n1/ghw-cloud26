import { chromium } from "playwright";

const baseUrl = process.env.SMOKE_BASE_URL ?? "http://127.0.0.1:4173";

const browser = await chromium.launch({ headless: true });

try {
  const page = await browser.newPage();

  await page.route("**/api/jobs?limit=25", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        jobs: [
          {
            id: "job_smoke_1",
            campaign_id: "camp_smoke_1",
            status: "queued",
            current_stage: "ready_for_analysis",
            progress_percent: 0,
            selected_slot_id: null,
            error_code: null,
            created_at: "2026-03-13T00:00:00Z",
            started_at: null,
            completed_at: null,
          },
        ],
      }),
    });
  });

  await page.goto(`${baseUrl}/upload`, { waitUntil: "networkidle" });
  await page.waitForSelector("text=Create video run");

  await page.goto(`${baseUrl}/studio`, { waitUntil: "networkidle" });
  await page.waitForSelector("text=Recent CAFAI jobs");
  await page.waitForSelector("text=job_smoke_1");

  console.log("video flow smoke passed");
} finally {
  await browser.close();
}
