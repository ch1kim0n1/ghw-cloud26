import { expect, test } from "@playwright/test";
import { SAMPLE_PRODUCT, mockHealthOk, mockProductsList } from "./helpers";

const SAMPLE_CAMPAIGN = {
  campaign_id: "camp_e2e_1",
  job_id: "job_e2e_1",
  product_id: SAMPLE_PRODUCT.id,
  name: "Kitchen TV spot",
  status: "queued",
  current_stage: "ready_for_analysis",
  video_filename: "kitchen.mp4",
  video_path: "/tmp/kitchen.mp4",
  source_fps: 29.97,
  duration_seconds: 45.0,
  target_ad_duration_seconds: 6,
  created_at: "2026-03-14T08:00:00Z",
};

test.describe("Campaign intake page", () => {
  test.beforeEach(async ({ page }) => {
    await mockHealthOk(page);
  });

  test("shows campaign intake heading", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/campaigns/new");
    await expect(page.getByRole("heading", { name: "Campaign Intake" })).toBeVisible();
  });

  test("shows inline product mode when no existing products are available", async ({ page }) => {
    await mockProductsList(page, []);
    await page.goto("/campaigns/new");
    await expect(page.getByLabel("Product name")).toBeVisible();
    await expect(page.getByLabel("Source URL")).toBeVisible();
  });

  test("shows existing product selector when products are available", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/campaigns/new");
    await expect(page.getByText("Use existing product")).toBeVisible();
    await expect(page.getByRole("combobox")).toHaveValue(SAMPLE_PRODUCT.id);
  });

  test("can switch from existing product to inline product mode", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/campaigns/new");
    await page.getByText("Create inline product").click();
    await expect(page.getByLabel("Product name")).toBeVisible();
  });

  test("can switch from inline product back to existing product", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/campaigns/new");
    await page.getByText("Create inline product").click();
    await page.getByText("Use existing product").click();
    await expect(page.getByRole("combobox")).toHaveValue(SAMPLE_PRODUCT.id);
  });

  test("shows campaign name field", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/campaigns/new");
    await expect(page.getByLabel("Campaign name")).toBeVisible();
  });

  test("shows video file upload field", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/campaigns/new");
    await expect(page.getByLabel("Source video (H.264 MP4)")).toBeVisible();
  });

  test("shows target ad duration field", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/campaigns/new");
    await expect(page.getByLabel("Target ad duration (seconds)")).toBeVisible();
  });

  test("shows submit button", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/campaigns/new");
    await expect(page.getByRole("button", { name: "Create campaign" })).toBeVisible();
  });

  test("shows validation error when video file is missing", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/campaigns/new");

    await page.getByLabel("Campaign name").fill("Test campaign");
    await page.getByRole("button", { name: "Create campaign" }).click();

    await expect(page.getByText("A source video is required.")).toBeVisible();
  });

  test("shows success state after campaign creation", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.route("**/api/campaigns", (route) => {
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(SAMPLE_CAMPAIGN),
      });
    });

    await page.goto("/campaigns/new");
    await page.getByLabel("Campaign name").fill(SAMPLE_CAMPAIGN.name);

    // Attach a fake video file
    const fileInput = page.locator('input[type="file"][accept*="mp4"]');
    await fileInput.setInputFiles({
      name: "kitchen.mp4",
      mimeType: "video/mp4",
      buffer: Buffer.from("fake mp4 content"),
    });

    await page.getByRole("button", { name: "Create campaign" }).click();

    await expect(page.getByText("Campaign created")).toBeVisible();
    await expect(page.getByRole("heading", { name: SAMPLE_CAMPAIGN.name })).toBeVisible();
    await expect(page.getByText(SAMPLE_CAMPAIGN.job_id)).toBeVisible();
  });

  test("shows error when campaign creation fails", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.route("**/api/campaigns", (route) => {
      void route.fulfill({
        status: 400,
        contentType: "application/json",
        body: JSON.stringify({ error: "video codec not supported", error_code: "INVALID_INPUT" }),
      });
    });

    await page.goto("/campaigns/new");
    await page.getByLabel("Campaign name").fill("Bad campaign");

    const fileInput = page.locator('input[type="file"][accept*="mp4"]');
    await fileInput.setInputFiles({
      name: "bad.mp4",
      mimeType: "video/mp4",
      buffer: Buffer.from("fake content"),
    });

    await page.getByRole("button", { name: "Create campaign" }).click();
    await expect(page.getByText("video codec not supported")).toBeVisible();
  });

  test("shows error when products list fails to load", async ({ page }) => {
    await page.route("**/api/products", (route) => {
      void route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({ error: "service unavailable", error_code: "SERVICE_UNAVAILABLE" }),
      });
    });
    await page.goto("/campaigns/new");
    await expect(page.getByText("service unavailable")).toBeVisible();
  });

  test("shows info panel with phase 1 contract text", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/campaigns/new");
    await expect(page.getByText("Campaigns stop before analysis")).toBeVisible();
  });
});
