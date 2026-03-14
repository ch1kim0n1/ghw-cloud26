import { expect, test } from "@playwright/test";
import { HEALTH_RESPONSE, mockHealthOk, mockPreviewNotFound } from "./helpers";

test.describe("Navigation", () => {
  test.beforeEach(async ({ page }) => {
    await mockHealthOk(page);
  });

  test("root redirects to /products", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL(/\/products/);
    await expect(page.getByRole("heading", { name: "Product Catalog" })).toBeVisible();
  });

  test("nav bar links are visible on every page", async ({ page }) => {
    await page.goto("/products");
    await expect(page.getByRole("link", { name: "Products" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Create Campaign" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Job" })).toBeVisible();
  });

  test("navigating to Create Campaign via nav bar", async ({ page }) => {
    await page.goto("/products");
    await page.getByRole("link", { name: "Create Campaign" }).click();
    await expect(page).toHaveURL(/\/campaigns\/new/);
    await expect(page.getByRole("heading", { name: "Campaign Intake" })).toBeVisible();
  });

  test("navigating to job dashboard via nav bar", async ({ page }) => {
    await mockPreviewNotFound(page, "demo-job");
    await page.goto("/products");
    await page.getByRole("link", { name: "Job" }).click();
    await expect(page).toHaveURL(/\/jobs\/demo-job/);
    await expect(page.getByText(/Job dashboard demo-job/)).toBeVisible();
  });

  test("header shows application title", async ({ page }) => {
    await page.goto("/products");
    await expect(page.getByRole("heading", { name: "Cloud-assisted ad insertion dashboard" })).toBeVisible();
  });

  test("header shows CAFAI phase label", async ({ page }) => {
    await page.goto("/products");
    await expect(page.getByText("CAFAI phase 4")).toBeVisible();
  });
});

test.describe("Health status panel", () => {
  test("shows healthy status when backend responds ok", async ({ page }) => {
    await mockHealthOk(page);
    await page.goto("/products");
    await expect(page.getByText("healthy")).toBeVisible();
    await expect(page.getByText(`Version ${HEALTH_RESPONSE.version}`)).toBeVisible();
    await expect(page.getByText(`Provider ${HEALTH_RESPONSE.provider_profile}`)).toBeVisible();
  });

  test("shows loading state before health response arrives", async ({ page }) => {
    // Delay the health response so we can observe the loading state
    await page.route("**/api/health", async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 500));
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(HEALTH_RESPONSE),
      });
    });
    await page.goto("/products");
    await expect(page.getByText("Checking health")).toBeVisible();
    await expect(page.getByText("healthy")).toBeVisible();
  });

  test("shows error state when backend is unavailable", async ({ page }) => {
    await page.route("**/api/health", (route) => {
      void route.fulfill({
        status: 503,
        contentType: "application/json",
        body: JSON.stringify({ error: "backend unavailable", error_code: "SERVICE_UNAVAILABLE" }),
      });
    });
    await page.goto("/products");
    await expect(page.getByText("Connection failed")).toBeVisible();
  });
});
