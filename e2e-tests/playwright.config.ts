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
    // Console tests (UI-only, no cloud provisioning)
    {
      name: 'console',
      testDir: './tests/console',
      use: {
        ...devices['Desktop Chrome'],
        baseURL: 'http://localhost:3002',
      },
      dependencies: ['warmup-console'],
      testIgnore: [/global\.setup\.ts/, /\/providers\//],
    },
    // AWS installation lifecycle (long-running)
    {
      name: 'aws',
      testDir: './tests/console/providers/aws',
      timeout: TIMEOUTS.providerTest,
      use: {
        ...devices['Desktop Chrome'],
        baseURL: 'http://localhost:3002',
      },
      dependencies: ['warmup-console'],
    },
    // GCP installation lifecycle (long-running)
    {
      name: 'gcp',
      testDir: './tests/console/providers/gcp',
      timeout: TIMEOUTS.providerTest,
      use: {
        ...devices['Desktop Chrome'],
        baseURL: 'http://localhost:3002',
      },
      dependencies: ['warmup-console'],
    },
    // Azure installation lifecycle (long-running)
    {
      name: 'azure',
      testDir: './tests/console/providers/azure',
      timeout: TIMEOUTS.providerTest,
      use: {
        ...devices['Desktop Chrome'],
        baseURL: 'http://localhost:3002',
      },
      dependencies: ['warmup-console'],
    },
  ],
})
