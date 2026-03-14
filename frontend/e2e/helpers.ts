import type { Page, Route } from "@playwright/test";

export interface MockProduct {
  id: string;
  name: string;
  description?: string;
  category?: string;
  context_keywords?: string[];
  source_url?: string;
  created_at: string;
}

export interface MockJob {
  id: string;
  campaign_id: string;
  status: string;
  current_stage?: string;
  progress_percent: number;
  selected_slot_id: string | null;
  error_message: string | null;
  error_code: string | null;
  created_at: string;
  started_at: string | null;
  completed_at: string | null;
  metadata?: Record<string, unknown>;
}

export interface MockSlot {
  id: string;
  job_id: string;
  rank: number;
  scene_id: string;
  anchor_start_frame: number;
  anchor_end_frame: number;
  source_fps: number;
  quiet_window_seconds: number;
  score: number;
  reasoning: string;
  status: string;
  suggested_product_line: string | null;
  final_product_line: string | null;
}

export const HEALTH_RESPONSE = {
  status: "healthy",
  timestamp: "2026-03-14T08:00:00Z",
  version: "0.1.0-mvp",
  provider_profile: "azure",
};

export const SAMPLE_PRODUCT: MockProduct = {
  id: "prod_e2e_1",
  name: "Sparkling Water",
  description: "Light citrus refreshment",
  category: "beverage",
  context_keywords: ["drink", "water", "refreshment"],
  source_url: "https://example.com/sparkling-water",
  created_at: "2026-03-14T08:00:00Z",
};

export const SAMPLE_JOB: MockJob = {
  id: "job_e2e_1",
  campaign_id: "camp_e2e_1",
  status: "queued",
  current_stage: "ready_for_analysis",
  progress_percent: 0,
  selected_slot_id: null,
  error_message: null,
  error_code: null,
  created_at: "2026-03-14T08:00:00Z",
  started_at: null,
  completed_at: null,
};

export const SAMPLE_SLOT: MockSlot = {
  id: "slot_e2e_1",
  job_id: "job_e2e_1",
  rank: 1,
  scene_id: "scene_e2e_1",
  anchor_start_frame: 100,
  anchor_end_frame: 150,
  source_fps: 29.97,
  quiet_window_seconds: 3.0,
  score: 0.87,
  reasoning: "Good narrative pause with minimal motion",
  status: "pending",
  suggested_product_line: "Stay refreshed with Sparkling Water.",
  final_product_line: null,
};

/** Mock the health API endpoint. */
export async function mockHealthOk(page: Page): Promise<void> {
  await page.route("**/api/health", (route: Route) => {
    void route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(HEALTH_RESPONSE),
    });
  });
}

/** Mock the products list endpoint. */
export async function mockProductsList(page: Page, products: MockProduct[]): Promise<void> {
  await page.route("**/api/products", (route: Route) => {
    if (route.request().method() === "GET") {
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ products }),
      });
    } else {
      void route.continue();
    }
  });
}

/** Mock the products list and product creation endpoints. */
export async function mockProductsWithCreate(
  page: Page,
  existingProducts: MockProduct[],
  newProduct: MockProduct,
): Promise<void> {
  let created = false;
  await page.route("**/api/products", (route: Route) => {
    if (route.request().method() === "GET") {
      const products = created ? [newProduct, ...existingProducts] : existingProducts;
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ products }),
      });
    } else if (route.request().method() === "POST") {
      created = true;
      void route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(newProduct),
      });
    } else {
      void route.continue();
    }
  });
}

/** Mock a job endpoint. */
export async function mockJob(page: Page, job: MockJob): Promise<void> {
  await page.route(`**/api/jobs/${job.id}`, (route: Route) => {
    void route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(job),
    });
  });
}

/** Mock a job logs endpoint. */
export async function mockJobLogs(
  page: Page,
  jobId: string,
  logs: Array<{ timestamp: string; event_type: string; stage_name?: string; message: string }>,
): Promise<void> {
  await page.route(`**/api/jobs/${jobId}/logs`, (route: Route) => {
    void route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ job_id: jobId, logs }),
    });
  });
}

/** Mock a slots list endpoint. */
export async function mockSlots(page: Page, jobId: string, slots: MockSlot[]): Promise<void> {
  await page.route(`**/api/jobs/${jobId}/slots`, (route: Route) => {
    void route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ job_id: jobId, slots }),
    });
  });
}

/** Mock a preview status endpoint returning no preview yet. */
export async function mockPreviewNotFound(page: Page, jobId: string): Promise<void> {
  await page.route(`**/api/jobs/${jobId}/preview`, (route: Route) => {
    void route.fulfill({
      status: 404,
      contentType: "application/json",
      body: JSON.stringify({ error: "preview not found", error_code: "RESOURCE_NOT_FOUND" }),
    });
  });
}
