import { expect, test } from "@playwright/test";
import {
  SAMPLE_PRODUCT,
  mockHealthOk,
  mockProductsList,
  mockProductsWithCreate,
} from "./helpers";

test.describe("Products page", () => {
  test.beforeEach(async ({ page }) => {
    await mockHealthOk(page);
  });

  test("shows empty state when no products exist", async ({ page }) => {
    await mockProductsList(page, []);
    await page.goto("/products");
    await expect(page.getByText("No products yet.")).toBeVisible();
  });

  test("shows product catalog heading", async ({ page }) => {
    await mockProductsList(page, []);
    await page.goto("/products");
    await expect(page.getByRole("heading", { name: "Product Catalog" })).toBeVisible();
  });

  test("renders existing products in the catalog", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/products");
    await expect(page.getByText(SAMPLE_PRODUCT.name)).toBeVisible();
    await expect(page.getByText(SAMPLE_PRODUCT.description!)).toBeVisible();
    await expect(page.getByText(SAMPLE_PRODUCT.category!)).toBeVisible();
  });

  test("renders product source URL", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/products");
    await expect(page.getByText(`Source: ${SAMPLE_PRODUCT.source_url}`)).toBeVisible();
  });

  test("renders product context keywords", async ({ page }) => {
    await mockProductsList(page, [SAMPLE_PRODUCT]);
    await page.goto("/products");
    await expect(page.getByText(/drink, water, refreshment/)).toBeVisible();
  });

  test("renders product without optional fields gracefully", async ({ page }) => {
    const minimal = {
      id: "prod_minimal",
      name: "Minimal product",
      created_at: "2026-03-14T08:00:00Z",
    };
    await mockProductsList(page, [minimal]);
    await page.goto("/products");
    await expect(page.getByText("Minimal product")).toBeVisible();
    await expect(page.getByText("No description provided.")).toBeVisible();
    await expect(page.getByText("Uncategorized")).toBeVisible();
    await expect(page.getByText("Keywords: None")).toBeVisible();
    await expect(page.getByText("Source: Uploaded image only")).toBeVisible();
  });

  test("form contains all required fields", async ({ page }) => {
    await mockProductsList(page, []);
    await page.goto("/products");
    await expect(page.getByLabel("Name")).toBeVisible();
    await expect(page.getByLabel("Description")).toBeVisible();
    await expect(page.getByLabel("Category")).toBeVisible();
    await expect(page.getByLabel("Context keywords")).toBeVisible();
    await expect(page.getByLabel("Source URL")).toBeVisible();
    await expect(page.getByLabel(/Image file/)).toBeVisible();
    await expect(page.getByRole("button", { name: "Create product" })).toBeVisible();
  });

  test("creates a product and shows it in the catalog", async ({ page }) => {
    const newProduct = { ...SAMPLE_PRODUCT, id: "prod_new" };
    await mockProductsWithCreate(page, [], newProduct);
    await page.goto("/products");

    await page.getByLabel("Name").fill(newProduct.name);
    await page.getByLabel("Description").fill(newProduct.description!);
    await page.getByLabel("Source URL").fill(newProduct.source_url!);
    await page.getByRole("button", { name: "Create product" }).click();

    await expect(page.getByText(`Created ${newProduct.name}.`)).toBeVisible();
    await expect(page.getByRole("heading", { name: newProduct.name })).toBeVisible();
  });

  test("shows validation error when name is missing", async ({ page }) => {
    await mockProductsList(page, []);
    await page.goto("/products");

    // HTML5 required validation prevents submission without a name
    const nameInput = page.getByLabel("Name");
    await expect(nameInput).toHaveAttribute("required");
  });

  test("shows error message when product creation fails", async ({ page }) => {
    await mockProductsList(page, []);
    await page.route("**/api/products", async (route) => {
      if (route.request().method() === "POST") {
        await route.fulfill({
          status: 500,
          contentType: "application/json",
          body: JSON.stringify({ error: "internal server error", error_code: "INTERNAL_ERROR" }),
        });
      } else {
        await route.continue();
      }
    });

    await page.goto("/products");
    await page.getByLabel("Name").fill("failing product");
    await page.getByRole("button", { name: "Create product" }).click();
    await expect(page.getByText("internal server error")).toBeVisible();
  });

  test("refresh button reloads the product list", async ({ page }) => {
    let callCount = 0;
    await page.route("**/api/products", (route) => {
      if (route.request().method() === "GET") {
        callCount++;
        void route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({ products: [] }),
        });
      } else {
        void route.continue();
      }
    });

    await page.goto("/products");
    await expect(page.getByText("No products yet.")).toBeVisible();

    await page.getByRole("button", { name: "Refresh" }).click();
    await expect(page.getByText("No products yet.")).toBeVisible();
    expect(callCount).toBeGreaterThanOrEqual(2);
  });

  test("shows multiple products in the catalog", async ({ page }) => {
    const products = [
      { ...SAMPLE_PRODUCT, id: "prod_1", name: "Product A" },
      { ...SAMPLE_PRODUCT, id: "prod_2", name: "Product B" },
    ];
    await mockProductsList(page, products);
    await page.goto("/products");
    await expect(page.getByText("Product A")).toBeVisible();
    await expect(page.getByText("Product B")).toBeVisible();
  });
});
