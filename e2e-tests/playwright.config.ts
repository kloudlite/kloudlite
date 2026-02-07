import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  expect: {
    timeout: 5_000,
  },
  use: {
    baseURL: 'http://localhost:3001',
    navigationTimeout: 15_000,
    actionTimeout: 10_000,
    trace: 'on-first-retry',
  },
  projects: [
    // Warm up — pre-compiles Turbopack chunks so tests don't timeout
    {
      name: 'warmup',
      testMatch: /global\.setup\.ts/,
      timeout: 30_000,
    },
    // Main tests
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
      dependencies: ['warmup'],
      testIgnore: /global\.setup\.ts/,
    },
  ],
})
