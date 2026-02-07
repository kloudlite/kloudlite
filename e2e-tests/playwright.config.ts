import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  expect: {
    timeout: 5_000,
  },
  use: {
    navigationTimeout: 15_000,
    actionTimeout: 10_000,
    trace: 'on-first-retry',
  },
  projects: [
    // Dashboard warm up — pre-compiles Turbopack chunks
    {
      name: 'warmup-dashboard',
      testDir: './tests/dashboard',
      testMatch: /global\.setup\.ts/,
      timeout: 30_000,
      use: { baseURL: 'http://localhost:3001' },
    },
    // Dashboard tests
    {
      name: 'dashboard',
      testDir: './tests/dashboard',
      use: {
        ...devices['Desktop Chrome'],
        baseURL: 'http://localhost:3001',
      },
      dependencies: ['warmup-dashboard'],
      testIgnore: /global\.setup\.ts/,
    },
    // Console warm up — pre-compiles Turbopack chunks
    {
      name: 'warmup-console',
      testDir: './tests/console',
      testMatch: /global\.setup\.ts/,
      timeout: 30_000,
      use: { baseURL: 'http://localhost:3002' },
    },
    // Console tests
    {
      name: 'console',
      testDir: './tests/console',
      use: {
        ...devices['Desktop Chrome'],
        baseURL: 'http://localhost:3002',
      },
      dependencies: ['warmup-console'],
      testIgnore: /global\.setup\.ts/,
    },
  ],
})
