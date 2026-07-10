import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  timeout: 30_000,
  // Single worker: tests share one dev-server/backend instance, not isolated per-worker.
  workers: 1,
  use: {
    baseURL: "http://localhost:3000",
  },
  reporter: [["list"]],
});
