import { test, expect } from "@playwright/test";

test.describe("Search playground", () => {
  test("runs a hybrid search and shows RRF + BM25 + similarity scores", async ({ page }) => {
    await page.goto("/playground/search");

    // Collection select populates from the API — wait for a real option.
    await expect(page.getByRole("combobox").first()).toBeVisible();

    await page.getByPlaceholder("What is machine learning?").fill("Platner");
    const searchButton = page.getByRole("button", { name: "Search", exact: true });
    await expect(searchButton).toBeEnabled({ timeout: 10_000 });
    await searchButton.click();

    await expect(page.getByText(/result/i).first()).toBeVisible({ timeout: 15_000 });

    // At least one chunk card should show the hybrid score breakdown.
    await expect(page.getByText(/RRF \d/).first()).toBeVisible();
    await expect(page.getByText(/BM25 \d/).first()).toBeVisible();
    await expect(page.getByText("vector similarity").first()).toBeVisible();
  });

  test("vector search only hides keyword/RRF scores", async ({ page }) => {
    await page.goto("/playground/search");

    await page.getByPlaceholder("What is machine learning?").fill("Platner");
    await page.getByLabel("Vector search only").check();
    const searchButton = page.getByRole("button", { name: "Search", exact: true });
    await expect(searchButton).toBeEnabled({ timeout: 10_000 });
    await searchButton.click();

    await expect(page.getByText(/result/i).first()).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText("vector similarity").first()).toBeVisible();
    await expect(page.getByText(/RRF \d/)).toHaveCount(0);
    await expect(page.getByText(/BM25 \d/)).toHaveCount(0);
  });

  test("hybrid settings panel expands to show weight overrides", async ({ page }) => {
    await page.goto("/playground/search");

    await page.getByRole("button", { name: "Hybrid search settings" }).click();
    await expect(page.getByLabel("Full-text weight")).toBeVisible();
    await expect(page.getByLabel("Semantic weight")).toBeVisible();
    await expect(page.getByLabel("RRF k")).toBeVisible();
    await expect(page.getByLabel("Full-text limit")).toBeVisible();
  });

  test("document filters panel expands and inserts a builtin field", async ({ page }) => {
    await page.goto("/playground/search");

    await page.getByRole("button", { name: "Document Filters" }).click();
    await expect(page.getByText("Filterable fields — click to insert")).toBeVisible();

    await page.getByRole("button", { name: /mime_type/ }).click();
    const textarea = page.getByPlaceholder(/author.*Alice/);
    await expect(textarea).toHaveValue(/mime_type/);
  });

  test("invalid filter JSON shows an error and blocks submit", async ({ page }) => {
    await page.goto("/playground/search");

    await page.getByPlaceholder("What is machine learning?").fill("test");
    await page.getByRole("button", { name: "Document Filters" }).click();
    await page.getByPlaceholder(/author.*Alice/).fill("{not valid json");
    const searchButton = page.getByRole("button", { name: "Search", exact: true });
    await expect(searchButton).toBeEnabled({ timeout: 10_000 });
    await searchButton.click();

    await expect(page.getByText("Invalid JSON")).toBeVisible();
  });
});
