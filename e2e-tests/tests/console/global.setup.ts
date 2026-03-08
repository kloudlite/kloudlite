import { test } from '@playwright/test'

/**
 * Global setup for console tests.
 * Hits key pages to pre-compile Turbopack chunks before real tests run.
 */
test('warm up console', async ({ page }) => {
  // Hit login page to trigger initial compilation
  await page.goto('/login', { waitUntil: 'domcontentloaded' })

  // Hit dev-login to compile auth routes
  await page.goto('/api/dev-login', { waitUntil: 'domcontentloaded' })
  await page.waitForURL('**/installations', { timeout: 30_000 })

  // Hit the new-kl-cloud page to compile billing form
  await page.goto('/installations/new-kl-cloud', { waitUntil: 'domcontentloaded' })
})
