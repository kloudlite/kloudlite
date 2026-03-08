import { defineConfig, devices } from '@playwright/test'
import { TIMEOUTS } from './lib/constants'

export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  expect: {
    timeout: 5_000,
  },
  use: {
    navigationTimeout: TIMEOUTS.navigation,
    actionTimeout: TIMEOUTS.action,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
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
    // Console tests (quick UI tests)
    {
      name: 'console',
      testDir: './tests/console',
      use: {
        ...devices['Desktop Chrome'],
        baseURL: 'http://localhost:3002',
      },
      dependencies: ['warmup-console'],
      testIgnore: [/global\.setup\.ts/, /\/billing\//],
    },
    // Console billing flow (requires Stripe, longer timeout)
    {
      name: 'console-billing',
      testDir: './tests/console/billing',
      timeout: 120_000,
      use: {
        ...devices['Desktop Chrome'],
        baseURL: 'http://localhost:3002',
      },
      dependencies: ['warmup-console'],
    },
  ],
})
