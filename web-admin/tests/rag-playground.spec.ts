import { test, expect } from "@playwright/test";

test.describe("RAG playground", () => {
  test("top_k input defaults to 2", async ({ page }) => {
    await page.goto("/playground/rag");
    await expect(page.getByLabel("Top K")).toHaveValue("2");
  });

  test("hybrid settings and document filters panels are present", async ({ page }) => {
    await page.goto("/playground/rag");

    await expect(page.getByRole("button", { name: "Hybrid search settings" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Document Filters" })).toBeVisible();
  });

  test("runs RAG and shows an answer with source chunks carrying scores", async ({ page }) => {
    await page.goto("/playground/rag");

    await page.getByPlaceholder("What is machine learning?").fill("Platner");
    await page.getByRole("button", { name: "Run RAG" }).click();

    await expect(page.getByText("Formatted Prompt")).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(/source/i).first()).toBeVisible();
    await expect(page.getByText("vector similarity").first()).toBeVisible();
  });

  test("document filters panel expands and inserts a builtin field", async ({ page }) => {
    await page.goto("/playground/rag");

    await page.getByRole("button", { name: "Document Filters" }).click();
    await page.getByRole("button", { name: /mime_type/ }).click();
    const textarea = page.getByPlaceholder(/author.*Alice/);
    await expect(textarea).toHaveValue(/mime_type/);
  });
});
