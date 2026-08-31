import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: ".",
  testMatch: "webrtc.spec.ts",
  timeout: 90_000,
  use: {
    browserName: "chromium",
    headless: true,
    permissions: ["microphone"],
    launchOptions: {
      args: ["--use-fake-device-for-media-stream", "--use-fake-ui-for-media-stream"],
    },
  },
  webServer: {
    command: "npx vite --host 127.0.0.1",
    port: 4173,
    reuseExistingServer: false,
  },
});
